package srvgames_test

import (
	"FGG-Service/src/common"
	"FGG-Service/src/games/service"
	"FGG-Service/src/games/types"
	"FGG-Service/src/timers/types"
	"FGG-Service/tests/games/mock"
	"FGG-Service/tests/games/mock/srvgames"
	"FGG-Service/tests/timers/mock"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type FinishCurrentGameTestCase struct {
	Name            string
	UserId          int
	SetupMock       func() (*dbgamesmock.DatabaseMock, *srvtimersmock.ServiceMock, *srvgamesmock.GettingServiceMock)
	ExpectedErrorAs interface{}
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

			gettingServiceMock.On("GetCurrentGame", 1).Return(typegames.CurrentGame{Id: 1, Name: "Half-Life 1"}, nil)
			timerServiceMock.On("GetCurrentTimerTimeSpent", 1).Return(time.Duration(0), nil)

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

			gettingServiceMock.On("GetCurrentGame", 1).Return(typegames.CurrentGame{Id: 1, Name: "Half-Life 1"}, nil)
			timerServiceMock.On("GetCurrentTimerTimeSpent", 1).Return(time.Duration(0), dbError)

			return databaseMock, timerServiceMock, gettingServiceMock
		},
		ExpectedErrorIs: dbError,
	},
	{
		// The game has no time yet, but the current timer does. The game can be finished.
		Name:   "Success_OnlyTimerTimeSpent",
		UserId: 1,
		SetupMock: func() (*dbgamesmock.DatabaseMock, *srvtimersmock.ServiceMock, *srvgamesmock.GettingServiceMock) {
			databaseMock := new(dbgamesmock.DatabaseMock)
			timerServiceMock := new(srvtimersmock.ServiceMock)
			gettingServiceMock := new(srvgamesmock.GettingServiceMock)

			gettingServiceMock.On("GetCurrentGame", 1).Return(typegames.CurrentGame{Id: 1, Name: "Half-Life 1"}, nil)
			timerServiceMock.On("GetCurrentTimerTimeSpent", 1).Return(30*time.Minute, nil)
			timerServiceMock.On("ForceStopCurrentTimer", 1).Return(typetimers.Timer{}, nil)
			databaseMock.On("FinishCurrentGameCommand", 1, 1, 1, 1, (*int)(nil)).Return(nil)

			return databaseMock, timerServiceMock, gettingServiceMock
		},
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
			err := sut.FinishCurrentGame(testCase.UserId)

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

			databaseMock.AssertExpectations(test)
			timerServiceMock.AssertExpectations(test)
			gettingServiceMock.AssertExpectations(test)
		})
	}
}
