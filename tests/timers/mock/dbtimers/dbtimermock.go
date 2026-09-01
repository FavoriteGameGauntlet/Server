package dbtimermock

import (
	"FGG-Service/src/timers/types"
	"time"

	"github.com/stretchr/testify/mock"
)

type DatabaseMock struct {
	mock.Mock
}

func (m *DatabaseMock) GetCurrentTimerCommand(userId int, partyId int) (timer typetimers.CurrentTimer, err error) {
	args := m.Called(userId, partyId)
	timer = args.Get(0).(typetimers.CurrentTimer)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) CreateCurrentTimerCommand(userId int, partyId int, gameId int, duration time.Duration) (timer typetimers.CreatedTimer, err error) {
	args := m.Called(userId, partyId, gameId, duration)
	timer = args.Get(0).(typetimers.CreatedTimer)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) ActTimerCommand(timerId int, timerState typetimers.TimerStateType, timeSpent time.Duration) error {
	args := m.Called(timerId, timerState, timeSpent)
	return args.Error(0)
}

func (m *DatabaseMock) GetCompletedTimerUsersCommand() (timers []typetimers.EndedTimer, err error) {
	args := m.Called()
	timers = args.Get(0).([]typetimers.EndedTimer)
	err = args.Error(1)
	return
}
