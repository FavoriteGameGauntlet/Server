package srvgames_test

import (
	srvgames "FGG-Service/src/games/service"
	typegames "FGG-Service/src/games/types"
	dbgamesmock "FGG-Service/tests/games/mock"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type RateGameCreatedTestCase struct {
	Name            string
	UpdatedDate     time.Time
	ExpectedCreated bool
}

var ratedDate = time.Date(2026, 1, 3, 12, 0, 0, 0, time.UTC)

var RateGameCreatedTestCases = []RateGameCreatedTestCase{
	// A rating inserted now has both dates set to the same NOW(). created is true.
	{Name: "FirstRating_Created", UpdatedDate: ratedDate, ExpectedCreated: true},
	// A rating replaced later has moved its UpdatedDate. created is false.
	{Name: "ReplacedRating_NotCreated", UpdatedDate: ratedDate.Add(time.Hour), ExpectedCreated: false},
}

func TestSrvGames_RateGame_Created(test *testing.T) {
	for _, testCase := range RateGameCreatedTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			databaseMock := new(dbgamesmock.DatabaseMock)
			databaseMock.On("GetGameHistoryCommand", 1, 1).Return([]typegames.GameHistoryEntry{
				{GameId: 7, Name: "Doom", EndState: gameHistoryStatePtr("finished")},
			}, nil)
			databaseMock.On("RateGameCommand", 1, 1, 7, 8, (*string)(nil)).Return(typegames.GameRating{
				CreatedDate: ratedDate,
				UpdatedDate: testCase.UpdatedDate,
			}, nil)

			sut := srvgames.Service{Database: databaseMock}

			// Act
			created, err := sut.RateGame(1, 1, "Doom", 8, nil)

			// Assert
			require.NoError(test, err)
			require.Equal(test, testCase.ExpectedCreated, created)
			databaseMock.AssertExpectations(test)
		})
	}
}
