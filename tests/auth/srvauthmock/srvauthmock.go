package srvauthmock

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
)

type ServiceMock struct {
	mock.Mock
}

func (m *ServiceMock) DoesUserSessionExist(ctx echo.Context) (doesExist bool, err error) {
	args := m.Called(ctx)
	doesExist = args.Get(0).(bool)
	err = args.Error(1)
	return
}

func (m *ServiceMock) GetUserId(ctx echo.Context) (userId int, err error) {
	args := m.Called(ctx)
	userId = args.Get(0).(int)
	err = args.Error(1)
	return
}

func (m *ServiceMock) GetUserIdByLogin(_ context.Context, login string) (userId int, err error) {
	args := m.Called(login)
	userId = args.Get(0).(int)
	err = args.Error(1)
	return
}

func (m *ServiceMock) IsAdmin(_ context.Context, userId int, partyId int) (isAdmin bool, err error) {
	args := m.Called(userId, partyId)
	isAdmin = args.Get(0).(bool)
	err = args.Error(1)
	return
}

func (m *ServiceMock) GetLoginByUserId(_ context.Context, userId int) (login string, err error) {
	args := m.Called(userId)
	login = args.Get(0).(string)
	err = args.Error(1)
	return
}
