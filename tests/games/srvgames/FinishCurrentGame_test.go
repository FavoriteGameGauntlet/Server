package srvgames_test

import (
	"FGG-Service/src/common"
	"FGG-Service/src/games/srvgames"
	"FGG-Service/src/games/typegames"
	"FGG-Service/src/timers/typetimers"
	"FGG-Service/tests/games/dbgamesmock"
	"FGG-Service/tests/games/srvgamesmock"
	"FGG-Service/tests/timers/srvtimersmock"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type FinishCurrentGameTestCase struct {
	Name            string
	UserId          int
	SetupMock       func() (*dbgamesmock.DatabaseMock, *srvtimersmock.ServiceMock, *srvgamesmock.GettingServiceMock)
	ExpectedGame    typegames.CurrentGame
	ExpectedErrorAs any
	ExpectedErrorIs error
}

var FinishCurrentGameTestCases = []FinishCurrentGameTestCase{
	{
		// Neither the game nor the current timer has any time. The game isn't finished and the timer is
		// left alone.
		Name:   "TimeSpentIsZero",
		UserId: 1,
		SetupMock: func() (*dbgamesmock.DatabaseMock, *srvtimersmock.ServiceMock, *srvgamesmock.GettingServiceMock) {
			databaseMock := new(dbgamesmock.DatabaseMock)
			timerServiceMock := new(srvtimersmock.ServiceMock)
			gettingServiceMock := new(srvgamesmock.GettingServiceMock)

			gettingServiceMock.On("GetCurrentGame", 1, 1).Return(typegames.CurrentGame{Id: 1, Name: "Half-Life 1"}, nil)
			timerServiceMock.On("GetCurrentTimerTimeSpent", 1, 1).Return(time.Duration(0), nil)

			return databaseMock, timerServiceMock, gettingServiceMock
		},
		ExpectedErrorAs: new(common.NotFoundError),
	},
	{
		// GetCurrentTimerTimeSpent returns a database error. The error will return.
		Name:   "GetCurrentTimerTimeSpent_DatabaseError",
		UserId: 1,
		SetupMock: func() (*dbgamesmock.DatabaseMock, *srvtimersmock.ServiceMock, *srvgamesmock.GettingServiceMock) {
			databaseMock := new(dbgamesmock.DatabaseMock)
			timerServiceMock := new(srvtimersmock.ServiceMock)
			gettingServiceMock := new(srvgamesmock.GettingServiceMock)

			gettingServiceMock.On("GetCurrentGame", 1, 1).Return(typegames.CurrentGame{Id: 1, Name: "Half-Life 1"}, nil)
			timerServiceMock.On("GetCurrentTimerTimeSpent", 1, 1).Return(time.Duration(0), dbError)

			return databaseMock, timerServiceMock, gettingServiceMock
		},
		ExpectedErrorIs: dbError,
	},
	{
		// FinishCurrentGameCommand returns a database error. The error will return.
		Name:   "FinishCurrentGameCommand_DatabaseError",
		UserId: 1,
		SetupMock: func() (*dbgamesmock.DatabaseMock, *srvtimersmock.ServiceMock, *srvgamesmock.GettingServiceMock) {
			databaseMock := new(dbgamesmock.DatabaseMock)
			timerServiceMock := new(srvtimersmock.ServiceMock)
			gettingServiceMock := new(srvgamesmock.GettingServiceMock)

			gettingServiceMock.On("GetCurrentGame", 1, 1).Return(typegames.CurrentGame{Id: 1, Name: "Half-Life 1"}, nil)
			timerServiceMock.On("GetCurrentTimerTimeSpent", 1, 1).Return(30*time.Minute, nil)
			timerServiceMock.On("ForceStopCurrentTimer", 1, 1).Return(typetimers.Timer{}, nil)
			databaseMock.On("FinishCurrentGameCommand", 1, 1, 1, 1, (*int)(nil)).Return(typegames.UserGame{}, dbError)

			return databaseMock, timerServiceMock, gettingServiceMock
		},
		ExpectedErrorIs: dbError,
	},
	{
		// The game has no time yet, but the current timer does. The game can be finished and returns
		// as FinishCurrentGameCommand returned it.
		Name:   "Success_OnlyTimerTimeSpent",
		UserId: 1,
		SetupMock: func() (*dbgamesmock.DatabaseMock, *srvtimersmock.ServiceMock, *srvgamesmock.GettingServiceMock) {
			databaseMock := new(dbgamesmock.DatabaseMock)
			timerServiceMock := new(srvtimersmock.ServiceMock)
			gettingServiceMock := new(srvgamesmock.GettingServiceMock)

			gettingServiceMock.On("GetCurrentGame", 1, 1).Return(typegames.CurrentGame{Id: 1, Name: "Half-Life 1"}, nil)
			timerServiceMock.On("GetCurrentTimerTimeSpent", 1, 1).Return(30*time.Minute, nil)
			timerServiceMock.On("ForceStopCurrentTimer", 1, 1).Return(typetimers.Timer{}, nil)
			databaseMock.On("FinishCurrentGameCommand", 1, 1, 1, 1, (*int)(nil)).
				Return(typegames.UserGame{Id: 1, Name: "Half-Life 1", TimeSpent: 30 * time.Minute, StartDate: gameStartDate}, nil)

			return databaseMock, timerServiceMock, gettingServiceMock
		},
		ExpectedGame: typegames.CurrentGame{Id: 1, Name: "Half-Life 1", TimeSpent: 30 * time.Minute, StartDate: gameStartDate},
	},
}

func TestSrvGames_FinishCurrentGame(test *testing.T) {
	for _, testCase := range FinishCurrentGameTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			databaseMock, timerServiceMock, gettingServiceMock := testCase.SetupMock()
			sut := srvgames.Service{
				Database:       databaseMock,
				TimerService:   timerServiceMock,
				GettingService: gettingServiceMock,
			}

			// Act
			game, err := sut.FinishCurrentGame(test.Context(), testCase.UserId, 1)

			// Assert
			if testCase.ExpectedErrorAs != nil {
				require.Error(test, err)
				require.ErrorAs(test, err, &testCase.ExpectedErrorAs)
			}

			if testCase.ExpectedErrorIs != nil {
				require.ErrorIs(test, err, testCase.ExpectedErrorIs)
			}

			if testCase.ExpectedErrorAs == nil && testCase.ExpectedErrorIs == nil {
				require.NoError(test, err)
			}

			require.Equal(test, testCase.ExpectedGame, game)

			databaseMock.AssertExpectations(test)
			timerServiceMock.AssertExpectations(test)
			gettingServiceMock.AssertExpectations(test)
		})
	}
}
