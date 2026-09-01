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
