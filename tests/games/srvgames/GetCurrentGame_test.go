package srvgames_test

import (
	"FGG-Service/src/common"
	"FGG-Service/src/games/srvgames"
	"FGG-Service/src/games/typegames"
	"FGG-Service/tests/games/dbgamesmock"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type GetCurrentGameTestCase struct {
	Name            string
	UserId          int
	SetupMock       func() *dbgamesmock.DatabaseMock
	ExpectedGame    *typegames.CurrentGame
	ExpectedErrorAs any
	ExpectedErrorIs error
}

var GetCurrentGameTestCases = []GetCurrentGameTestCase{
	{
		// GetCurrentGameCommand returns sql.ErrNoRows. The CurrentGameNotFoundError will return.
		Name:   "NotFound_NoRows",
		UserId: 1,
		SetupMock: func() *dbgamesmock.DatabaseMock {
			databaseMock := new(dbgamesmock.DatabaseMock)

			databaseMock.
				On("GetCurrentGameCommand",
					1, 1).
				Return(typegames.UserGame{}, sql.ErrNoRows)

			return databaseMock
		},
		ExpectedErrorAs: new(common.NotFoundError),
	},
	{
		// GetCurrentGameCommand returns a database error. The error will return.
		Name:   "DatabaseError",
		UserId: 1,
		SetupMock: func() *dbgamesmock.DatabaseMock {
			databaseMock := new(dbgamesmock.DatabaseMock)

			databaseMock.
				On("GetCurrentGameCommand",
					1, 1).
				Return(typegames.UserGame{}, dbError)

			return databaseMock
		},
		ExpectedErrorIs: dbError,
	},
	{
		// GetCurrentGameCommand succeeds. The current game (with TimeSpent already scanned) will return.
		Name:   "SuccessReturn",
		UserId: 1,
		SetupMock: func() *dbgamesmock.DatabaseMock {
			databaseMock := new(dbgamesmock.DatabaseMock)

			databaseMock.
				On("GetCurrentGameCommand",
					1, 1).
				Return(typegames.UserGame{
					Id:        1,
					Name:      "Half-Life 1",
					TimeSpent: 2 * time.Hour,
					StartDate: time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)}, nil)

			return databaseMock
		},
		ExpectedGame: &typegames.CurrentGame{
			Id:        1,
			Name:      "Half-Life 1",
			TimeSpent: 2 * time.Hour,
			StartDate: time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)},
	},
}

func TestSrvGames_GetCurrentGame(test *testing.T) {
	for _, testCase := range GetCurrentGameTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			databaseMock := testCase.SetupMock()
			sut := srvgames.GettingService{Database: databaseMock}

			// Act
			game, err := sut.GetCurrentGame(test.Context(), testCase.UserId, 1)

			// Assert
			if testCase.ExpectedErrorAs != nil {
				require.Error(test, err)
				require.ErrorAs(test, err, &testCase.ExpectedErrorAs)
			}

			if testCase.ExpectedErrorIs != nil {
				require.ErrorIs(test, err, testCase.ExpectedErrorIs)
			}

			if testCase.ExpectedGame != nil {
				require.NoError(test, err)
				require.Equal(test, *testCase.ExpectedGame, game)
			}

			databaseMock.AssertExpectations(test)
		})
	}
}
