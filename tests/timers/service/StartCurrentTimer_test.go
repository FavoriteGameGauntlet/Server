package srvtimers_test

import (
	"FGG-Service/src/common"
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

type StartCurrentTimerTestCase struct {
	Name            string
	UserId          int
	SetupMocks      func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *dbwheeleffectsmock.DatabaseMock)
	ExpectedTimer   *typetimers.Timer
	ExpectedErrorAs interface{}
	ExpectedErrorIs error
}

var createdCurrentTimer = typetimers.CurrentTimer{
	Id:             1,
	GameId:         1,
	State:          typetimers.TimerStateCreated,
	Duration:       2 * time.Hour,
	TimeSpent:      0,
	LastActionDate: time.Now(),
}

var createdTimer = typetimers.Timer{
	Id:             1,
	Duration:       2 * time.Hour,
	RemainingTime:  2 * time.Hour,
	State:          typetimers.TimerStateCreated,
	LastActionDate: createdCurrentTimer.LastActionDate,
}

var runningCurrentTimerResult = typetimers.CurrentTimer{
	Id:             1,
	GameId:         1,
	State:          typetimers.TimerStateRunning,
	Duration:       2 * time.Hour,
	TimeSpent:      0,
	LastActionDate: time.Now(),
}

var runningTimerResult = typetimers.Timer{
	Id:             1,
	Duration:       2 * time.Hour,
	RemainingTime:  2 * time.Hour,
	State:          typetimers.TimerStateRunning,
	LastActionDate: runningCurrentTimerResult.LastActionDate,
}

var StartCurrentTimerTestCases = []StartCurrentTimerTestCase{
	{
		// GetCurrentTimerCommand returns sql.ErrNoRows. The CurrentTimerNotFoundError will return.
		Name:   "NotFound",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *dbwheeleffectsmock.DatabaseMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)
			wheelDb := new(dbwheeleffectsmock.DatabaseMock)

			timerDb.On("GetCurrentTimerCommand", 1, 1).Return(typetimers.CurrentTimer{}, sql.ErrNoRows)

			return timerDb, gamesDb, wheelDb
		},
		ExpectedErrorAs: new(common.NotFoundError),
	},
	{
		// GetCurrentTimerCommand returns a database error. The error will return.
		Name:   "DatabaseError",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *dbwheeleffectsmock.DatabaseMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)
			wheelDb := new(dbwheeleffectsmock.DatabaseMock)

			timerDb.On("GetCurrentTimerCommand", 1, 1).Return(typetimers.CurrentTimer{}, dbError)

			return timerDb, gamesDb, wheelDb
		},
		ExpectedErrorIs: dbError,
	},
	{
		// Timer is already Running. The IncorrectStateConflictError will return.
		Name:   "IncorrectState_Running",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *dbwheeleffectsmock.DatabaseMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)
			wheelDb := new(dbwheeleffectsmock.DatabaseMock)

			timerDb.On("GetCurrentTimerCommand", 1, 1).Return(
				typetimers.CurrentTimer{State: typetimers.TimerStateRunning}, nil)

			return timerDb, gamesDb, wheelDb
		},
		ExpectedErrorAs: new(common.ConflictError),
	},
	{
		// Timer is Finished. The IncorrectStateConflictError will return.
		Name:   "IncorrectState_Finished",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *dbwheeleffectsmock.DatabaseMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)
			wheelDb := new(dbwheeleffectsmock.DatabaseMock)

			timerDb.On("GetCurrentTimerCommand", 1, 1).Return(
				typetimers.CurrentTimer{State: typetimers.TimerStateFinished}, nil)

			return timerDb, gamesDb, wheelDb
		},
		ExpectedErrorAs: new(common.ConflictError),
	},
	{
		// ActTimerCommand returns a database error. The error will return.
		Name:   "ActTimerDatabaseError",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *dbwheeleffectsmock.DatabaseMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)
			wheelDb := new(dbwheeleffectsmock.DatabaseMock)

			timerDb.On("GetCurrentTimerCommand", 1, 1).Return(createdCurrentTimer, nil)
			timerDb.On("ActTimerCommand", 1, typetimers.TimerStateRunning, createdCurrentTimer.TimeSpent).
				Return(dbError)

			return timerDb, gamesDb, wheelDb
		},
		ExpectedErrorIs: dbError,
	},
	{
		// Timer is Created. ActTimerCommand is called with zero time spent (not the full duration).
		// The updated running timer will return.
		Name:   "Success_Created",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *dbwheeleffectsmock.DatabaseMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)
			wheelDb := new(dbwheeleffectsmock.DatabaseMock)

			timerDb.On("GetCurrentTimerCommand", 1, 1).Once().Return(createdCurrentTimer, nil)
			timerDb.On("ActTimerCommand", 1, typetimers.TimerStateRunning, time.Duration(0)).
				Return(nil)
			timerDb.On("GetCurrentTimerCommand", 1, 1).Once().Return(runningCurrentTimerResult, nil)

			return timerDb, gamesDb, wheelDb
		},
		ExpectedTimer: &runningTimerResult,
	},
	{
		// Timer is Paused. ActTimerCommand is called with the time already spent.
		// The updated running timer will return.
		Name:   "Success_Paused",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *dbwheeleffectsmock.DatabaseMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)
			wheelDb := new(dbwheeleffectsmock.DatabaseMock)

			pausedCurrentTimerLocal := typetimers.CurrentTimer{
				Id:        1,
				Duration:  2 * time.Hour,
				TimeSpent: 75 * time.Minute,
				State:     typetimers.TimerStatePaused,
			}
			timerDb.On("GetCurrentTimerCommand", 1, 1).Once().Return(pausedCurrentTimerLocal, nil)
			timerDb.On("ActTimerCommand", 1, typetimers.TimerStateRunning, 75*time.Minute).
				Return(nil)
			timerDb.On("GetCurrentTimerCommand", 1, 1).Once().Return(runningCurrentTimerResult, nil)

			return timerDb, gamesDb, wheelDb
		},
		ExpectedTimer: &runningTimerResult,
	},
}

func TestSrvTimers_StartCurrentTimer(test *testing.T) {
	for _, testCase := range StartCurrentTimerTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			timerDb, gamesDb, wheelDb := testCase.SetupMocks()
			sut := srvtimers.Service{
				Database:             timerDb,
				GamesDatabase:        gamesDb,
				WheelEffectsDatabase: wheelDb,
			}

			// Act
			timer, err := sut.StartCurrentTimer(testCase.UserId, 1)

			// Assert
			if testCase.ExpectedErrorAs != nil {
				require.Error(test, err)
				require.ErrorAs(test, err, &testCase.ExpectedErrorAs)
			}

			if testCase.ExpectedErrorIs != nil {
				require.ErrorIs(test, err, testCase.ExpectedErrorIs)
			}

			if testCase.ExpectedTimer != nil {
				require.NoError(test, err)
				require.Equal(test, *testCase.ExpectedTimer, timer)
			}

			timerDb.AssertExpectations(test)
			gamesDb.AssertExpectations(test)
			wheelDb.AssertExpectations(test)
		})
	}
}
