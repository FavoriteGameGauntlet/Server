package dbitemsmock

import (
	"FGG-Service/src/changes/types"
	"FGG-Service/src/items/types"

	"github.com/stretchr/testify/mock"
)

type DatabaseMock struct {
	mock.Mock
}

func (m *DatabaseMock) CreateItemCommand(partyId int, name string, description string, useCount int, change typechanges.Change) (item typeitems.ItemWithChange, err error) {
	args := m.Called(partyId, name, description, useCount, change)
	item = args.Get(0).(typeitems.ItemWithChange)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetItemCommand(partyId int, itemId int) (item typeitems.ItemWithChange, err error) {
	args := m.Called(partyId, itemId)
	item = args.Get(0).(typeitems.ItemWithChange)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetActualItemsCommand(partyId int) (items []typeitems.Item, err error) {
	args := m.Called(partyId)
	items = args.Get(0).([]typeitems.Item)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetRemovedItemsCommand(partyId int) (items []typeitems.Item, err error) {
	args := m.Called(partyId)
	items = args.Get(0).([]typeitems.Item)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) RemoveItemCommand(partyId int, itemId int) error {
	args := m.Called(partyId, itemId)
	return args.Error(0)
}

func (m *DatabaseMock) CreateUserItemCommand(userId int, partyId int, itemId int, sourceEventId int) (userItem typeitems.UserItem, err error) {
	args := m.Called(userId, partyId, itemId, sourceEventId)
	userItem = args.Get(0).(typeitems.UserItem)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetUserItemCommand(userId int, partyId int, itemId int) (userItem typeitems.UserItem, err error) {
	args := m.Called(userId, partyId, itemId)
	userItem = args.Get(0).(typeitems.UserItem)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetUserItemsCommand(userId int, partyId int) (userItems []typeitems.UserItem, err error) {
	args := m.Called(userId, partyId)
	userItems = args.Get(0).([]typeitems.UserItem)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) ChangeUserItemUsesLeftCommand(userId int, partyId int, itemId int, usesLeft int) error {
	args := m.Called(userId, partyId, itemId, usesLeft)
	return args.Error(0)
}

func (m *DatabaseMock) CreateItemHistoryCommand(userId int, partyId int, itemId int, usesLeft int, sourceEventId int) (entry typeitems.ItemHistoryEntry, err error) {
	args := m.Called(userId, partyId, itemId, usesLeft, sourceEventId)
	entry = args.Get(0).(typeitems.ItemHistoryEntry)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetItemHistoryCommand(userId int, partyId int) (history []typeitems.ItemHistory, err error) {
	args := m.Called(userId, partyId)
	history = args.Get(0).([]typeitems.ItemHistory)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) DeleteUserItemCommand(userId int, partyId int, itemId int, sourceEventId int) error {
	args := m.Called(userId, partyId, itemId, sourceEventId)
	return args.Error(0)
}
