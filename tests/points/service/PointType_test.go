package srvpoints_test

import (
	"FGG-Service/src/points/service"
	"FGG-Service/src/points/type"
	"FGG-Service/tests/points/mock"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

var availableRollsType = typepoints.PointTypeInfo{
	Id: 5, PartyId: 1, Name: typepoints.PointTypeAvailableRolls, Minimum: 0, Maximum: 10,
}

// --- GetPointValueByTypeName ---

type GetPointValueByTypeNameTestCase struct {
	Name           string
	SetupMock      func() *dbpointsmock.DatabaseMock
	ExpectedValue  int
	ExpectedError  error
	ExpectAnyError bool
}

var GetPointValueByTypeNameTestCases = []GetPointValueByTypeNameTestCase{
	{
		// GetPointTypeByNameCommand fails. The error returns.
		Name: "GetPointTypeByName_DatabaseError",
		SetupMock: func() *dbpointsmock.DatabaseMock {
			databaseMock := new(dbpointsmock.DatabaseMock)
			databaseMock.On("GetPointTypeByNameCommand", 1, typepoints.PointTypeAvailableRolls).Return(typepoints.PointTypeInfo{}, dbError)
			return databaseMock
		},
		ExpectedError: dbError,
	},
	{
		// No PointType matches the requested name. An error returns.
		Name: "NameNotFound_Error",
		SetupMock: func() *dbpointsmock.DatabaseMock {
			databaseMock := new(dbpointsmock.DatabaseMock)
			databaseMock.On("GetPointTypeByNameCommand", 1, typepoints.PointTypeAvailableRolls).Return(typepoints.PointTypeInfo{}, sql.ErrNoRows)
			return databaseMock
		},
		ExpectAnyError: true,
	},
	{
		// GetUserPointCommand returns sql.ErrNoRows. 0 returns with no error.
		Name: "NoUserPointRow_ZeroValue",
		SetupMock: func() *dbpointsmock.DatabaseMock {
			databaseMock := new(dbpointsmock.DatabaseMock)
			databaseMock.On("GetPointTypeByNameCommand", 1, typepoints.PointTypeAvailableRolls).Return(availableRollsType, nil)
			databaseMock.On("GetUserPointCommand", 2, 1, availableRollsType.Id).Return(typepoints.UserPoint{}, sql.ErrNoRows)
			return databaseMock
		},
		ExpectedValue: 0,
	},
	{
		// GetUserPointCommand returns a database error. The error returns.
		Name: "GetUserPoint_DatabaseError",
		SetupMock: func() *dbpointsmock.DatabaseMock {
			databaseMock := new(dbpointsmock.DatabaseMock)
			databaseMock.On("GetPointTypeByNameCommand", 1, typepoints.PointTypeAvailableRolls).Return(availableRollsType, nil)
			databaseMock.On("GetUserPointCommand", 2, 1, availableRollsType.Id).Return(typepoints.UserPoint{}, dbError)
			return databaseMock
		},
		ExpectedError: dbError,
	},
	{
		// GetUserPointCommand succeeds. Its value returns.
		Name: "Success",
		SetupMock: func() *dbpointsmock.DatabaseMock {
			databaseMock := new(dbpointsmock.DatabaseMock)
			databaseMock.On("GetPointTypeByNameCommand", 1, typepoints.PointTypeAvailableRolls).Return(availableRollsType, nil)
			databaseMock.On("GetUserPointCommand", 2, 1, availableRollsType.Id).
				Return(typepoints.UserPoint{Id: 1, UserId: 2, PartyId: 1, PointTypeId: availableRollsType.Id, Value: 7}, nil)
			return databaseMock
		},
		ExpectedValue: 7,
	},
}

func TestSrvPoints_GetPointValueByTypeName(test *testing.T) {
	for _, testCase := range GetPointValueByTypeNameTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			databaseMock := testCase.SetupMock()
			sut := srvpoints.Service{Database: databaseMock}

			// Act
			value, err := sut.GetPointValueByTypeName(2, 1, typepoints.PointTypeAvailableRolls)

			// Assert
			if testCase.ExpectedError != nil {
				require.ErrorIs(test, err, testCase.ExpectedError)
			} else if testCase.ExpectAnyError {
				require.Error(test, err)
			} else {
				require.NoError(test, err)
				require.Equal(test, testCase.ExpectedValue, value)
			}

			databaseMock.AssertExpectations(test)
		})
	}
}

// --- ChangeUserPointValueClamped ---

type ChangeUserPointValueClampedTestCase struct {
	Name          string
	ChangeValue   int
	SetupMock     func() *dbpointsmock.DatabaseMock
	ExpectedError error
}

var ChangeUserPointValueClampedTestCases = []ChangeUserPointValueClampedTestCase{
	{
		// GetPointTypeCommand fails. The error returns and nothing else happens.
		Name:        "GetPointType_DatabaseError",
		ChangeValue: 1,
		SetupMock: func() *dbpointsmock.DatabaseMock {
			databaseMock := new(dbpointsmock.DatabaseMock)
			databaseMock.On("GetPointTypeCommand", 1, availableRollsType.Id).Return(typepoints.PointTypeInfo{}, dbError)
			return databaseMock
		},
		ExpectedError: dbError,
	},
	{
		// GetUserPointCommand fails with something other than sql.ErrNoRows. The error returns.
		Name:        "GetUserPoint_DatabaseError",
		ChangeValue: 1,
		SetupMock: func() *dbpointsmock.DatabaseMock {
			databaseMock := new(dbpointsmock.DatabaseMock)
			databaseMock.On("GetPointTypeCommand", 1, availableRollsType.Id).Return(availableRollsType, nil)
			databaseMock.On("GetUserPointCommand", 2, 1, availableRollsType.Id).Return(typepoints.UserPoint{}, dbError)
			return databaseMock
		},
		ExpectedError: dbError,
	},
	{
		// Increasing the value beyond Maximum clamps it to Maximum, and the clamped delta is what's
		// written and recorded (not the raw requested change).
		Name:        "ClampToMaximum",
		ChangeValue: 5,
		SetupMock: func() *dbpointsmock.DatabaseMock {
			databaseMock := new(dbpointsmock.DatabaseMock)
			databaseMock.On("GetPointTypeCommand", 1, availableRollsType.Id).Return(availableRollsType, nil) // Maximum: 10
			databaseMock.On("GetUserPointCommand", 2, 1, availableRollsType.Id).
				Return(typepoints.UserPoint{Value: 8}, nil)
			// 8 + 5 = 13 > Maximum(10) -> clamp to 10, actual change = 2
			databaseMock.On("ChangeUserPointValueCommand", 2, 1, availableRollsType.Id, 2).Return(nil)
			databaseMock.On("CreateUserPointHistoryCommand", 2, 1, availableRollsType.Id, 9, 5, 2, 10, 999).
				Return(typepoints.UserPointHistoryEntry{}, nil)
			return databaseMock
		},
	},
	{
		// Decreasing the value below Minimum clamps it to Minimum.
		Name:        "ClampToMinimum",
		ChangeValue: -5,
		SetupMock: func() *dbpointsmock.DatabaseMock {
			databaseMock := new(dbpointsmock.DatabaseMock)
			databaseMock.On("GetPointTypeCommand", 1, availableRollsType.Id).Return(availableRollsType, nil) // Minimum: 0
			databaseMock.On("GetUserPointCommand", 2, 1, availableRollsType.Id).
				Return(typepoints.UserPoint{Value: 3}, nil)
			// 3 - 5 = -2 < Minimum(0) -> clamp to 0, actual change = -3
			databaseMock.On("ChangeUserPointValueCommand", 2, 1, availableRollsType.Id, -3).Return(nil)
			databaseMock.On("CreateUserPointHistoryCommand", 2, 1, availableRollsType.Id, 9, -5, -3, 0, 999).
				Return(typepoints.UserPointHistoryEntry{}, nil)
			return databaseMock
		},
	},
	{
		// A user with no existing row (sql.ErrNoRows) is treated as starting from 0.
		Name:        "NoExistingRow_StartsAtZero",
		ChangeValue: 1,
		SetupMock: func() *dbpointsmock.DatabaseMock {
			databaseMock := new(dbpointsmock.DatabaseMock)
			databaseMock.On("GetPointTypeCommand", 1, availableRollsType.Id).Return(availableRollsType, nil)
			databaseMock.On("GetUserPointCommand", 2, 1, availableRollsType.Id).Return(typepoints.UserPoint{}, sql.ErrNoRows)
			databaseMock.On("ChangeUserPointValueCommand", 2, 1, availableRollsType.Id, 1).Return(nil)
			databaseMock.On("CreateUserPointHistoryCommand", 2, 1, availableRollsType.Id, 9, 1, 1, 1, 999).
				Return(typepoints.UserPointHistoryEntry{}, nil)
			return databaseMock
		},
	},
	{
		// Within bounds, the requested change applies unmodified.
		Name:        "NoClamp",
		ChangeValue: 2,
		SetupMock: func() *dbpointsmock.DatabaseMock {
			databaseMock := new(dbpointsmock.DatabaseMock)
			databaseMock.On("GetPointTypeCommand", 1, availableRollsType.Id).Return(availableRollsType, nil)
			databaseMock.On("GetUserPointCommand", 2, 1, availableRollsType.Id).
				Return(typepoints.UserPoint{Value: 3}, nil)
			databaseMock.On("ChangeUserPointValueCommand", 2, 1, availableRollsType.Id, 2).Return(nil)
			databaseMock.On("CreateUserPointHistoryCommand", 2, 1, availableRollsType.Id, 9, 2, 2, 5, 999).
				Return(typepoints.UserPointHistoryEntry{}, nil)
			return databaseMock
		},
	},
	{
		// ChangeUserPointValueCommand fails. The error returns and history isn't recorded.
		Name:        "ChangeUserPointValue_DatabaseError",
		ChangeValue: 2,
		SetupMock: func() *dbpointsmock.DatabaseMock {
			databaseMock := new(dbpointsmock.DatabaseMock)
			databaseMock.On("GetPointTypeCommand", 1, availableRollsType.Id).Return(availableRollsType, nil)
			databaseMock.On("GetUserPointCommand", 2, 1, availableRollsType.Id).
				Return(typepoints.UserPoint{Value: 3}, nil)
			databaseMock.On("ChangeUserPointValueCommand", 2, 1, availableRollsType.Id, 2).Return(dbError)
			return databaseMock
		},
		ExpectedError: dbError,
	},
	{
		// ChangeUserPointValueCommand succeeds but CreateUserPointHistoryCommand fails. The error returns.
		Name:        "CreateUserPointHistory_DatabaseError",
		ChangeValue: 2,
		SetupMock: func() *dbpointsmock.DatabaseMock {
			databaseMock := new(dbpointsmock.DatabaseMock)
			databaseMock.On("GetPointTypeCommand", 1, availableRollsType.Id).Return(availableRollsType, nil)
			databaseMock.On("GetUserPointCommand", 2, 1, availableRollsType.Id).
				Return(typepoints.UserPoint{Value: 3}, nil)
			databaseMock.On("ChangeUserPointValueCommand", 2, 1, availableRollsType.Id, 2).Return(nil)
			databaseMock.On("CreateUserPointHistoryCommand", 2, 1, availableRollsType.Id, 9, 2, 2, 5, 999).
				Return(typepoints.UserPointHistoryEntry{}, dbError)
			return databaseMock
		},
		ExpectedError: dbError,
	},
}

func TestSrvPoints_ChangeUserPointValueClamped(test *testing.T) {
	for _, testCase := range ChangeUserPointValueClampedTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			databaseMock := testCase.SetupMock()
			sut := srvpoints.Service{Database: databaseMock}

			// Act
			err := sut.ChangeUserPointValueClamped(2, 1, availableRollsType.Id, testCase.ChangeValue, 9, 999)

			// Assert
			if testCase.ExpectedError != nil {
				require.ErrorIs(test, err, testCase.ExpectedError)
			} else {
				require.NoError(test, err)
			}

			databaseMock.AssertExpectations(test)
		})
	}
}

// --- ChangePointValueByTypeNameNoHistory ---

func TestSrvPoints_ChangePointValueByTypeNameNoHistory(test *testing.T) {
	test.Run("NameNotFound_Error", func(test *testing.T) {
		databaseMock := new(dbpointsmock.DatabaseMock)
		databaseMock.On("GetPointTypeByNameCommand", 1, typepoints.PointTypeAvailableRolls).Return(typepoints.PointTypeInfo{}, sql.ErrNoRows)

		sut := srvpoints.Service{Database: databaseMock}

		err := sut.ChangePointValueByTypeNameNoHistory(2, 1, typepoints.PointTypeAvailableRolls, -1)

		require.Error(test, err)
		databaseMock.AssertExpectations(test)
	})

	test.Run("Success_NoHistoryRecorded", func(test *testing.T) {
		databaseMock := new(dbpointsmock.DatabaseMock)
		databaseMock.On("GetPointTypeByNameCommand", 1, typepoints.PointTypeAvailableRolls).Return(availableRollsType, nil)
		databaseMock.On("GetUserPointCommand", 2, 1, availableRollsType.Id).
			Return(typepoints.UserPoint{Value: 5}, nil)
		databaseMock.On("ChangeUserPointValueCommand", 2, 1, availableRollsType.Id, -1).Return(nil)
		// Deliberately no "CreateUserPointHistoryCommand" expectation - AssertExpectations only
		// checks the calls that were declared, but calling it unexpectedly would panic the mock.

		sut := srvpoints.Service{Database: databaseMock}

		err := sut.ChangePointValueByTypeNameNoHistory(2, 1, typepoints.PointTypeAvailableRolls, -1)

		require.NoError(test, err)
		databaseMock.AssertExpectations(test)
		databaseMock.AssertNotCalled(test, "CreateUserPointHistoryCommand")
	})
}
