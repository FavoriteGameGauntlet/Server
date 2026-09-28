package srvtimers_test

import (
	"FGG-Service/src/changes/types"
	"FGG-Service/src/timers/service"
	"FGG-Service/src/timers/types"
	"FGG-Service/tests/changes/srvmock"
	"FGG-Service/tests/games/mock"
	"FGG-Service/tests/timers/mock/dbtimers"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func endedTimerOf(userId int, gameId int) typetimers.EndedTimer {
	return typetimers.EndedTimer{
		Id:        userId * 10,
		UserId:    userId,
		PartyId:   1,
		GameId:    gameId,
		State:     typetimers.TimerStateRunning,
		Duration:  2 * time.Hour,
		TimeSpent: 2 * time.Hour,
	}
}

// A completed timer is recorded as a timer event, and that event is the source of both the time it
// adds to its game and the reward it grants to the timer's user.
func TestSrvTimers_StopAllCompletedTimers_GrantsRewardFromTimerEvent(test *testing.T) {
	// Arrange
	timerDb := new(dbtimermock.DatabaseMock)
	gamesDb := new(dbgamesmock.DatabaseMock)
	changes := new(srvchangesmock.ServiceMock)

	pointTypeId := 4
	itemId := 5
	rewardId := 3
	userId := 2
	historyId := 11

	timerDb.On("GetCompletedTimerUsersCommand").Return([]typetimers.EndedTimer{endedTimerOf(userId, 7)}, nil)
	timerDb.On("GetTimerRewardCommand", 1).Return(typetimers.TimerReward{
		Id: rewardId,
		Change: typechanges.Change{Entries: []typechanges.ChangeEntry{
			{Amount: 1, PointTypeId: &pointTypeId},
			{Amount: 1, ItemId: &itemId},
		}},
	}, nil)
	timerDb.On("CreateTimerHistoryCommand", userId, 1, &rewardId, userId).Return(typetimers.TimerHistoryEntry{Id: historyId}, nil)
	gamesDb.On("ChangeGameTimeSpentCommand", userId, 1, 7, 2*time.Hour, userId, &historyId).Return(nil)
	changes.On("ApplyChangeEntries", 1, []typechanges.ChangeEntry{
		{Amount: 1, PointTypeId: &pointTypeId, UserId: &userId},
		{Amount: 1, ItemId: &itemId, UserId: &userId},
	}, userId, historyId).Return(nil)

	sut := srvtimers.Service{Database: timerDb, GamesDatabase: gamesDb, ChangesService: changes}

	// Act
	err := sut.StopAllCompletedTimers()

	// Assert
	require.NoError(test, err)
	timerDb.AssertExpectations(test)
	gamesDb.AssertExpectations(test)
	changes.AssertExpectations(test)
}

// Without a reward set, a completed timer still gets its event and adds its time to its game.
func TestSrvTimers_StopAllCompletedTimers_WithoutReward_StillAddsGameTime(test *testing.T) {
	// Arrange
	timerDb := new(dbtimermock.DatabaseMock)
	gamesDb := new(dbgamesmock.DatabaseMock)
	changes := new(srvchangesmock.ServiceMock)

	historyId := 11

	timerDb.On("GetCompletedTimerUsersCommand").Return([]typetimers.EndedTimer{endedTimerOf(2, 7)}, nil)
	timerDb.On("GetTimerRewardCommand", 1).Return(typetimers.TimerReward{}, sql.ErrNoRows)
	timerDb.On("CreateTimerHistoryCommand", 2, 1, (*int)(nil), 2).Return(typetimers.TimerHistoryEntry{Id: historyId}, nil)
	gamesDb.On("ChangeGameTimeSpentCommand", 2, 1, 7, 2*time.Hour, 2, &historyId).Return(nil)
	changes.On("ApplyChangeEntries", 1, []typechanges.ChangeEntry(nil), 2, historyId).Return(nil)

	sut := srvtimers.Service{Database: timerDb, GamesDatabase: gamesDb, ChangesService: changes}

	// Act
	err := sut.StopAllCompletedTimers()

	// Assert
	require.NoError(test, err)
	timerDb.AssertExpectations(test)
	gamesDb.AssertExpectations(test)
}

// A timer that fails to complete does not keep the timers after it from being rewarded.
func TestSrvTimers_StopAllCompletedTimers_FailedTimer_DoesNotStopOthers(test *testing.T) {
	// Arrange
	timerDb := new(dbtimermock.DatabaseMock)
	gamesDb := new(dbgamesmock.DatabaseMock)
	changes := new(srvchangesmock.ServiceMock)

	historyId := 12

	timerDb.On("GetCompletedTimerUsersCommand").Return([]typetimers.EndedTimer{endedTimerOf(2, 7), endedTimerOf(3, 8)}, nil)
	timerDb.On("GetTimerRewardCommand", 1).Return(typetimers.TimerReward{}, sql.ErrNoRows)
	timerDb.On("CreateTimerHistoryCommand", 2, 1, (*int)(nil), 2).Return(typetimers.TimerHistoryEntry{}, dbError)
	timerDb.On("CreateTimerHistoryCommand", 3, 1, (*int)(nil), 3).Return(typetimers.TimerHistoryEntry{Id: historyId}, nil)
	gamesDb.On("ChangeGameTimeSpentCommand", 3, 1, 8, 2*time.Hour, 3, &historyId).Return(nil)
	changes.On("ApplyChangeEntries", 1, []typechanges.ChangeEntry(nil), 3, historyId).Return(nil)

	sut := srvtimers.Service{Database: timerDb, GamesDatabase: gamesDb, ChangesService: changes}

	// Act
	err := sut.StopAllCompletedTimers()

	// Assert
	require.NoError(test, err)
	gamesDb.AssertExpectations(test)
	changes.AssertExpectations(test)
}

// A failing lookup of completed timers is reported, since nothing was ended.
func TestSrvTimers_StopAllCompletedTimers_LookupFails_ReturnsError(test *testing.T) {
	// Arrange
	timerDb := new(dbtimermock.DatabaseMock)
	timerDb.On("GetCompletedTimerUsersCommand").Return([]typetimers.EndedTimer(nil), dbError)

	sut := srvtimers.Service{Database: timerDb}

	// Act
	err := sut.StopAllCompletedTimers()

	// Assert
	require.ErrorIs(test, err, dbError)
}
