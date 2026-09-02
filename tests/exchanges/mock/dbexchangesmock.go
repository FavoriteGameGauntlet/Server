package dbexchangesmock

import (
	typechanges "FGG-Service/src/changes/types"
	typeexchanges "FGG-Service/src/exchanges/types"

	"github.com/stretchr/testify/mock"
)

type DatabaseMock struct {
	mock.Mock
}

func (m *DatabaseMock) CreateExchangeCommand(partyId int, name string, description string, sourceChange typechanges.Change, targetChange typechanges.Change) (exchange typeexchanges.ExchangeWithChanges, err error) {
	args := m.Called(partyId, name, description, sourceChange, targetChange)
	exchange = args.Get(0).(typeexchanges.ExchangeWithChanges)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetExchangeWithEntriesCommand(partyId int, exchangeId int) (exchange typeexchanges.ExchangeWithChanges, err error) {
	args := m.Called(partyId, exchangeId)
	exchange = args.Get(0).(typeexchanges.ExchangeWithChanges)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetActualExchangesCommand(partyId int) (exchanges []typeexchanges.Exchange, err error) {
	args := m.Called(partyId)
	exchanges = args.Get(0).([]typeexchanges.Exchange)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetRemovedExchangesCommand(partyId int) (exchanges []typeexchanges.Exchange, err error) {
	args := m.Called(partyId)
	exchanges = args.Get(0).([]typeexchanges.Exchange)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) RemoveExchangeCommand(partyId int, exchangeId int) error {
	args := m.Called(partyId, exchangeId)
	return args.Error(0)
}

func (m *DatabaseMock) CreateExchangeHistoryCommand(userId int, partyId int, exchangeId int, sourceEventId *int) (entry typeexchanges.ExchangeHistoryEntry, err error) {
	args := m.Called(userId, partyId, exchangeId, sourceEventId)
	entry = args.Get(0).(typeexchanges.ExchangeHistoryEntry)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetExchangeHistoryCommand(userId int, partyId int) (history []typeexchanges.ExchangeHistory, err error) {
	args := m.Called(userId, partyId)
	history = args.Get(0).([]typeexchanges.ExchangeHistory)
	err = args.Error(1)
	return
}