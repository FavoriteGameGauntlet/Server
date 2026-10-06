package srvtimers_test

import (
	"FGG-Service/src/common"
	"FGG-Service/src/games/types"
	"FGG-Service/src/sysparams/types"
	"FGG-Service/src/timers/service"
	"FGG-Service/src/timers/types"
	"FGG-Service/tests/games/mock"
	"FGG-Service/tests/sysparams/srvmock"
	"FGG-Service/tests/timers/mock/dbtimers"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type CreateCurrentTimerTestCase struct {
	Name            string
	UserId          int
	SetupMocks      func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *srvsysparamsmock.ServiceMock)
	ExpectedTimer   *typetimers.Timer
	ExpectedErrorAs any
	ExpectedErrorIs error
}

var timerLastActionDate = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

var currentGame = typegames.UserGame{Id: 1, Name: "Half-Life 1"}

// existingCurrentTimer is what the DB returns; existingTimer is the derived Timer (RemainingTime =
// Duration - TimeSpent) that GetCurrentTimer must return for it.
var existingCurrentTimer = typetimers.CurrentTimer{
	Id:             1,
	GameId:         1,
	State:          typetimers.TimerStateRunning,
	Duration:       2 * time.Hour,
	TimeSpent:      30 * time.Minute,
	LastActionDate: timerLastActionDate,
}

var existingTimer = typetimers.Timer{
	Id:             1,
	Duration:       2 * time.Hour,
	RemainingTime:  90 * time.Minute,
	State:          typetimers.TimerStateRunning,
	LastActionDate: timerLastActionDate,
}

var newCurrentTimer = typetimers.CurrentTimer{
	Id:             2,
	GameId:         1,
	State:          typetimers.TimerStateCreated,
	Duration:       2 * time.Hour,
	TimeSpent:      0,
	LastActionDate: timerLastActionDate,
}

var newTimer = typetimers.Timer{
	Id:             2,
	Duration:       2 * time.Hour,
	RemainingTime:  2 * time.Hour,
	State:          typetimers.TimerStateCreated,
	LastActionDate: timerLastActionDate,
}

var CreateCurrentTimerTestCases = []CreateCurrentTimerTestCase{
	{
		// GetCurrentGameCommand returns sql.ErrNoRows. The CurrentGameNotFoundError will return.
		Name:   "NotFound_NoRows",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *srvsysparamsmock.ServiceMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)
			spSvc := new(srvsysparamsmock.ServiceMock)

			gamesDb.On("GetCurrentGameCommand", 1, 1).Return(typegames.UserGame{}, sql.ErrNoRows)

			return timerDb, gamesDb, spSvc
		},
		ExpectedErrorAs: new(common.NotFoundError),
	},
	{
		// GetCurrentGameCommand returns a database error. The error will return.
		Name:   "GamesDatabaseError",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *srvsysparamsmock.ServiceMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)
			spSvc := new(srvsysparamsmock.ServiceMock)

			gamesDb.On("GetCurrentGameCommand", 1, 1).Return(typegames.UserGame{}, dbError)

			return timerDb, gamesDb, spSvc
		},
		ExpectedErrorIs: dbError,
	},
	{
		// GetCurrentGameCommand succeeds. GetCurrentTimerCommand returns an existing timer. The
		// CurrentTimerAlreadyExists conflict will return and nothing is created.
		Name:   "ExistingTimer_Conflict",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *srvsysparamsmock.ServiceMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)
			spSvc := new(srvsysparamsmock.ServiceMock)

			gamesDb.On("GetCurrentGameCommand", 1, 1).Return(currentGame, nil)
			timerDb.On("GetCurrentTimerCommand", 1, 1).Return(existingCurrentTimer, nil)

			return timerDb, gamesDb, spSvc
		},
		ExpectedErrorAs: new(common.ConflictError),
	},
	{
		// GetCurrentGameCommand succeeds. GetCurrentTimerCommand returns a database error. The error will return.
		Name:   "TimerDatabaseError",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *srvsysparamsmock.ServiceMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)
			spSvc := new(srvsysparamsmock.ServiceMock)

			gamesDb.On("GetCurrentGameCommand", 1, 1).Return(currentGame, nil)
			timerDb.On("GetCurrentTimerCommand", 1, 1).Return(typetimers.CurrentTimer{}, dbError)

			return timerDb, gamesDb, spSvc
		},
		ExpectedErrorIs: dbError,
	},
	{
		// GetCurrentGameCommand and GetCurrentTimerCommand (no rows) succeed.
		// GetInt("TimerDurationInS") returns a database error. The error will return.
		Name:   "SysParams_DatabaseError",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *srvsysparamsmock.ServiceMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)
			spSvc := new(srvsysparamsmock.ServiceMock)

			gamesDb.On("GetCurrentGameCommand", 1, 1).Return(currentGame, nil)
			timerDb.On("GetCurrentTimerCommand", 1, 1).Return(typetimers.CurrentTimer{}, sql.ErrNoRows)
			spSvc.On("GetInt", 1, typesysparams.ParamTimerDurationInS).Return(0, dbError)

			return timerDb, gamesDb, spSvc
		},
		ExpectedErrorIs: dbError,
	},
	{
		// GetCurrentGameCommand, GetCurrentTimerCommand (no rows), and GetInt("TimerDurationInS") succeed.
		// CreateCurrentTimerCommand returns a database error. The error will return.
		Name:   "CreateTimerDatabaseError",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *srvsysparamsmock.ServiceMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)
			spSvc := new(srvsysparamsmock.ServiceMock)

			gamesDb.On("GetCurrentGameCommand", 1, 1).Return(currentGame, nil)
			timerDb.On("GetCurrentTimerCommand", 1, 1).Return(typetimers.CurrentTimer{}, sql.ErrNoRows)
			spSvc.On("GetInt", 1, typesysparams.ParamTimerDurationInS).Return(30, nil)
			timerDb.On("CreateCurrentTimerCommand", 1, 1, currentGame.Id, 30*time.Second).Return(typetimers.CreatedTimer{}, dbError)

			return timerDb, gamesDb, spSvc
		},
		ExpectedErrorIs: dbError,
	},
	{
		// All commands succeed. A new timer is created and returned.
		Name:   "NewTimerCreated",
		UserId: 1,
		SetupMocks: func() (*dbtimermock.DatabaseMock, *dbgamesmock.DatabaseMock, *srvsysparamsmock.ServiceMock) {
			timerDb := new(dbtimermock.DatabaseMock)
			gamesDb := new(dbgamesmock.DatabaseMock)
			spSvc := new(srvsysparamsmock.ServiceMock)

			gamesDb.On("GetCurrentGameCommand", 1, 1).Return(currentGame, nil)
			timerDb.On("GetCurrentTimerCommand", 1, 1).Once().Return(typetimers.CurrentTimer{}, sql.ErrNoRows)
			spSvc.On("GetInt", 1, typesysparams.ParamTimerDurationInS).Return(30, nil)
			timerDb.On("CreateCurrentTimerCommand", 1, 1, currentGame.Id, 30*time.Second).Return(typetimers.CreatedTimer{}, nil)
			timerDb.On("GetCurrentTimerCommand", 1, 1).Once().Return(newCurrentTimer, nil)

			return timerDb, gamesDb, spSvc
		},
		ExpectedTimer: &newTimer,
	},
}

func TestSrvTimers_CreateCurrentTimer(test *testing.T) {
	for _, testCase := range CreateCurrentTimerTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			timerDb, gamesDb, spSvc := testCase.SetupMocks()
			sut := srvtimers.Service{
				Database:         timerDb,
				GamesDatabase:    gamesDb,
				SysParamsService: spSvc,
			}

			// Act
			timer, err := sut.CreateCurrentTimer(testCase.UserId, 1)

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
			spSvc.AssertExpectations(test)
		})
	}
}

// A user who already has a timer gets the CURRENT_TIMER_ALREADY_EXISTS conflict, not a second timer.
func TestSrvTimers_CreateCurrentTimer_ExistingTimer_AlreadyExistsCode(test *testing.T) {
	// Arrange
	timerDb := new(dbtimermock.DatabaseMock)
	gamesDb := new(dbgamesmock.DatabaseMock)

	gamesDb.On("GetCurrentGameCommand", 1, 1).Return(currentGame, nil)
	timerDb.On("GetCurrentTimerCommand", 1, 1).Return(existingCurrentTimer, nil)

	sut := srvtimers.Service{Database: timerDb, GamesDatabase: gamesDb}

	// Act
	_, err := sut.CreateCurrentTimer(1, 1)

	// Assert
	var conflict *common.ConflictError
	require.ErrorAs(test, err, &conflict)
	require.Equal(test, "CURRENT_TIMER_ALREADY_EXISTS", conflict.GetCode())
	timerDb.AssertNotCalled(test, "CreateCurrentTimerCommand")
}
