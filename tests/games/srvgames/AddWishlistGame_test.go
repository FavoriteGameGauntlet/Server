package srvgames_test

import (
	"FGG-Service/src/common"
	"FGG-Service/src/games/srvgames"
	"FGG-Service/src/games/typegames"
	"FGG-Service/tests/games/dbgamesmock"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

var halfLife = typegames.Game{Id: 1, PartyId: 1, Name: "Half-Life 1"}

type AddWishlistGameTestCase struct {
	Name            string
	UserId          int
	WishlistGame    typegames.WishlistGame
	SetupMock       func() *dbgamesmock.DatabaseMock
	ExpectedErrorAs any
	ExpectedErrorIs error
}

var AddWishlistGameTestCases = []AddWishlistGameTestCase{
	{
		// GetGameByNameCommand returns a database error. The error will return.
		Name:         "GetGameByName_DatabaseError",
		UserId:       1,
		WishlistGame: typegames.WishlistGame{Name: "Half-Life 1"},
		SetupMock: func() *dbgamesmock.DatabaseMock {
			databaseMock := new(dbgamesmock.DatabaseMock)

			databaseMock.
				On("GetGameByNameCommand", 1, "Half-Life 1").
				Return(typegames.Game{}, dbError)

			return databaseMock
		},
		ExpectedErrorIs: dbError,
	},
	{
		// Nobody has named the game before, so it is registered first. CreateGameCommand returns a
		// database error. The error will return.
		Name:         "CreateGame_DatabaseError",
		UserId:       1,
		WishlistGame: typegames.WishlistGame{Name: "Half-Life 1"},
		SetupMock: func() *dbgamesmock.DatabaseMock {
			databaseMock := new(dbgamesmock.DatabaseMock)

			databaseMock.
				On("GetGameByNameCommand", 1, "Half-Life 1").
				Return(typegames.Game{}, sql.ErrNoRows)
			databaseMock.
				On("CreateGameCommand", 1, "Half-Life 1").
				Return(typegames.Game{}, dbError)

			return databaseMock
		},
		ExpectedErrorIs: dbError,
	},
	{
		// GetWishlistGameCommand fails for a reason other than a missing row. The error will return
		// rather than being read as "not on the wishlist yet".
		Name:         "GetWishlistGame_DatabaseError",
		UserId:       1,
		WishlistGame: typegames.WishlistGame{Name: "Half-Life 1"},
		SetupMock: func() *dbgamesmock.DatabaseMock {
			databaseMock := new(dbgamesmock.DatabaseMock)

			databaseMock.
				On("GetGameByNameCommand", 1, "Half-Life 1").
				Return(halfLife, nil)
			databaseMock.
				On("GetWishlistGameCommand", 1, 1, halfLife.Id).
				Return(typegames.WishlistGame{}, dbError)

			return databaseMock
		},
		ExpectedErrorIs: dbError,
	},
	{
		// GetWishlistGameCommand finds a row, so the game is already on the wishlist. The
		// WishlistGameAlreadyExistsConflictError will return.
		Name:         "AlreadyExists",
		UserId:       1,
		WishlistGame: typegames.WishlistGame{Name: "Half-Life 1"},
		SetupMock: func() *dbgamesmock.DatabaseMock {
			databaseMock := new(dbgamesmock.DatabaseMock)

			databaseMock.
				On("GetGameByNameCommand", 1, "Half-Life 1").
				Return(halfLife, nil)
			databaseMock.
				On("GetWishlistGameCommand", 1, 1, halfLife.Id).
				Return(typegames.WishlistGame{Id: 3, GameId: halfLife.Id, Name: halfLife.Name}, nil)

			return databaseMock
		},
		ExpectedErrorAs: new(common.ConflictError),
	},
	{
		// CreateWishlistGameCommand returns a database error. The error will return.
		Name:         "CreateWishlistGame_DatabaseError",
		UserId:       1,
		WishlistGame: typegames.WishlistGame{Name: "Half-Life 1"},
		SetupMock: func() *dbgamesmock.DatabaseMock {
			databaseMock := new(dbgamesmock.DatabaseMock)

			databaseMock.
				On("GetGameByNameCommand", 1, "Half-Life 1").
				Return(halfLife, nil)
			databaseMock.
				On("GetWishlistGameCommand", 1, 1, halfLife.Id).
				Return(typegames.WishlistGame{}, sql.ErrNoRows)
			databaseMock.
				On("CreateWishlistGameCommand", 1, 1, halfLife.Id).
				Return(typegames.CreatedWishlistGame{}, dbError)

			return databaseMock
		},
		ExpectedErrorIs: dbError,
	},
	{
		// The game is already known to the party. The wishlist entry will be created successfully.
		Name:         "SuccessReturn",
		UserId:       1,
		WishlistGame: typegames.WishlistGame{Name: "Half-Life 1"},
		SetupMock: func() *dbgamesmock.DatabaseMock {
			databaseMock := new(dbgamesmock.DatabaseMock)

			databaseMock.
				On("GetGameByNameCommand", 1, "Half-Life 1").
				Return(halfLife, nil)
			databaseMock.
				On("GetWishlistGameCommand", 1, 1, halfLife.Id).
				Return(typegames.WishlistGame{}, sql.ErrNoRows)
			databaseMock.
				On("CreateWishlistGameCommand", 1, 1, halfLife.Id).
				Return(typegames.CreatedWishlistGame{Id: 1, UserId: 1, PartyId: 1, GameId: halfLife.Id}, nil)

			return databaseMock
		},
	},
	{
		// Nobody has named the game before. It will be registered in the party and then put on the
		// wishlist, using the id the insert returned.
		Name:         "SuccessCreateAndReturn",
		UserId:       1,
		WishlistGame: typegames.WishlistGame{Name: "Half-Life 1"},
		SetupMock: func() *dbgamesmock.DatabaseMock {
			databaseMock := new(dbgamesmock.DatabaseMock)

			databaseMock.
				On("GetGameByNameCommand", 1, "Half-Life 1").
				Return(typegames.Game{}, sql.ErrNoRows)
			databaseMock.
				On("CreateGameCommand", 1, "Half-Life 1").
				Return(halfLife, nil)
			databaseMock.
				On("GetWishlistGameCommand", 1, 1, halfLife.Id).
				Return(typegames.WishlistGame{}, sql.ErrNoRows)
			databaseMock.
				On("CreateWishlistGameCommand", 1, 1, halfLife.Id).
				Return(typegames.CreatedWishlistGame{Id: 1, UserId: 1, PartyId: 1, GameId: halfLife.Id}, nil)

			return databaseMock
		},
	},
}

func TestSrvGames_AddWishlistGame(test *testing.T) {
	for _, testCase := range AddWishlistGameTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			databaseMock := testCase.SetupMock()
			sut := srvgames.Service{Database: databaseMock}

			// Act
			err := sut.AddWishlistGame(test.Context(), testCase.UserId, 1, testCase.WishlistGame)

			// Assert
			if testCase.ExpectedErrorAs != nil {
				require.Error(test, err)
				require.ErrorAs(test, err, &testCase.ExpectedErrorAs)
			}

			if testCase.ExpectedErrorIs != nil {
				require.ErrorIs(test, err, testCase.ExpectedErrorIs)
			}

			if testCase.ExpectedErrorAs == nil && testCase.ExpectedErrorIs == nil {
				require.NoError(test, err)
			}

			databaseMock.AssertExpectations(test)
		})
	}
}
