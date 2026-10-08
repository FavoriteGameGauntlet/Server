package dbchangesmock

import (
	"FGG-Service/src/changes/typechanges"
	"context"

	"github.com/stretchr/testify/mock"
)

type DatabaseMock struct {
	mock.Mock
}

func (m *DatabaseMock) ChangeEntryToJsonbCommand(_ context.Context, entryId *int, amount int, pointTypeId *int, itemId *int, perkId *int, effectId *int, userId *int) (entry typechanges.ChangeEntry, err error) {
	args := m.Called(entryId, amount, pointTypeId, itemId, perkId, effectId, userId)
	entry = args.Get(0).(typechanges.ChangeEntry)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) ChangeToJsonbCommand(_ context.Context, changeId *int, shouldApplyToAll bool, isManualChange bool, entries []typechanges.ChangeEntry) (change typechanges.Change, err error) {
	args := m.Called(changeId, shouldApplyToAll, isManualChange, entries)
	change = args.Get(0).(typechanges.Change)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) CreateChangeFromJsonbCommand(_ context.Context, partyId int, change typechanges.Change) (created typechanges.Change, err error) {
	args := m.Called(partyId, change)
	created = args.Get(0).(typechanges.Change)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) CreateUserChangeFromJsonbCommand(_ context.Context, partyId int, entries []typechanges.ChangeEntry) (created typechanges.UserChange, err error) {
	args := m.Called(partyId, entries)
	created = args.Get(0).(typechanges.UserChange)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetChangeEntriesJsonbCommand(_ context.Context, partyId int, changeId int) (entries []typechanges.ChangeEntry, err error) {
	args := m.Called(partyId, changeId)
	entries = args.Get(0).([]typechanges.ChangeEntry)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetNamedChangeEntriesCommand(_ context.Context, partyId int, changeId int) (entries []typechanges.NamedChangeEntry, err error) {
	args := m.Called(partyId, changeId)
	entries = args.Get(0).([]typechanges.NamedChangeEntry)
	err = args.Error(1)
	return
}
