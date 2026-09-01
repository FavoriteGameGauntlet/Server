package dbwheeleffectsmock

import (
	"FGG-Service/src/wheeleffects/types"

	"github.com/stretchr/testify/mock"
)

type DatabaseMock struct {
	mock.Mock
}

func (m *DatabaseMock) GetAvailableRollsCountCommand(userId int) (count int, err error) {
	args := m.Called(userId)
	count = args.Get(0).(int)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetAvailableWheelRowsCommand(userId int, partyId int, collectionId int) (rows []typewheeleffects.WheelRow, err error) {
	args := m.Called(userId, partyId, collectionId)
	rows = args.Get(0).([]typewheeleffects.WheelRow)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetEffectHistoryCommand(userId int, partyId int) (history []typewheeleffects.WheelRowHistory, err error) {
	args := m.Called(userId, partyId)
	history = args.Get(0).([]typewheeleffects.WheelRowHistory)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetEffectHistoryByEffectNameCommand(userId int, partyId int, wheelRowName string) (history typewheeleffects.WheelRowHistory, err error) {
	args := m.Called(userId, partyId, wheelRowName)
	history = args.Get(0).(typewheeleffects.WheelRowHistory)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) ClearLastWheelEffectsCommand(userId int, partyId int) error {
	args := m.Called(userId, partyId)
	return args.Error(0)
}

func (m *DatabaseMock) AddLastRolledWheelEffectsCommand(userId int, partyId int, rows []typewheeleffects.RolledWheelRowInput) (created []typewheeleffects.CreatedLastWheelRow, err error) {
	args := m.Called(userId, partyId, rows)
	created = args.Get(0).([]typewheeleffects.CreatedLastWheelRow)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetLastRolledWheelEffectsCommand(userId int, partyId int) (rows []typewheeleffects.LastWheelRow, err error) {
	args := m.Called(userId, partyId)
	rows = args.Get(0).([]typewheeleffects.LastWheelRow)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) AddWheelEffectHistoryCommand(userId int, partyId int, wheelRowId int, sourceEventId *int) (created typewheeleffects.CreatedWheelRowHistory, err error) {
	args := m.Called(userId, partyId, wheelRowId, sourceEventId)
	created = args.Get(0).(typewheeleffects.CreatedWheelRowHistory)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) CreateWheelCollectionCommand(partyId int, name string, shouldCheckHistory bool) (collection typewheeleffects.WheelCollection, err error) {
	args := m.Called(partyId, name, shouldCheckHistory)
	collection = args.Get(0).(typewheeleffects.WheelCollection)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetWheelCollectionsCommand(partyId int) (collections []typewheeleffects.WheelCollection, err error) {
	args := m.Called(partyId)
	collections = args.Get(0).([]typewheeleffects.WheelCollection)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) CreateWheelRowCommand(partyId int, name string, description string, changeId int, collectionId int) (row typewheeleffects.CreatedWheelRow, err error) {
	args := m.Called(partyId, name, description, changeId, collectionId)
	row = args.Get(0).(typewheeleffects.CreatedWheelRow)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetWheelRowCommand(partyId int, wheelRowId int) (row typewheeleffects.WheelRow, err error) {
	args := m.Called(partyId, wheelRowId)
	row = args.Get(0).(typewheeleffects.WheelRow)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetWheelRowsCommand(partyId int) (rows []typewheeleffects.WheelRow, err error) {
	args := m.Called(partyId)
	rows = args.Get(0).([]typewheeleffects.WheelRow)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) DeleteWheelRowCommand(partyId int, wheelRowId int) error {
	args := m.Called(partyId, wheelRowId)
	return args.Error(0)
}
