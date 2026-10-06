package dbitemsmock

import (
	"FGG-Service/src/changes/typechanges"
	"FGG-Service/src/items/typeitems"
	"context"

	"github.com/stretchr/testify/mock"
)

type DatabaseMock struct {
	mock.Mock
}

func (m *DatabaseMock) CreateItemCommand(_ context.Context, partyId int, name string, description string, useCount int, change typechanges.Change) (item typeitems.ItemWithChange, err error) {
	args := m.Called(partyId, name, description, useCount, change)
	item = args.Get(0).(typeitems.ItemWithChange)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetItemCommand(_ context.Context, partyId int, itemId int) (item typeitems.ItemWithChange, err error) {
	args := m.Called(partyId, itemId)
	item = args.Get(0).(typeitems.ItemWithChange)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetActualItemsCommand(_ context.Context, partyId int) (items []typeitems.Item, err error) {
	args := m.Called(partyId)
	items = args.Get(0).([]typeitems.Item)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetRemovedItemsCommand(_ context.Context, partyId int) (items []typeitems.Item, err error) {
	args := m.Called(partyId)
	items = args.Get(0).([]typeitems.Item)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) RemoveItemCommand(_ context.Context, partyId int, itemId int) error {
	args := m.Called(partyId, itemId)
	return args.Error(0)
}

func (m *DatabaseMock) CreateUserItemCommand(_ context.Context, userId int, partyId int, itemId int, actorUserId int, sourceEventId int) (userItem typeitems.UserItem, err error) {
	args := m.Called(userId, partyId, itemId, actorUserId, sourceEventId)
	userItem = args.Get(0).(typeitems.UserItem)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetUserItemCommand(_ context.Context, userId int, partyId int, itemId int) (userItem typeitems.UserItem, err error) {
	args := m.Called(userId, partyId, itemId)
	userItem = args.Get(0).(typeitems.UserItem)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetUserItemsCommand(_ context.Context, userId int, partyId int) (userItems []typeitems.UserItemDetail, err error) {
	args := m.Called(userId, partyId)
	userItems = args.Get(0).([]typeitems.UserItemDetail)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) ChangeUserItemUsesLeftCommand(_ context.Context, userId int, partyId int, itemId int, usesLeft int, actorUserId int, sourceEventId *int) (historyEventId int, err error) {
	args := m.Called(userId, partyId, itemId, usesLeft, actorUserId, sourceEventId)
	historyEventId = args.Int(0)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetItemHistoryCommand(_ context.Context, userId int, partyId int) (history []typeitems.ItemHistory, err error) {
	args := m.Called(userId, partyId)
	history = args.Get(0).([]typeitems.ItemHistory)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) DeleteUserItemCommand(_ context.Context, userId int, partyId int, itemId int, actorUserId int, sourceEventId *int) error {
	args := m.Called(userId, partyId, itemId, actorUserId, sourceEventId)
	return args.Error(0)
}
