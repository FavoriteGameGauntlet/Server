package srvwheeleffects_test

import (
	"FGG-Service/src/common"
	"FGG-Service/src/wheeleffects/service"
	"FGG-Service/src/wheeleffects/types"
	"FGG-Service/tests/timers/mock/dbwheeleffects"
	"testing"

	"github.com/stretchr/testify/require"
)

type GetLastRolledWheelEffectsTestCase struct {
	Name              string
	UserId            int
	SetupMock         func() *dbwheeleffectsmock.DatabaseMock
	ExpectedRows      []typewheeleffects.LastWheelRow
	ExpectedError     error
	ExpectedErrorCode string
}

var GetLastRolledWheelEffectsTestCases = []GetLastRolledWheelEffectsTestCase{
	{
		// GetLastRolledWheelEffectsCommand returns a database error. The error returns as-is.
		Name:   "DatabaseError",
		UserId: 1,
		SetupMock: func() *dbwheeleffectsmock.DatabaseMock {
			databaseMock := new(dbwheeleffectsmock.DatabaseMock)

			databaseMock.
				On("GetLastRolledWheelEffectsCommand", 1, 1).
				Return([]typewheeleffects.LastWheelRow{}, dbError)

			return databaseMock
		},
		ExpectedError: dbError,
	},
	{
		// GetLastRolledWheelEffectsCommand returns an empty slice. The LAST_WHEEL_EFFECTS_NOT_FOUND error returns.
		Name:   "Empty_NotFoundError",
		UserId: 1,
		SetupMock: func() *dbwheeleffectsmock.DatabaseMock {
			databaseMock := new(dbwheeleffectsmock.DatabaseMock)

			databaseMock.
				On("GetLastRolledWheelEffectsCommand", 1, 1).
				Return([]typewheeleffects.LastWheelRow{}, nil)

			return databaseMock
		},
		ExpectedErrorCode: "LAST_WHEEL_EFFECTS_NOT_FOUND",
	},
	{
		// GetLastRolledWheelEffectsCommand returns rows. They return as-is.
		Name:   "Success",
		UserId: 1,
		SetupMock: func() *dbwheeleffectsmock.DatabaseMock {
			databaseMock := new(dbwheeleffectsmock.DatabaseMock)

			databaseMock.
				On("GetLastRolledWheelEffectsCommand", 1, 1).
				Return([]typewheeleffects.LastWheelRow{
					{Id: 42, Name: "test-effect", Description: "desc"},
				}, nil)

			return databaseMock
		},
		ExpectedRows: []typewheeleffects.LastWheelRow{
			{Id: 42, Name: "test-effect", Description: "desc"},
		},
	},
}

func TestSrvWheelEffects_GetLastRolledWheelEffects(test *testing.T) {
	for _, testCase := range GetLastRolledWheelEffectsTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			databaseMock := testCase.SetupMock()
			sut := srvwheeleffects.Service{Database: databaseMock}

			// Act
			rows, err := sut.GetLastRolledWheelEffects(testCase.UserId)

			// Assert
			if testCase.ExpectedErrorCode != "" {
				require.Error(test, err)
				appErr, ok := err.(common.AppError)
				require.True(test, ok)
				require.Equal(test, testCase.ExpectedErrorCode, appErr.GetCode())
			} else if testCase.ExpectedError != nil {
				require.ErrorIs(test, err, testCase.ExpectedError)
			} else {
				require.NoError(test, err)
				require.Equal(test, testCase.ExpectedRows, rows)
			}

			databaseMock.AssertExpectations(test)
		})
	}
}
