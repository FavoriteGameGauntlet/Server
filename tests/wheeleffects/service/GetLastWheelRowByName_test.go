package srvwheeleffects_test

import (
	"FGG-Service/src/common"
	"FGG-Service/src/wheeleffects/service"
	"FGG-Service/src/wheeleffects/types"
	"FGG-Service/tests/timers/mock/dbwheeleffects"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

type GetLastWheelRowByNameTestCase struct {
	Name              string
	UserId            int
	WheelRowName      string
	SetupMock         func() *dbwheeleffectsmock.DatabaseMock
	ExpectedRow       *typewheeleffects.LastWheelRow
	ExpectedError     error
	ExpectedErrorCode string
}

var GetLastWheelRowByNameTestCases = []GetLastWheelRowByNameTestCase{
	{
		// GetLastRolledWheelEffectsCommand returns a database error. The error returns as-is.
		Name:         "DatabaseError",
		UserId:       1,
		WheelRowName: "test-effect",
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
		// No last-rolled row matches the requested name. The WHEEL_EFFECT_NAME_NOT_FOUND error returns.
		Name:         "NameNotFound_Error",
		UserId:       1,
		WheelRowName: "unknown-effect",
		SetupMock: func() *dbwheeleffectsmock.DatabaseMock {
			databaseMock := new(dbwheeleffectsmock.DatabaseMock)

			databaseMock.
				On("GetLastRolledWheelEffectsCommand", 1, 1).
				Return([]typewheeleffects.LastWheelRow{
					{Id: 42, Name: "test-effect"},
				}, nil)

			return databaseMock
		},
		ExpectedErrorCode: "WHEEL_EFFECT_NAME_NOT_FOUND",
	},
	{
		// The name matches. Checking whether it was already applied fails. The error returns as-is.
		Name:         "AppliedCheck_DatabaseError",
		UserId:       1,
		WheelRowName: "test-effect",
		SetupMock: func() *dbwheeleffectsmock.DatabaseMock {
			databaseMock := new(dbwheeleffectsmock.DatabaseMock)

			databaseMock.
				On("GetLastRolledWheelEffectsCommand", 1, 1).
				Return([]typewheeleffects.LastWheelRow{
					{Id: 42, Name: "test-effect"},
				}, nil)
			databaseMock.
				On("GetEffectHistoryByEffectNameCommand", 1, 1, "test-effect").
				Return(typewheeleffects.WheelRowHistory{}, dbError)

			return databaseMock
		},
		ExpectedError: dbError,
	},
	{
		// A WheelRowHistory row already exists for the name (no error from the lookup). The
		// WHEEL_EFFECT_ROLL_ALREADY_APPLIED error returns.
		Name:         "AlreadyApplied_ConflictError",
		UserId:       1,
		WheelRowName: "test-effect",
		SetupMock: func() *dbwheeleffectsmock.DatabaseMock {
			databaseMock := new(dbwheeleffectsmock.DatabaseMock)

			databaseMock.
				On("GetLastRolledWheelEffectsCommand", 1, 1).
				Return([]typewheeleffects.LastWheelRow{
					{Id: 42, Name: "test-effect"},
				}, nil)
			databaseMock.
				On("GetEffectHistoryByEffectNameCommand", 1, 1, "test-effect").
				Return(typewheeleffects.WheelRowHistory{Name: "test-effect"}, nil)

			return databaseMock
		},
		ExpectedErrorCode: "WHEEL_EFFECT_ROLL_ALREADY_APPLIED",
	},
	{
		// The name matches and no WheelRowHistory row exists yet (sql.ErrNoRows). The row returns.
		Name:         "Success",
		UserId:       1,
		WheelRowName: "test-effect",
		SetupMock: func() *dbwheeleffectsmock.DatabaseMock {
			databaseMock := new(dbwheeleffectsmock.DatabaseMock)

			databaseMock.
				On("GetLastRolledWheelEffectsCommand", 1, 1).
				Return([]typewheeleffects.LastWheelRow{
					{Id: 42, Name: "test-effect"},
				}, nil)
			databaseMock.
				On("GetEffectHistoryByEffectNameCommand", 1, 1, "test-effect").
				Return(typewheeleffects.WheelRowHistory{}, sql.ErrNoRows)

			return databaseMock
		},
		ExpectedRow: &typewheeleffects.LastWheelRow{Id: 42, Name: "test-effect"},
	},
}

func TestSrvWheelEffects_GetLastWheelRowByName(test *testing.T) {
	for _, testCase := range GetLastWheelRowByNameTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			databaseMock := testCase.SetupMock()
			sut := srvwheeleffects.Service{Database: databaseMock}

			// Act
			row, err := sut.GetLastWheelRowByName(test.Context(), testCase.UserId, 1, testCase.WheelRowName)

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
				require.Equal(test, *testCase.ExpectedRow, row)
			}

			databaseMock.AssertExpectations(test)
		})
	}
}
