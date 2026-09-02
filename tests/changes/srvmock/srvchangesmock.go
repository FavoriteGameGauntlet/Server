package srvchangesmock

import (
	"FGG-Service/src/changes/types"

	"github.com/stretchr/testify/mock"
)

type ServiceMock struct {
	mock.Mock
}

func (m *ServiceMock) ApplyChangeEntries(partyId int, entries []typechanges.ChangeEntry, sourceEventId int) error {
	args := m.Called(partyId, entries, sourceEventId)
	return args.Error(0)
}

func (m *ServiceMock) ResolveChangeEntries(partyId int, inputs []typechanges.ChangeEntryInput) (entries []typechanges.ChangeEntry, err error) {
	args := m.Called(partyId, inputs)
	entries = args.Get(0).([]typechanges.ChangeEntry)
	err = args.Error(1)
	return
}