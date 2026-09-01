package srvgames_test

import (
	"FGG-Service/src/games/service"
	"FGG-Service/src/games/types"
	"FGG-Service/tests/games/mock"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func gameHistoryStatePtr(s string) *string { return &s }

var createdDate1 = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
var createdDate2 = time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)

type GetGameHistoryTestCase struct {
	Name            string
	UserId          int
	SetupMock       func() *dbgamesmock.DatabaseMock
	ExpectedGames   typegames.CurrentGames
	ExpectedErrorIs error
}

var GetGameHistoryTestCases = []GetGameHistoryTestCase{
	{
		// GetGameHistoryCommand returns a database error. The error will return.
		Name:   "GetGameHistoryCommand_DatabaseError",
		UserId: 1,
		SetupMock: func() *dbgamesmock.DatabaseMock {
			databaseMock := new(dbgamesmock.DatabaseMock)

			databaseMock.
				On("GetGameHistoryCommand",
					1, 1).
				Return([]typegames.GameHistoryEntry{}, dbError)

			return databaseMock
		},
		ExpectedErrorIs: dbError,
	},
	{
		// GetGameHistoryCommand succeeds for multiple games. Each entry maps to a CurrentGame keyed by
		// GameId, with TimeSpent/State already carried on the entry (no per-game follow-up query).
		Name:   "SuccessReturn_MultipleGames",
		UserId: 1,
		SetupMock: func() *dbgamesmock.DatabaseMock {
			databaseMock := new(dbgamesmock.DatabaseMock)

			databaseMock.
				On("GetGameHistoryCommand",
					1, 1).
				Return([]typegames.GameHistoryEntry{
					{
						GameId:      1,
						Name:        "Half-Life 1",
						TimeSpent:   2 * time.Hour,
						EndState:    gameHistoryStatePtr("finished"),
						CreatedDate: createdDate1,
					},
					{
						GameId:      2,
						Name:        "Half-Life 2",
						TimeSpent:   30 * time.Minute,
						EndState:    gameHistoryStatePtr("cancelled"),
						CreatedDate: createdDate2,
					},
				}, nil)

			return databaseMock
		},
		ExpectedGames: typegames.CurrentGames{
			typegames.CurrentGame{Id: 1, Name: "Half-Life 1", State: typegames.GameStateFinished, TimeSpent: 2 * time.Hour, StartDate: createdDate1},
			typegames.CurrentGame{Id: 2, Name: "Half-Life 2", State: typegames.GameStateCancelled, TimeSpent: 30 * time.Minute, StartDate: createdDate2},
		},
	},
}

func TestSrvGames_GetGameHistory(test *testing.T) {
	for _, testCase := range GetGameHistoryTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			databaseMock := testCase.SetupMock()
			sut := srvgames.Service{Database: databaseMock}

			// Act
			games, err := sut.GetGameHistory(testCase.UserId)

			// Assert
			if testCase.ExpectedErrorIs != nil {
				require.ErrorIs(test, err, testCase.ExpectedErrorIs)
			}

			if testCase.ExpectedGames != nil {
				require.NoError(test, err)
				require.Equal(test, testCase.ExpectedGames, games)
			}

			databaseMock.AssertExpectations(test)
		})
	}
}
