package srvgames_test

import (
	"FGG-Service/src/common"
	"FGG-Service/src/games/service"
	"FGG-Service/src/games/types"
	"FGG-Service/tests/games/mock"
	"testing"

	"github.com/stretchr/testify/require"
)

type AddWishlistGameTestCase struct {
	Name            string
	UserId          int
	WishlistGame    typegames.WishlistGame
	SetupMock       func() *dbgamesmock.DatabaseMock
	ExpectedErrorAs interface{}
	ExpectedErrorIs error
}

var AddWishlistGameTestCases = []AddWishlistGameTestCase{
	{
		// DoesGameExistCommand returns a database error. The error will return.
		Name:         "DoesGameExist_DatabaseError",
		UserId:       1,
		WishlistGame: typegames.WishlistGame{GameId: 1, Name: "Half-Life 1"},
		SetupMock: func() *dbgamesmock.DatabaseMock {
			databaseMock := new(dbgamesmock.DatabaseMock)

			databaseMock.
				On("DoesGameExistCommand", 1, "Half-Life 1").
				Return(false, dbError)

			return databaseMock
		},
		ExpectedErrorIs: dbError,
	},
	{
		// The game exists. GetWishlistGameCommand returns a database error. The error will return.
		Name:         "GetWishlistGame_DatabaseError",
		UserId:       1,
		WishlistGame: typegames.WishlistGame{GameId: 1, Name: "Half-Life 1"},
		SetupMock: func() *dbgamesmock.DatabaseMock {
			databaseMock := new(dbgamesmock.DatabaseMock)

			databaseMock.
				On("DoesGameExistCommand", 1, "Half-Life 1").
				Return(true, nil)
			databaseMock.
				On("GetWishlistGameCommand", "Half-Life 1").
				Return(typegames.WishlistGame{}, dbError)

			return databaseMock
		},
		ExpectedErrorIs: dbError,
	},
	{
		// The game doesn't exist. CreateGameCommand returns a database error. The error will return.
		Name:         "CreateGame_DatabaseError",
		UserId:       1,
		WishlistGame: typegames.WishlistGame{GameId: 1, Name: "Half-Life 1"},
		SetupMock: func() *dbgamesmock.DatabaseMock {
			databaseMock := new(dbgamesmock.DatabaseMock)

			databaseMock.
				On("DoesGameExistCommand", 1, "Half-Life 1").
				Return(false, nil)
			databaseMock.
				On("CreateGameCommand", 1, "Half-Life 1").
				Return(typegames.Game{}, dbError)

			return databaseMock
		},
		ExpectedErrorIs: dbError,
	},
	{
		// The game doesn't exist. CreateGameCommand succeeds. GetWishlistGameCommand returns a database error. The error will return.
		Name:         "GetWishlistGameAfterCreate_DatabaseError",
		UserId:       1,
		WishlistGame: typegames.WishlistGame{GameId: 1, Name: "Half-Life 1"},
		SetupMock: func() *dbgamesmock.DatabaseMock {
			databaseMock := new(dbgamesmock.DatabaseMock)

			databaseMock.
				On("DoesGameExistCommand", 1, "Half-Life 1").
				Return(false, nil)
			databaseMock.
				On("CreateGameCommand", 1, "Half-Life 1").
				Return(typegames.Game{Id: 1, PartyId: 1, Name: "Half-Life 1"}, nil)
			databaseMock.
				On("GetWishlistGameCommand", "Half-Life 1").
				Return(typegames.WishlistGame{}, dbError)

			return databaseMock
		},
		ExpectedErrorIs: dbError,
	},
	{
		// DoesWishlistGameExistCommand returns a database error. The error will return.
		Name:         "DoesWishlistGameExist_DatabaseError",
		UserId:       1,
		WishlistGame: typegames.WishlistGame{GameId: 1, Name: "Half-Life 1"},
		SetupMock: func() *dbgamesmock.DatabaseMock {
			databaseMock := new(dbgamesmock.DatabaseMock)

			databaseMock.
				On("DoesGameExistCommand", 1, "Half-Life 1").
				Return(true, nil)
			databaseMock.
				On("GetWishlistGameCommand", "Half-Life 1").
				Return(typegames.WishlistGame{GameId: 1, Name: "Half-Life 1"}, nil)
			databaseMock.
				On("DoesWishlistGameExistCommand", 1, 1, 1).
				Return(false, dbError)

			return databaseMock
		},
		ExpectedErrorIs: dbError,
	},
	{
		// DoesWishlistGameExistCommand returns true. The WishlistGameAlreadyExistsConflictError will return.
		Name:         "AlreadyExists",
		UserId:       1,
		WishlistGame: typegames.WishlistGame{GameId: 1, Name: "Half-Life 1"},
		SetupMock: func() *dbgamesmock.DatabaseMock {
			databaseMock := new(dbgamesmock.DatabaseMock)

			databaseMock.
				On("DoesGameExistCommand", 1, "Half-Life 1").
				Return(true, nil)
			databaseMock.
				On("GetWishlistGameCommand", "Half-Life 1").
				Return(typegames.WishlistGame{GameId: 1, Name: "Half-Life 1"}, nil)
			databaseMock.
				On("DoesWishlistGameExistCommand", 1, 1, 1).
				Return(true, nil)

			return databaseMock
		},
		ExpectedErrorAs: new(common.ConflictError),
	},
	{
		// CreateWishlistGameCommand returns a database error. The error will return.
		Name:         "CreateWishlistGame_DatabaseError",
		UserId:       1,
		WishlistGame: typegames.WishlistGame{GameId: 1, Name: "Half-Life 1"},
		SetupMock: func() *dbgamesmock.DatabaseMock {
			databaseMock := new(dbgamesmock.DatabaseMock)

			databaseMock.
				On("DoesGameExistCommand", 1, "Half-Life 1").
				Return(true, nil)
			databaseMock.
				On("GetWishlistGameCommand", "Half-Life 1").
				Return(typegames.WishlistGame{GameId: 1, Name: "Half-Life 1"}, nil)
			databaseMock.
				On("DoesWishlistGameExistCommand", 1, 1, 1).
				Return(false, nil)
			databaseMock.
				On("CreateWishlistGameCommand", 1, 1, 1).
				Return(typegames.CreatedWishlistGame{}, dbError)

			return databaseMock
		},
		ExpectedErrorIs: dbError,
	},
	{
		// The game exists. The wishlist entry will be created successfully.
		Name:         "SuccessReturn",
		UserId:       1,
		WishlistGame: typegames.WishlistGame{GameId: 1, Name: "Half-Life 1"},
		SetupMock: func() *dbgamesmock.DatabaseMock {
			databaseMock := new(dbgamesmock.DatabaseMock)

			databaseMock.
				On("DoesGameExistCommand", 1, "Half-Life 1").
				Return(true, nil)
			databaseMock.
				On("GetWishlistGameCommand", "Half-Life 1").
				Return(typegames.WishlistGame{GameId: 1, Name: "Half-Life 1"}, nil)
			databaseMock.
				On("DoesWishlistGameExistCommand", 1, 1, 1).
				Return(false, nil)
			databaseMock.
				On("CreateWishlistGameCommand", 1, 1, 1).
				Return(typegames.CreatedWishlistGame{Id: 1, UserId: 1, PartyId: 1, GameId: 1}, nil)

			return databaseMock
		},
	},
	{
		// The game doesn't exist. The game will be created and the wishlist entry created successfully.
		Name:         "SuccessCreateAndReturn",
		UserId:       1,
		WishlistGame: typegames.WishlistGame{GameId: 1, Name: "Half-Life 1"},
		SetupMock: func() *dbgamesmock.DatabaseMock {
			databaseMock := new(dbgamesmock.DatabaseMock)

			databaseMock.
				On("DoesGameExistCommand", 1, "Half-Life 1").
				Return(false, nil)
			databaseMock.
				On("CreateGameCommand", 1, "Half-Life 1").
				Return(typegames.Game{Id: 1, PartyId: 1, Name: "Half-Life 1"}, nil)
			databaseMock.
				On("GetWishlistGameCommand", "Half-Life 1").
				Return(typegames.WishlistGame{GameId: 1, Name: "Half-Life 1"}, nil)
			databaseMock.
				On("DoesWishlistGameExistCommand", 1, 1, 1).
				Return(false, nil)
			databaseMock.
				On("CreateWishlistGameCommand", 1, 1, 1).
				Return(typegames.CreatedWishlistGame{Id: 1, UserId: 1, PartyId: 1, GameId: 1}, nil)

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
			err := sut.AddWishlistGame(testCase.UserId, testCase.WishlistGame)

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
