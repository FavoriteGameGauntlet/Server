package dbperksmock

import (
	"FGG-Service/src/perks/types"
	"context"

	"github.com/stretchr/testify/mock"
)

type DatabaseMock struct {
	mock.Mock
}

func (m *DatabaseMock) CreatePerkCommand(_ context.Context, partyId int, name string, description string, effectId int) (perk typeperks.PerkWithRemoved, err error) {
	args := m.Called(partyId, name, description, effectId)
	perk = args.Get(0).(typeperks.PerkWithRemoved)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetPerkCommand(_ context.Context, partyId int, perkId int) (perk typeperks.Perk, err error) {
	args := m.Called(partyId, perkId)
	perk = args.Get(0).(typeperks.Perk)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetActualPerksCommand(_ context.Context, partyId int) (perks []typeperks.Perk, err error) {
	args := m.Called(partyId)
	perks = args.Get(0).([]typeperks.Perk)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetRemovedPerksCommand(_ context.Context, partyId int) (perks []typeperks.Perk, err error) {
	args := m.Called(partyId)
	perks = args.Get(0).([]typeperks.Perk)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) RemovePerkCommand(_ context.Context, partyId int, perkId int) error {
	args := m.Called(partyId, perkId)
	return args.Error(0)
}

func (m *DatabaseMock) CreateUserPerkCommand(_ context.Context, userId int, partyId int, perkId int, actorUserId int, sourceEventId int) (userPerk typeperks.UserPerk, err error) {
	args := m.Called(userId, partyId, perkId, actorUserId, sourceEventId)
	userPerk = args.Get(0).(typeperks.UserPerk)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetUserPerkCommand(_ context.Context, userId int, partyId int, perkId int) (userPerk typeperks.UserPerk, err error) {
	args := m.Called(userId, partyId, perkId)
	userPerk = args.Get(0).(typeperks.UserPerk)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetUserPerksCommand(_ context.Context, userId int, partyId int) (userPerks []typeperks.UserPerk, err error) {
	args := m.Called(userId, partyId)
	userPerks = args.Get(0).([]typeperks.UserPerk)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) DeleteUserPerkCommand(_ context.Context, userId int, partyId int, perkId int, actorUserId int, sourceEventId *int) error {
	args := m.Called(userId, partyId, perkId, actorUserId, sourceEventId)
	return args.Error(0)
}

func (m *DatabaseMock) GetPerkHistoryCommand(_ context.Context, userId int, partyId int) (history []typeperks.PerkHistory, err error) {
	args := m.Called(userId, partyId)
	history = args.Get(0).([]typeperks.PerkHistory)
	err = args.Error(1)
	return
}
