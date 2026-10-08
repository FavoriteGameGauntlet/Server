package srvgamesmock

import (
	"FGG-Service/src/games/typegames"
	"context"

	"github.com/stretchr/testify/mock"
)

type GettingServiceMock struct {
	mock.Mock
}

func (m *GettingServiceMock) GetWishlistGames(_ context.Context, userId int, partyId int) (typegames.WishlistGames, error) {
	args := m.Called(userId, partyId)
	return args.Get(0).(typegames.WishlistGames), args.Error(1)
}

func (m *GettingServiceMock) GetCurrentGame(_ context.Context, userId int, partyId int) (typegames.CurrentGame, error) {
	args := m.Called(userId, partyId)
	return args.Get(0).(typegames.CurrentGame), args.Error(1)
}

type ServiceMock struct {
	mock.Mock
}

func (m *ServiceMock) GetCurrentGame(_ context.Context, userId int, partyId int) (typegames.CurrentGame, error) {
	args := m.Called(userId, partyId)
	return args.Get(0).(typegames.CurrentGame), args.Error(1)
}

func (m *ServiceMock) CancelCurrentGame(_ context.Context, userId int, partyId int) (typegames.CurrentGame, error) {
	args := m.Called(userId, partyId)
	return args.Get(0).(typegames.CurrentGame), args.Error(1)
}

func (m *ServiceMock) FinishCurrentGame(_ context.Context, userId int, partyId int) (typegames.CurrentGame, error) {
	args := m.Called(userId, partyId)
	return args.Get(0).(typegames.CurrentGame), args.Error(1)
}

func (m *ServiceMock) StartCurrentGame(_ context.Context, userId int, partyId int) (typegames.CurrentGame, error) {
	args := m.Called(userId, partyId)
	return args.Get(0).(typegames.CurrentGame), args.Error(1)
}

func (m *ServiceMock) GetGameHistory(_ context.Context, userId int, partyId int) ([]typegames.GameHistoryEntry, error) {
	args := m.Called(userId, partyId)
	return args.Get(0).([]typegames.GameHistoryEntry), args.Error(1)
}

func (m *ServiceMock) RateGame(_ context.Context, userId int, partyId int, name string, rating int, reviewComment *string) (bool, error) {
	args := m.Called(userId, partyId, name, rating, reviewComment)
	return args.Bool(0), args.Error(1)
}

func (m *ServiceMock) GetGameReview(_ context.Context, userId int, partyId int, name string) (typegames.GameReview, error) {
	args := m.Called(userId, partyId, name)
	return args.Get(0).(typegames.GameReview), args.Error(1)
}

func (m *ServiceMock) GetWishlistGames(_ context.Context, userId int, partyId int) (typegames.WishlistGames, error) {
	args := m.Called(userId, partyId)
	return args.Get(0).(typegames.WishlistGames), args.Error(1)
}

func (m *ServiceMock) AddWishlistGame(_ context.Context, userId int, partyId int, wishlistGame typegames.WishlistGame) error {
	args := m.Called(userId, partyId, wishlistGame)
	return args.Error(0)
}

func (m *ServiceMock) GetAllCurrentGames(_ context.Context, partyId int) ([]typegames.CurrentGameWithLogin, error) {
	args := m.Called(partyId)
	return args.Get(0).([]typegames.CurrentGameWithLogin), args.Error(1)
}
