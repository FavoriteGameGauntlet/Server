package srvgames_test

import (
	"FGG-Service/src/games/srvgames"
	"FGG-Service/src/games/typegames"
	"FGG-Service/tests/games/dbgamesmock"
	"testing"

	"github.com/stretchr/testify/require"
)

type GetWishlistGamesTestCase struct {
	Name            string
	UserId          int
	SetupMock       func() *dbgamesmock.DatabaseMock
	ExpectedGames   typegames.WishlistGames
	ExpectedErrorIs error
}

var GetWishlistGamesTestCases = []GetWishlistGamesTestCase{
	{
		// GetWishlistGamesCommand returns a database error. The error will return.
		Name:   "DatabaseError",
		UserId: 1,
		SetupMock: func() *dbgamesmock.DatabaseMock {
			databaseMock := new(dbgamesmock.DatabaseMock)

			databaseMock.
				On("GetWishlistGamesCommand",
					1, 1).
				Return(typegames.WishlistGames{}, dbError)

			return databaseMock
		},
		ExpectedErrorIs: dbError,
	},
	{
		// GetWishlistGamesCommand returns an empty list. An empty list will return.
		Name:   "EmptyList",
		UserId: 1,
		SetupMock: func() *dbgamesmock.DatabaseMock {
			databaseMock := new(dbgamesmock.DatabaseMock)

			databaseMock.
				On("GetWishlistGamesCommand",
					1, 1).
				Return(typegames.WishlistGames{}, nil)

			return databaseMock
		},
		ExpectedGames: typegames.WishlistGames{},
	},
	{
		// GetWishlistGamesCommand succeeds. The list of wishlist games will return.
		Name:   "SuccessReturn",
		UserId: 1,
		SetupMock: func() *dbgamesmock.DatabaseMock {
			databaseMock := new(dbgamesmock.DatabaseMock)

			databaseMock.
				On("GetWishlistGamesCommand",
					1, 1).
				Return(typegames.WishlistGames{
					{Id: 1, GameId: 10, Name: "Half-Life 1"},
					{Id: 2, GameId: 11, Name: "Half-Life 2"},
				}, nil)

			return databaseMock
		},
		ExpectedGames: typegames.WishlistGames{
			{Id: 1, GameId: 10, Name: "Half-Life 1"},
			{Id: 2, GameId: 11, Name: "Half-Life 2"},
		},
	},
}

func TestSrvGames_GetWishlistGames(test *testing.T) {
	for _, testCase := range GetWishlistGamesTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			databaseMock := testCase.SetupMock()
			sut := srvgames.GettingService{Database: databaseMock}

			// Act
			games, err := sut.GetWishlistGames(test.Context(), testCase.UserId, 1)

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
