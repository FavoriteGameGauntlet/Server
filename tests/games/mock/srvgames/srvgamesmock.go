package srvgamesmock

import (
	typegames "FGG-Service/src/games/types"

	"github.com/stretchr/testify/mock"
)

type GettingServiceMock struct {
	mock.Mock
}

func (m *GettingServiceMock) GetWishlistGames(userId int) (typegames.WishlistGames, error) {
	args := m.Called(userId)
	return args.Get(0).(typegames.WishlistGames), args.Error(1)
}

func (m *GettingServiceMock) GetCurrentGame(userId int) (typegames.CurrentGame, error) {
	args := m.Called(userId)
	return args.Get(0).(typegames.CurrentGame), args.Error(1)
}

type ServiceMock struct {
	mock.Mock
}

func (m *ServiceMock) GetCurrentGame(userId int) (typegames.CurrentGame, error) {
	args := m.Called(userId)
	return args.Get(0).(typegames.CurrentGame), args.Error(1)
}

func (m *ServiceMock) CancelCurrentGame(userId int) error {
	args := m.Called(userId)
	return args.Error(0)
}

func (m *ServiceMock) FinishCurrentGame(userId int) error {
	args := m.Called(userId)
	return args.Error(0)
}

func (m *ServiceMock) MakeGameRoll(userId int) (typegames.CurrentGame, error) {
	args := m.Called(userId)
	return args.Get(0).(typegames.CurrentGame), args.Error(1)
}

func (m *ServiceMock) GetGameHistory(userId int) ([]typegames.GameHistoryEntry, error) {
	args := m.Called(userId)
	return args.Get(0).([]typegames.GameHistoryEntry), args.Error(1)
}

func (m *ServiceMock) RateCurrentGame(userId int, rating int, reviewComment *string) error {
	args := m.Called(userId, rating, reviewComment)
	return args.Error(0)
}

func (m *ServiceMock) GetUnplayedGames(userId int) (typegames.WishlistGames, error) {
	args := m.Called(userId)
	return args.Get(0).(typegames.WishlistGames), args.Error(1)
}

func (m *ServiceMock) AddWishlistGame(userId int, wishlistGame typegames.WishlistGame) error {
	args := m.Called(userId, wishlistGame)
	return args.Error(0)
}

func (m *ServiceMock) GetAllCurrentGames() ([]typegames.CurrentGameWithLogin, error) {
	args := m.Called()
	return args.Get(0).([]typegames.CurrentGameWithLogin), args.Error(1)
}
