package dbpartiesmock

import (
	typeparties "FGG-Service/src/parties/types"

	"github.com/stretchr/testify/mock"
)

type DatabaseMock struct {
	mock.Mock
}

func (m *DatabaseMock) CreatePartyCommand(name string) (party typeparties.Party, err error) {
	args := m.Called(name)
	party = args.Get(0).(typeparties.Party)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetPartyCommand(partyId int) (party typeparties.Party, err error) {
	args := m.Called(partyId)
	party = args.Get(0).(typeparties.Party)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetPartiesCommand() (parties []typeparties.Party, err error) {
	args := m.Called()
	parties = args.Get(0).([]typeparties.Party)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) ChangePartyNameCommand(partyId int, name string) error {
	args := m.Called(partyId, name)
	return args.Error(0)
}

func (m *DatabaseMock) DeletePartyCommand(partyId int) error {
	args := m.Called(partyId)
	return args.Error(0)
}

func (m *DatabaseMock) CreateMemberCommand(userId int, partyId int, displayName string, isAdmin bool) (member typeparties.Member, err error) {
	args := m.Called(userId, partyId, displayName, isAdmin)
	member = args.Get(0).(typeparties.Member)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetMemberCommand(userId int, partyId int) (member typeparties.MemberWithLogin, err error) {
	args := m.Called(userId, partyId)
	member = args.Get(0).(typeparties.MemberWithLogin)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetMembersCommand(partyId int) (members []typeparties.MemberWithLogin, err error) {
	args := m.Called(partyId)
	members = args.Get(0).([]typeparties.MemberWithLogin)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) ChangeMemberAdminStatusCommand(userId int, partyId int, isAdmin bool) error {
	args := m.Called(userId, partyId, isAdmin)
	return args.Error(0)
}

func (m *DatabaseMock) ChangeMemberDisplayNameCommand(userId int, partyId int, displayName string) error {
	args := m.Called(userId, partyId, displayName)
	return args.Error(0)
}

func (m *DatabaseMock) RemoveMemberCommand(userId int, partyId int) error {
	args := m.Called(userId, partyId)
	return args.Error(0)
}