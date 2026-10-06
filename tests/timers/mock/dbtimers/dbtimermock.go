package dbtimermock

import (
	"FGG-Service/src/changes/types"
	"FGG-Service/src/timers/types"
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type DatabaseMock struct {
	mock.Mock
}

func (m *DatabaseMock) GetCurrentTimerCommand(_ context.Context, userId int, partyId int) (timer typetimers.CurrentTimer, err error) {
	args := m.Called(userId, partyId)
	timer = args.Get(0).(typetimers.CurrentTimer)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) CreateCurrentTimerCommand(_ context.Context, userId int, partyId int, gameId int, duration time.Duration) (timer typetimers.CreatedTimer, err error) {
	args := m.Called(userId, partyId, gameId, duration)
	timer = args.Get(0).(typetimers.CreatedTimer)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) ActTimerCommand(_ context.Context, timerId int, timerState typetimers.TimerStateType, timeSpent time.Duration) error {
	args := m.Called(timerId, timerState, timeSpent)
	return args.Error(0)
}

func (m *DatabaseMock) DeleteCurrentTimerCommand(_ context.Context, userId int, partyId int) (timer typetimers.EndedTimer, err error) {
	args := m.Called(userId, partyId)
	timer = args.Get(0).(typetimers.EndedTimer)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetCompletedTimerUsersCommand(_ context.Context) (timers []typetimers.EndedTimer, err error) {
	args := m.Called()
	timers = args.Get(0).([]typetimers.EndedTimer)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) SetTimerRewardCommand(_ context.Context, partyId int, change typechanges.Change) (timerRewardId int, err error) {
	args := m.Called(partyId, change)
	timerRewardId = args.Int(0)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) RemoveTimerRewardCommand(_ context.Context, partyId int) error {
	args := m.Called(partyId)
	return args.Error(0)
}

func (m *DatabaseMock) GetTimerRewardCommand(_ context.Context, partyId int) (reward typetimers.TimerReward, err error) {
	args := m.Called(partyId)
	reward = args.Get(0).(typetimers.TimerReward)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetTimerRewardEntriesCommand(_ context.Context, partyId int) (entries []typechanges.ChangeEntryInput, err error) {
	args := m.Called(partyId)
	entries = args.Get(0).([]typechanges.ChangeEntryInput)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) CreateTimerHistoryCommand(_ context.Context, userId int, partyId int, timerRewardId *int, actorUserId int) (entry typetimers.TimerHistoryEntry, err error) {
	args := m.Called(userId, partyId, timerRewardId, actorUserId)
	entry = args.Get(0).(typetimers.TimerHistoryEntry)
	err = args.Error(1)
	return
}
