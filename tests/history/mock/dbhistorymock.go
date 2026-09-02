package dbhistorymock

import (
	typechanges "FGG-Service/src/changes/types"
	typehistory "FGG-Service/src/history/types"

	"github.com/stretchr/testify/mock"
)

type DatabaseMock struct {
	mock.Mock
}

func (m *DatabaseMock) CreateManualHistoryCommand(partyId int, actorUserId int, entries []typechanges.ChangeEntry, sourceEventId *int) (created []typehistory.ManualHistoryEntry, err error) {
	args := m.Called(partyId, actorUserId, entries, sourceEventId)
	created = args.Get(0).([]typehistory.ManualHistoryEntry)
	err = args.Error(1)
	return
}