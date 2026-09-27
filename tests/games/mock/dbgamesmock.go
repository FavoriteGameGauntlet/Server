package dbgamesmock

import (
	"FGG-Service/src/games/types"
	"time"

	"github.com/stretchr/testify/mock"
)

type DatabaseMock struct {
	mock.Mock
}

func (m *DatabaseMock) GetGameByNameCommand(partyId int, name string) (game typegames.Game, err error) {
	args := m.Called(partyId, name)
	game = args.Get(0).(typegames.Game)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) CreateGameCommand(partyId int, name string) (game typegames.Game, err error) {
	args := m.Called(partyId, name)
	game = args.Get(0).(typegames.Game)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetGameCommand(partyId int, gameId int) (game typegames.Game, err error) {
	args := m.Called(partyId, gameId)
	game = args.Get(0).(typegames.Game)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetWishlistGameCommand(userId int, partyId int, gameId int) (game typegames.WishlistGame, err error) {
	args := m.Called(userId, partyId, gameId)
	game = args.Get(0).(typegames.WishlistGame)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) CreateWishlistGameCommand(userId int, partyId int, gameId int) (game typegames.CreatedWishlistGame, err error) {
	args := m.Called(userId, partyId, gameId)
	game = args.Get(0).(typegames.CreatedWishlistGame)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) DeleteUnplayedGameCommand(userId int, partyId int, gameId int) error {
	args := m.Called(userId, partyId, gameId)
	return args.Error(0)
}

func (m *DatabaseMock) GetWishlistGamesCommand(userId int, partyId int) (games typegames.WishlistGames, err error) {
	args := m.Called(userId, partyId)
	games = args.Get(0).(typegames.WishlistGames)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) CreateCurrentGameCommand(userId int, partyId int, gameId int, actorUserId int, sourceEventId *int) (game typegames.CreatedUserGame, err error) {
	args := m.Called(userId, partyId, gameId, actorUserId, sourceEventId)
	game = args.Get(0).(typegames.CreatedUserGame)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetCurrentGameCommand(userId int, partyId int) (game typegames.UserGame, err error) {
	args := m.Called(userId, partyId)
	game = args.Get(0).(typegames.UserGame)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetGameTimeSpentCommand(userId int, gameId int) (timeSpent time.Duration, err error) {
	args := m.Called(userId, gameId)
	timeSpent = args.Get(0).(time.Duration)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) ChangeGameTimeSpentCommand(userId int, partyId int, gameId int, changeValue time.Duration, actorUserId int, sourceEventId *int) error {
	args := m.Called(userId, partyId, gameId, changeValue, actorUserId, sourceEventId)
	return args.Error(0)
}

func (m *DatabaseMock) CancelCurrentGameCommand(userId int, partyId int, gameId int, actorUserId int, sourceEventId *int) error {
	args := m.Called(userId, partyId, gameId, actorUserId, sourceEventId)
	return args.Error(0)
}

func (m *DatabaseMock) FinishCurrentGameCommand(userId int, partyId int, gameId int, actorUserId int, sourceEventId *int) error {
	args := m.Called(userId, partyId, gameId, actorUserId, sourceEventId)
	return args.Error(0)
}

func (m *DatabaseMock) RateGameCommand(userId int, partyId int, gameId int, rating int, reviewComment *string) (gameRating typegames.GameRating, err error) {
	args := m.Called(userId, partyId, gameId, rating, reviewComment)
	gameRating, _ = args.Get(0).(typegames.GameRating)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetGameReviewCommand(userId int, partyId int, gameId int) (review typegames.GameReview, err error) {
	args := m.Called(userId, partyId, gameId)
	review, _ = args.Get(0).(typegames.GameReview)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetGameHistoryCommand(userId int, partyId int) (games []typegames.GameHistoryEntry, err error) {
	args := m.Called(userId, partyId)
	games = args.Get(0).([]typegames.GameHistoryEntry)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetAllCurrentGamesCommand(partyId int) (games []typegames.UserGameWithLogin, err error) {
	args := m.Called(partyId)
	games = args.Get(0).([]typegames.UserGameWithLogin)
	err = args.Error(1)
	return
}
