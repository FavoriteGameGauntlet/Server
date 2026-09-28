package srvtimersmock

import (
	"FGG-Service/src/timers/types"
	"time"

	"github.com/stretchr/testify/mock"
)

type ServiceMock struct {
	mock.Mock
}

func (m *ServiceMock) GetCurrentTimerTimeSpent(userId int) (time.Duration, error) {
	args := m.Called(userId)
	return args.Get(0).(time.Duration), args.Error(1)
}

func (m *ServiceMock) ForceStopCurrentTimer(userId int) (typetimers.Timer, error) {
	args := m.Called(userId)
	return args.Get(0).(typetimers.Timer), args.Error(1)
}
