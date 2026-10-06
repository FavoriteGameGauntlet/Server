package srvchangesmock

import (
	"FGG-Service/src/changes/types"
	"context"

	"github.com/stretchr/testify/mock"
)

type ServiceMock struct {
	mock.Mock
}

func (m *ServiceMock) ApplyChangeEntries(_ context.Context, partyId int, entries []typechanges.ChangeEntry, actorUserId int, sourceEventId int) error {
	args := m.Called(partyId, entries, actorUserId, sourceEventId)
	return args.Error(0)
}

func (m *ServiceMock) ResolveChangeEntries(_ context.Context, partyId int, inputs []typechanges.ChangeEntryInput) (entries []typechanges.ChangeEntry, err error) {
	args := m.Called(partyId, inputs)
	entries = args.Get(0).([]typechanges.ChangeEntry)
	err = args.Error(1)
	return
}
