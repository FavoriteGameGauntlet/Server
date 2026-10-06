package srvtimersmock

import (
	"FGG-Service/src/timers/typetimers"
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type ServiceMock struct {
	mock.Mock
}

func (m *ServiceMock) GetCurrentTimerTimeSpent(_ context.Context, userId int, partyId int) (time.Duration, error) {
	args := m.Called(userId, partyId)
	return args.Get(0).(time.Duration), args.Error(1)
}

func (m *ServiceMock) ForceStopCurrentTimer(_ context.Context, userId int, partyId int) (typetimers.Timer, error) {
	args := m.Called(userId, partyId)
	return args.Get(0).(typetimers.Timer), args.Error(1)
}
