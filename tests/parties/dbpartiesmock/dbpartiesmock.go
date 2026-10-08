package dbpartiesmock

import (
	"FGG-Service/src/parties/typeparties"
	"context"

	"github.com/stretchr/testify/mock"
)

type DatabaseMock struct {
	mock.Mock
}

func (m *DatabaseMock) CreatePartyCommand(_ context.Context, name string) (party typeparties.Party, err error) {
	args := m.Called(name)
	party = args.Get(0).(typeparties.Party)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetPartyCommand(_ context.Context, partyId int) (party typeparties.Party, err error) {
	args := m.Called(partyId)
	party = args.Get(0).(typeparties.Party)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetUserPartiesCommand(_ context.Context, userId int) (parties []typeparties.Party, err error) {
	args := m.Called(userId)
	parties = args.Get(0).([]typeparties.Party)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) ChangePartyNameCommand(_ context.Context, partyId int, name string) error {
	args := m.Called(partyId, name)
	return args.Error(0)
}

func (m *DatabaseMock) CreateMemberCommand(_ context.Context, userId int, partyId int, displayName *string, isAdmin bool) (member typeparties.Member, err error) {
	args := m.Called(userId, partyId, displayName, isAdmin)
	member = args.Get(0).(typeparties.Member)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetMemberCommand(_ context.Context, userId int, partyId int) (member typeparties.MemberWithLogin, err error) {
	args := m.Called(userId, partyId)
	member = args.Get(0).(typeparties.MemberWithLogin)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetMembersCommand(_ context.Context, partyId int) (members []typeparties.MemberWithLogin, err error) {
	args := m.Called(partyId)
	members = args.Get(0).([]typeparties.MemberWithLogin)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) ChangeMemberAdminStatusCommand(_ context.Context, userId int, partyId int, isAdmin bool) error {
	args := m.Called(userId, partyId, isAdmin)
	return args.Error(0)
}

func (m *DatabaseMock) ChangeMemberDisplayNameCommand(_ context.Context, userId int, partyId int, displayName *string) error {
	args := m.Called(userId, partyId, displayName)
	return args.Error(0)
}

func (m *DatabaseMock) RemoveMemberCommand(_ context.Context, userId int, partyId int) error {
	args := m.Called(userId, partyId)
	return args.Error(0)
}
