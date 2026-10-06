package srvwheeleffects_test

import (
	"FGG-Service/src/wheeleffects/service"
	"FGG-Service/tests/timers/mock/dbwheeleffects"
	"testing"

	"github.com/stretchr/testify/require"
)

type ClearLastWheelEffectsTestCase struct {
	Name          string
	UserId        int
	SetupMock     func() *dbwheeleffectsmock.DatabaseMock
	ExpectedError error
}

var ClearLastWheelEffectsTestCases = []ClearLastWheelEffectsTestCase{
	{
		// ClearLastWheelEffectsCommand succeeds. No error returns.
		Name:   "Success",
		UserId: 1,
		SetupMock: func() *dbwheeleffectsmock.DatabaseMock {
			databaseMock := new(dbwheeleffectsmock.DatabaseMock)

			databaseMock.On("ClearLastWheelEffectsCommand", 1, 1).Return(nil)

			return databaseMock
		},
	},
	{
		// ClearLastWheelEffectsCommand returns a database error. The error returns as-is.
		Name:   "DatabaseError",
		UserId: 1,
		SetupMock: func() *dbwheeleffectsmock.DatabaseMock {
			databaseMock := new(dbwheeleffectsmock.DatabaseMock)

			databaseMock.On("ClearLastWheelEffectsCommand", 1, 1).Return(dbError)

			return databaseMock
		},
		ExpectedError: dbError,
	},
}

func TestSrvWheelEffects_ClearLastWheelEffects(test *testing.T) {
	for _, testCase := range ClearLastWheelEffectsTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			databaseMock := testCase.SetupMock()
			sut := srvwheeleffects.Service{Database: databaseMock}

			// Act
			err := sut.ClearLastWheelEffects(test.Context(), testCase.UserId, 1)

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
