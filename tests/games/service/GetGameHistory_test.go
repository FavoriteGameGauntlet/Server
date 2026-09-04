package srvgames_test

import (
	srvgames "FGG-Service/src/games/service"
	typegames "FGG-Service/src/games/types"
	dbgamesmock "FGG-Service/tests/games/mock"
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
	ExpectedHistory []typegames.GameHistoryEntry
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
		// A history entry carries what happened to the game, so the entries return as they are
		// rather than being flattened into the current game shape.
		Name:   "SuccessReturn_KeepsWhatTheEntryCarries",
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
						Action:      "finished",
						TimeSpent:   2 * time.Hour,
						EndState:    gameHistoryStatePtr("finished"),
						CreatedDate: createdDate1,
					},
					{
						GameId:      2,
						Name:        "Half-Life 2",
						Action:      "cancelled",
						TimeSpent:   30 * time.Minute,
						EndState:    gameHistoryStatePtr("cancelled"),
						CreatedDate: createdDate2,
					},
				}, nil)

			return databaseMock
		},
		ExpectedHistory: []typegames.GameHistoryEntry{
			{
				GameId:      1,
				Name:        "Half-Life 1",
				Action:      "finished",
				TimeSpent:   2 * time.Hour,
				EndState:    gameHistoryStatePtr("finished"),
				CreatedDate: createdDate1,
			},
			{
				GameId:      2,
				Name:        "Half-Life 2",
				Action:      "cancelled",
				TimeSpent:   30 * time.Minute,
				EndState:    gameHistoryStatePtr("cancelled"),
				CreatedDate: createdDate2,
			},
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
			history, err := sut.GetGameHistory(testCase.UserId)

			// Assert
			if testCase.ExpectedErrorIs != nil {
				require.ErrorIs(test, err, testCase.ExpectedErrorIs)
			} else {
				require.NoError(test, err)
				require.Equal(test, testCase.ExpectedHistory, history)
			}

			databaseMock.AssertExpectations(test)
		})
	}
}