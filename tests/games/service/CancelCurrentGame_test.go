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

type CancelCurrentGameTestCase struct {
	Name            string
	UserId          int
	SetupMock       func() (*dbgamesmock.DatabaseMock, *srvtimersmock.ServiceMock, *srvgamesmock.GettingServiceMock)
	ExpectedGame    typegames.CurrentGame
	ExpectedErrorAs any
	ExpectedErrorIs error
}

var CancelCurrentGameTestCases = []CancelCurrentGameTestCase{
	{
		// GetCurrentGame returns the CurrentGameNotFoundError. The error will return.
		Name:   "GetCurrentGame_NotFound",
		UserId: 1,
		SetupMock: func() (*dbgamesmock.DatabaseMock, *srvtimersmock.ServiceMock, *srvgamesmock.GettingServiceMock) {
			databaseMock := new(dbgamesmock.DatabaseMock)
			timerServiceMock := new(srvtimersmock.ServiceMock)
			gettingServiceMock := new(srvgamesmock.GettingServiceMock)

			gettingServiceMock.
				On("GetCurrentGame",
					1, 1).
				Return(typegames.CurrentGame{}, common.NewCurrentGameNotFoundError())

			return databaseMock, timerServiceMock, gettingServiceMock
		},
		ExpectedErrorAs: new(common.NotFoundError),
	},
	{
		// GetCurrentGame returns a database error. The error will return.
		Name:   "GetCurrentGame_DatabaseError",
		UserId: 1,
		SetupMock: func() (*dbgamesmock.DatabaseMock, *srvtimersmock.ServiceMock, *srvgamesmock.GettingServiceMock) {
			databaseMock := new(dbgamesmock.DatabaseMock)
			timerServiceMock := new(srvtimersmock.ServiceMock)
			gettingServiceMock := new(srvgamesmock.GettingServiceMock)

			gettingServiceMock.
				On("GetCurrentGame",
					1, 1).
				Return(typegames.CurrentGame{}, dbError)

			return databaseMock, timerServiceMock, gettingServiceMock
		},
		ExpectedErrorIs: dbError,
	},
	{
		// GetCurrentGame succeeds. ForceStopCurrentTimer returns a database error. The error will return.
		Name:   "ForceStopCurrentTimer_DatabaseError",
		UserId: 1,
		SetupMock: func() (*dbgamesmock.DatabaseMock, *srvtimersmock.ServiceMock, *srvgamesmock.GettingServiceMock) {
			databaseMock := new(dbgamesmock.DatabaseMock)
			timerServiceMock := new(srvtimersmock.ServiceMock)
			gettingServiceMock := new(srvgamesmock.GettingServiceMock)

			gettingServiceMock.
				On("GetCurrentGame",
					1, 1).
				Return(typegames.CurrentGame{
					Id:   1,
					Name: "Half-Life 1"},
					nil)
			timerServiceMock.
				On("ForceStopCurrentTimer",
					1, 1).
				Return(typetimers.Timer{}, dbError)

			return databaseMock, timerServiceMock, gettingServiceMock
		},
		ExpectedErrorIs: dbError,
	},
	{
		// GetCurrentGame and ForceStopCurrentTimer succeed. CancelCurrentGameCommand returns a database error. The error will return.
		Name:   "CancelCurrentGameCommand_DatabaseError",
		UserId: 1,
		SetupMock: func() (*dbgamesmock.DatabaseMock, *srvtimersmock.ServiceMock, *srvgamesmock.GettingServiceMock) {
			databaseMock := new(dbgamesmock.DatabaseMock)
			timerServiceMock := new(srvtimersmock.ServiceMock)
			gettingServiceMock := new(srvgamesmock.GettingServiceMock)

			gettingServiceMock.
				On("GetCurrentGame",
					1, 1).
				Return(typegames.CurrentGame{Id: 1, Name: "Half-Life 1"}, nil)
			timerServiceMock.
				On("ForceStopCurrentTimer",
					1, 1).
				Return(typetimers.Timer{}, nil)
			databaseMock.
				On("CancelCurrentGameCommand",
					1, 1, 1, 1, (*int)(nil)).
				Return(typegames.UserGame{}, dbError)

			return databaseMock, timerServiceMock, gettingServiceMock
		},
		ExpectedErrorIs: dbError,
	},
	{
		// Everything succeeds. The game returns as CancelCurrentGameCommand returned it.
		Name:   "SuccessReturn",
		UserId: 1,
		SetupMock: func() (*dbgamesmock.DatabaseMock, *srvtimersmock.ServiceMock, *srvgamesmock.GettingServiceMock) {
			databaseMock := new(dbgamesmock.DatabaseMock)
			timerServiceMock := new(srvtimersmock.ServiceMock)
			gettingServiceMock := new(srvgamesmock.GettingServiceMock)

			gettingServiceMock.
				On("GetCurrentGame",
					1, 1).
				Return(typegames.CurrentGame{Id: 1, Name: "Half-Life 1", TimeSpent: time.Hour}, nil)
			timerServiceMock.
				On("ForceStopCurrentTimer",
					1, 1).
				Return(typetimers.Timer{}, nil)
			databaseMock.
				On("CancelCurrentGameCommand",
					1, 1, 1, 1, (*int)(nil)).
				Return(typegames.UserGame{Id: 1, Name: "Half-Life 1", TimeSpent: 90 * time.Minute, StartDate: gameStartDate}, nil)

			return databaseMock, timerServiceMock, gettingServiceMock
		},
		ExpectedGame: typegames.CurrentGame{Id: 1, Name: "Half-Life 1", TimeSpent: 90 * time.Minute, StartDate: gameStartDate},
	},
}

func TestSrvGames_CancelCurrentGame(test *testing.T) {
	for _, testCase := range CancelCurrentGameTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			databaseMock, timerServiceMock, gettingServiceMock := testCase.SetupMock()
			sut := srvgames.Service{
				Database:       databaseMock,
				TimerService:   timerServiceMock,
				GettingService: gettingServiceMock,
			}

			// Act
			game, err := sut.CancelCurrentGame(testCase.UserId, 1)

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
