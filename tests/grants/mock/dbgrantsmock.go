package dbgrantsmock

import (
	typechanges "FGG-Service/src/changes/types"
	typegrants "FGG-Service/src/grants/types"

	"github.com/stretchr/testify/mock"
)

type DatabaseMock struct {
	mock.Mock
}

func (m *DatabaseMock) CreateManualHistoryCommand(partyId int, actorUserId int, entries []typechanges.ChangeEntry, sourceEventId *int) (created []typegrants.ManualHistoryEntry, err error) {
	args := m.Called(partyId, actorUserId, entries, sourceEventId)
	created = args.Get(0).([]typegrants.ManualHistoryEntry)
	err = args.Error(1)
	return
}