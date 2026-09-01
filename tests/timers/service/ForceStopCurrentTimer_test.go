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

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type ForceStopCurrentTimerTestCase struct {
	Name            string
	UserId          int
	SetupMocks      func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *dbwheeleffectsmock.DatabaseMock)
	ExpectNoError   bool
	ExpectedErrorAs interface{}
	ExpectedErrorIs error
}

var finishedCurrentTimer = typetimers.CurrentTimer{
	Id:        1,
	Duration:  runningCurrentTimer.Duration,
	TimeSpent: runningCurrentTimer.TimeSpent,
	State:     typetimers.TimerStateFinished,
}

var ForceStopCurrentTimerTestCases = []ForceStopCurrentTimerTestCase{
	{
		// GetCurrentTimerCommand returns sql.ErrNoRows. CurrentTimerNotFoundError is swallowed.
		Name:   "NotFound",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *dbwheeleffectsmock.DatabaseMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)
			wheelDb := new(dbwheeleffectsmock.DatabaseMock)

			timerDb.On("GetCurrentTimerCommand", 1, 1).Return(typetimers.CurrentTimer{}, sql.ErrNoRows)

			return timerDb, gamesDb, wheelDb
		},
		ExpectNoError: true,
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
		// Timer is already Finished. The IncorrectStateConflictError will return.
		Name:   "IncorrectState_Finished",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *dbwheeleffectsmock.DatabaseMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)
			wheelDb := new(dbwheeleffectsmock.DatabaseMock)

			timerDb.On("GetCurrentTimerCommand", 1, 1).Return(finishedCurrentTimer, nil)

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

			timerDb.On("GetCurrentTimerCommand", 1, 1).Return(runningCurrentTimer, nil)
			timerDb.On("ActTimerCommand", 1, typetimers.TimerStateFinished, mock.AnythingOfType("time.Duration")).
				Return(dbError)

			return timerDb, gamesDb, wheelDb
		},
		ExpectedErrorIs: dbError,
	},
	{
		// Regression test: after ActTimerCommand successfully transitions the timer to Finished,
		// the re-fetch of the timer finds no rows (e.g. get_current_timer excludes finished timers).
		// This must not leak a raw sql.ErrNoRows - it should be swallowed just like the initial fetch.
		Name:   "Success_ReFetchNotFound",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *dbwheeleffectsmock.DatabaseMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)
			wheelDb := new(dbwheeleffectsmock.DatabaseMock)

			timerDb.On("GetCurrentTimerCommand", 1, 1).Once().Return(runningCurrentTimer, nil)
			timerDb.On("ActTimerCommand", 1, typetimers.TimerStateFinished, mock.AnythingOfType("time.Duration")).
				Return(nil)
			timerDb.On("GetCurrentTimerCommand", 1, 1).Once().Return(typetimers.CurrentTimer{}, sql.ErrNoRows)

			return timerDb, gamesDb, wheelDb
		},
		ExpectNoError: true,
	},
	{
		// Timer is Running. ActTimerCommand transitions it to Finished and the re-fetch finds it.
		Name:   "Success_ReFetchFound",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *dbwheeleffectsmock.DatabaseMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)
			wheelDb := new(dbwheeleffectsmock.DatabaseMock)

			timerDb.On("GetCurrentTimerCommand", 1, 1).Once().Return(runningCurrentTimer, nil)
			timerDb.On("ActTimerCommand", 1, typetimers.TimerStateFinished, mock.AnythingOfType("time.Duration")).
				Return(nil)
			timerDb.On("GetCurrentTimerCommand", 1, 1).Once().Return(finishedCurrentTimer, nil)

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
			if testCase.ExpectedErrorAs != nil {
				require.Error(test, err)
				require.ErrorAs(test, err, &testCase.ExpectedErrorAs)
			}

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
