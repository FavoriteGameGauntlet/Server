package srvtimers_test

import (
	"FGG-Service/src/common"
	"FGG-Service/src/timers/service"
	"FGG-Service/src/timers/types"
	"FGG-Service/tests/games/mock"
	"FGG-Service/tests/timers/mock/dbtimers"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type PauseCurrentTimerTestCase struct {
	Name            string
	UserId          int
	SetupMocks      func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock)
	ExpectedTimer   *typetimers.Timer
	ExpectedErrorAs interface{}
	ExpectedErrorIs error
}

// runningCurrentTimer is what the DB returns for a running timer; runningTimer is the derived Timer
// (RemainingTime = Duration - TimeSpent) that the service must return for it. Both share
// LastActionDate so RemainingTime-based assertions in other test files stay consistent.
var runningCurrentTimer = typetimers.CurrentTimer{
	Id:             1,
	GameId:         1,
	State:          typetimers.TimerStateRunning,
	Duration:       2 * time.Hour,
	TimeSpent:      30 * time.Minute,
	LastActionDate: time.Now(),
}

var runningTimer = typetimers.Timer{
	Id:             1,
	Duration:       2 * time.Hour,
	RemainingTime:  90 * time.Minute,
	State:          typetimers.TimerStateRunning,
	LastActionDate: runningCurrentTimer.LastActionDate,
}

var pausedCurrentTimer = typetimers.CurrentTimer{
	Id:             1,
	GameId:         1,
	State:          typetimers.TimerStatePaused,
	Duration:       2 * time.Hour,
	TimeSpent:      30 * time.Minute,
	LastActionDate: time.Now(),
}

var pausedTimerResult = typetimers.Timer{
	Id:             1,
	Duration:       2 * time.Hour,
	RemainingTime:  90 * time.Minute,
	State:          typetimers.TimerStatePaused,
	LastActionDate: pausedCurrentTimer.LastActionDate,
}

var PauseCurrentTimerTestCases = []PauseCurrentTimerTestCase{
	{
		// GetCurrentTimerCommand returns sql.ErrNoRows. The CurrentTimerNotFoundError will return.
		Name:   "NotFound",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)

			timerDb.On("GetCurrentTimerCommand", 1, 1).Return(typetimers.CurrentTimer{}, sql.ErrNoRows)

			return timerDb, gamesDb
		},
		ExpectedErrorAs: new(common.NotFoundError),
	},
	{
		// GetCurrentTimerCommand returns a database error. The error will return.
		Name:   "DatabaseError",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)

			timerDb.On("GetCurrentTimerCommand", 1, 1).Return(typetimers.CurrentTimer{}, dbError)

			return timerDb, gamesDb
		},
		ExpectedErrorIs: dbError,
	},
	{
		// Timer is in Created state. The IncorrectStateConflictError will return.
		Name:   "IncorrectState_Created",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)

			timerDb.On("GetCurrentTimerCommand", 1, 1).Return(
				typetimers.CurrentTimer{State: typetimers.TimerStateCreated}, nil)

			return timerDb, gamesDb
		},
		ExpectedErrorAs: new(common.ConflictError),
	},
	{
		// Timer is already Paused. The IncorrectStateConflictError will return.
		Name:   "IncorrectState_Paused",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)

			timerDb.On("GetCurrentTimerCommand", 1, 1).Return(
				typetimers.CurrentTimer{State: typetimers.TimerStatePaused}, nil)

			return timerDb, gamesDb
		},
		ExpectedErrorAs: new(common.ConflictError),
	},
	{
		// Timer is Finished. The IncorrectStateConflictError will return.
		Name:   "IncorrectState_Finished",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)

			timerDb.On("GetCurrentTimerCommand", 1, 1).Return(
				typetimers.CurrentTimer{State: typetimers.TimerStateFinished}, nil)

			return timerDb, gamesDb
		},
		ExpectedErrorAs: new(common.ConflictError),
	},
	{
		// ActTimerCommand returns a database error. The error will return.
		Name:   "ActTimerDatabaseError",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)

			timerDb.On("GetCurrentTimerCommand", 1, 1).Return(runningCurrentTimer, nil)
			timerDb.On("ActTimerCommand", 1, typetimers.TimerStatePaused, mock.AnythingOfType("time.Duration")).
				Return(dbError)

			return timerDb, gamesDb
		},
		ExpectedErrorIs: dbError,
	},
	{
		// Timer is Running. ActTimerCommand is called with the time spent reported by the DB
		// (which already includes the elapsed running time). The updated timer will return.
		Name:   "Success_Running",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)

			timerDb.On("GetCurrentTimerCommand", 1, 1).Once().Return(runningCurrentTimer, nil)
			timerDb.On("ActTimerCommand", 1, typetimers.TimerStatePaused, runningCurrentTimer.TimeSpent).
				Return(nil)
			timerDb.On("GetCurrentTimerCommand", 1, 1).Once().Return(pausedCurrentTimer, nil)

			return timerDb, gamesDb
		},
		ExpectedTimer: &pausedTimerResult,
	},
}

func TestSrvTimers_PauseCurrentTimer(test *testing.T) {
	for _, testCase := range PauseCurrentTimerTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			timerDb, gamesDb := testCase.SetupMocks()
			sut := srvtimers.Service{
				Database:      timerDb,
				GamesDatabase: gamesDb,
			}

			// Act
			timer, err := sut.PauseCurrentTimer(testCase.UserId, 1)

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
		})
	}
}
