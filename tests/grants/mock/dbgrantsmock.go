package dbgrantsmock

import (
	typegrants "FGG-Service/src/grants/types"

	"github.com/stretchr/testify/mock"
)

type DatabaseMock struct {
	mock.Mock
}

func (m *DatabaseMock) CreateManualHistoryCommand(userId int, partyId int, changeId int, actorUserId int, sourceEventId *int) (entry typegrants.ManualHistoryEntry, err error) {
	args := m.Called(userId, partyId, changeId, actorUserId, sourceEventId)
	entry = args.Get(0).(typegrants.ManualHistoryEntry)
	err = args.Error(1)
	return
}
