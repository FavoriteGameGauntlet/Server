package srvpartiesmock

import (
	"FGG-Service/src/parties/types"
	"context"

	"github.com/stretchr/testify/mock"
)

type ServiceMock struct {
	mock.Mock
}

func (m *ServiceMock) CreateParty(_ context.Context, creatorUserId int, name string) (typeparties.Party, error) {
	args := m.Called(creatorUserId, name)
	return args.Get(0).(typeparties.Party), args.Error(1)
}

func (m *ServiceMock) GetUserParties(_ context.Context, userId int) ([]typeparties.Party, error) {
	args := m.Called(userId)
	return args.Get(0).([]typeparties.Party), args.Error(1)
}

func (m *ServiceMock) RequireMember(_ context.Context, userId int, partyId int) error {
	args := m.Called(userId, partyId)
	return args.Error(0)
}

func (m *ServiceMock) RequireAdmin(_ context.Context, userId int, partyId int) error {
	args := m.Called(userId, partyId)
	return args.Error(0)
}

func (m *ServiceMock) GetParty(_ context.Context, partyId int) (typeparties.Party, error) {
	args := m.Called(partyId)
	return args.Get(0).(typeparties.Party), args.Error(1)
}

func (m *ServiceMock) ChangePartyName(_ context.Context, partyId int, name string) error {
	args := m.Called(partyId, name)
	return args.Error(0)
}

func (m *ServiceMock) GetMembers(_ context.Context, partyId int) ([]typeparties.MemberWithLogin, error) {
	args := m.Called(partyId)
	return args.Get(0).([]typeparties.MemberWithLogin), args.Error(1)
}

func (m *ServiceMock) AddMember(_ context.Context, userId int, partyId int, displayName *string, isAdmin bool) (typeparties.Member, error) {
	args := m.Called(userId, partyId, displayName, isAdmin)
	return args.Get(0).(typeparties.Member), args.Error(1)
}

func (m *ServiceMock) ChangeMember(_ context.Context, userId int, partyId int, displayName *string, isAdmin *bool) error {
	args := m.Called(userId, partyId, displayName, isAdmin)
	return args.Error(0)
}

func (m *ServiceMock) RemoveMember(_ context.Context, userId int, partyId int) error {
	args := m.Called(userId, partyId)
	return args.Error(0)
}
