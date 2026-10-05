package srvpartiesmock

import (
	"FGG-Service/src/parties/types"

	"github.com/stretchr/testify/mock"
)

type ServiceMock struct {
	mock.Mock
}

func (m *ServiceMock) CreateParty(creatorUserId int, name string) (typeparties.Party, error) {
	args := m.Called(creatorUserId, name)
	return args.Get(0).(typeparties.Party), args.Error(1)
}

func (m *ServiceMock) GetUserParties(userId int) ([]typeparties.Party, error) {
	args := m.Called(userId)
	return args.Get(0).([]typeparties.Party), args.Error(1)
}

func (m *ServiceMock) RequireMember(userId int, partyId int) error {
	args := m.Called(userId, partyId)
	return args.Error(0)
}

func (m *ServiceMock) RequireAdmin(userId int, partyId int) error {
	args := m.Called(userId, partyId)
	return args.Error(0)
}

func (m *ServiceMock) GetParty(partyId int) (typeparties.Party, error) {
	args := m.Called(partyId)
	return args.Get(0).(typeparties.Party), args.Error(1)
}

func (m *ServiceMock) ChangePartyName(partyId int, name string) error {
	args := m.Called(partyId, name)
	return args.Error(0)
}

func (m *ServiceMock) GetMembers(partyId int) ([]typeparties.MemberWithLogin, error) {
	args := m.Called(partyId)
	return args.Get(0).([]typeparties.MemberWithLogin), args.Error(1)
}

func (m *ServiceMock) AddMember(userId int, partyId int, displayName *string, isAdmin bool) (typeparties.Member, error) {
	args := m.Called(userId, partyId, displayName, isAdmin)
	return args.Get(0).(typeparties.Member), args.Error(1)
}

func (m *ServiceMock) ChangeMember(userId int, partyId int, displayName *string, isAdmin *bool) error {
	args := m.Called(userId, partyId, displayName, isAdmin)
	return args.Error(0)
}

func (m *ServiceMock) RemoveMember(userId int, partyId int) error {
	args := m.Called(userId, partyId)
	return args.Error(0)
}
