package srvtimers_test

import (
	srvpoints "FGG-Service/src/points/service"
	typepoints "FGG-Service/src/points/type"
	"FGG-Service/src/sysparams/types"
	"FGG-Service/src/timers/service"
	"FGG-Service/src/timers/types"
	"FGG-Service/tests/games/mock"
	dbpointsmock "FGG-Service/tests/points/mock"
	srvsysparamsmock "FGG-Service/tests/sysparams/srvmock"
	"FGG-Service/tests/timers/mock/dbtimers"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// A completed timer adds its time to the game it was started for. delete_ended_timers reports the
// full duration as the time spent.
func TestSrvTimers_StopAllCompletedTimers_AddsTimeToGame(test *testing.T) {
	// Arrange
	timerDb := new(dbtimermock.DatabaseMock)
	gamesDb := new(dbgamesmock.DatabaseMock)
	pointsDb := new(dbpointsmock.DatabaseMock)
	sysParams := new(srvsysparamsmock.ServiceMock)

	endedTimer := typetimers.EndedTimer{
		Id:        1,
		UserId:    2,
		PartyId:   1,
		GameId:    7,
		State:     typetimers.TimerStateRunning,
		Duration:  2 * time.Hour,
		TimeSpent: 2 * time.Hour,
	}

	timerDb.On("GetCompletedTimerUsersCommand").Return([]typetimers.EndedTimer{endedTimer}, nil)
	timerDb.On("GetCurrentTimerCommand", 2, 1).Return(typetimers.CurrentTimer{}, sql.ErrNoRows)
	sysParams.On("GetInt", typesysparams.ParamAvailableRollChangeByTimer).Return(1, nil)
	sysParams.On("GetInt", typesysparams.ParamTerritoryHourChangeByTimer).Return(1, nil)
	sysParams.On("GetInt", typesysparams.ParamExperiencePointChangeByTimer).Return(1, nil)
	// Point changes are covered elsewhere; failing the lookup keeps them out of this test.
	pointsDb.On("GetPointTypeByNameCommand", 1, mock.Anything).Return(typepoints.PointTypeInfo{}, dbError)
	gamesDb.On("ChangeGameTimeSpentCommand", 2, 1, 7, 2*time.Hour, 2, (*int)(nil)).Return(nil)

	sut := srvtimers.Service{
		Database:         timerDb,
		GamesDatabase:    gamesDb,
		PointsService:    &srvpoints.Service{Database: pointsDb},
		SysParamsService: sysParams,
	}

	// Act
	err := sut.StopAllCompletedTimers()

	// Assert
	require.NoError(test, err)
	gamesDb.AssertExpectations(test)
}
