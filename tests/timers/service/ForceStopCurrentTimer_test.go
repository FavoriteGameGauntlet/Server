package srvtimers_test

import (
	"FGG-Service/src/timers/service"
	"FGG-Service/src/timers/types"
	"FGG-Service/tests/games/mock"
	"FGG-Service/tests/timers/mock/dbtimers"
	"FGG-Service/tests/timers/mock/dbwheeleffects"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type ForceStopCurrentTimerTestCase struct {
	Name            string
	UserId          int
	SetupMocks      func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *dbwheeleffectsmock.DatabaseMock)
	ExpectNoError   bool
	ExpectedErrorIs error
}

// deletedTimer is what delete_timer returns for a timer stopped 30 minutes in.
var deletedTimer = typetimers.EndedTimer{
	Id:        1,
	UserId:    1,
	PartyId:   1,
	GameId:    7,
	State:     typetimers.TimerStateRunning,
	Duration:  2 * time.Hour,
	TimeSpent: 30 * time.Minute,
}

var ForceStopCurrentTimerTestCases = []ForceStopCurrentTimerTestCase{
	{
		// The user has no timer. Nothing is deleted or added and no error returns.
		Name:   "NotFound",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *dbwheeleffectsmock.DatabaseMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)
			wheelDb := new(dbwheeleffectsmock.DatabaseMock)

			timerDb.On("DeleteCurrentTimerCommand", 1, 1).Return(typetimers.EndedTimer{}, sql.ErrNoRows)

			return timerDb, gamesDb, wheelDb
		},
		ExpectNoError: true,
	},
	{
		// DeleteCurrentTimerCommand returns a database error. The error will return.
		Name:   "DeleteDatabaseError",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *dbwheeleffectsmock.DatabaseMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)
			wheelDb := new(dbwheeleffectsmock.DatabaseMock)

			timerDb.On("DeleteCurrentTimerCommand", 1, 1).Return(typetimers.EndedTimer{}, dbError)

			return timerDb, gamesDb, wheelDb
		},
		ExpectedErrorIs: dbError,
	},
	{
		// The timer never ran. It is deleted and the game time is left alone.
		Name:   "Success_NoTimeSpent",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *dbwheeleffectsmock.DatabaseMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)
			wheelDb := new(dbwheeleffectsmock.DatabaseMock)

			unstartedTimer := deletedTimer
			unstartedTimer.State = typetimers.TimerStateCreated
			unstartedTimer.TimeSpent = 0
			timerDb.On("DeleteCurrentTimerCommand", 1, 1).Return(unstartedTimer, nil)

			return timerDb, gamesDb, wheelDb
		},
		ExpectNoError: true,
	},
	{
		// ChangeGameTimeSpentCommand returns a database error. The error will return.
		Name:   "ChangeGameTimeDatabaseError",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *dbwheeleffectsmock.DatabaseMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)
			wheelDb := new(dbwheeleffectsmock.DatabaseMock)

			timerDb.On("DeleteCurrentTimerCommand", 1, 1).Return(deletedTimer, nil)
			gamesDb.On("ChangeGameTimeSpentCommand", 1, 1, 7, 30*time.Minute, 1, (*int)(nil)).Return(dbError)

			return timerDb, gamesDb, wheelDb
		},
		ExpectedErrorIs: dbError,
	},
	{
		// The timer ran for 30 minutes. It is deleted and those 30 minutes are added to its game.
		Name:   "Success_TimeSpentAddedToGame",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *dbwheeleffectsmock.DatabaseMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)
			wheelDb := new(dbwheeleffectsmock.DatabaseMock)

			timerDb.On("DeleteCurrentTimerCommand", 1, 1).Return(deletedTimer, nil)
			gamesDb.On("ChangeGameTimeSpentCommand", 1, 1, 7, 30*time.Minute, 1, (*int)(nil)).Return(nil)

			return timerDb, gamesDb, wheelDb
		},
		ExpectNoError: true,
	},
}

func TestSrvTimers_ForceStopCurrentTimer(test *testing.T) {
	for _, testCase := range ForceStopCurrentTimerTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			timerDb, gamesDb, wheelDb := testCase.SetupMocks()
			sut := srvtimers.Service{
				Database:             timerDb,
				GamesDatabase:        gamesDb,
				WheelEffectsDatabase: wheelDb,
			}

			// Act
			_, err := sut.ForceStopCurrentTimer(testCase.UserId)

			// Assert
			if testCase.ExpectedErrorIs != nil {
				require.ErrorIs(test, err, testCase.ExpectedErrorIs)
			}

			if testCase.ExpectNoError {
				require.NoError(test, err)
			}

			timerDb.AssertExpectations(test)
			gamesDb.AssertExpectations(test)
			wheelDb.AssertExpectations(test)
		})
	}
}
