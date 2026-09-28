package srvtimers

import (
	"FGG-Service/src/common"
	"FGG-Service/src/games/database"
	"FGG-Service/src/points/service"
	"FGG-Service/src/points/type"
	"FGG-Service/src/sysparams/service"
	"FGG-Service/src/sysparams/types"
	"FGG-Service/src/timers/database"
	"FGG-Service/src/timers/types"
	"FGG-Service/src/wheeleffects/database"
	"database/sql"
	"errors"
	"time"

	"github.com/go-co-op/gocron/v2"
)

// defaultPartyId is a stopgap until real party-context resolution exists (see project plan).
const defaultPartyId = 1

type IService interface {
	ForceStopCurrentTimer(userId int) (typetimers.Timer, error)
}

// toTimer computes the derived Timer view (with RemainingTime) from a raw CurrentTimer DB row.
func toTimer(currentTimer typetimers.CurrentTimer) typetimers.Timer {
	remainingTime := currentTimer.Duration - currentTimer.TimeSpent

	if remainingTime < 0 {
		remainingTime = 0
	}

	return typetimers.Timer{
		Id:             currentTimer.Id,
		Duration:       currentTimer.Duration,
		RemainingTime:  remainingTime,
		State:          currentTimer.State,
		LastActionDate: currentTimer.LastActionDate,
	}
}

type Service struct {
	Database               dbtimers.IDatabase
	GamesDatabase          dbgames.IDatabase
	PointsService          *srvpoints.Service
	WheelEffectsDatabase   dbwheeleffects.IDatabase
	SysParamsService       srvsysparams.IService
	TimerFinisherScheduler gocron.Scheduler
}

func NewService() *Service {
	db := new(dbtimers.Database)
	gdb := new(dbgames.Database)
	ps := srvpoints.NewService()
	wedb := new(dbwheeleffects.Database)
	sp := srvsysparams.NewService()

	s := &Service{
		Database:             db,
		GamesDatabase:        gdb,
		PointsService:        ps,
		WheelEffectsDatabase: wedb,
		SysParamsService:     sp,
	}

	s.StartTimerFinisherScheduler()

	return s
}

func (s *Service) StartTimerFinisherScheduler() {
	scheduler, err := gocron.NewScheduler()

	if err != nil {
		panic(err)
	}

	intervalInS, err := s.SysParamsService.GetInt(typesysparams.ParamTimerFinisherSchedulerIntervalInS)

	if err != nil {
		panic(err)
	}

	_, err = scheduler.NewJob(
		gocron.DurationJob(time.Duration(intervalInS)*time.Second),
		gocron.NewTask(s.StopAllCompletedTimers),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)

	if err != nil {
		panic(err)
	}

	s.TimerFinisherScheduler = scheduler

	scheduler.Start()
}

func (s *Service) GetOrCreateCurrentTimer(userId int) (timer typetimers.Timer, err error) {
	game, err := s.GamesDatabase.GetCurrentGameCommand(userId, defaultPartyId)

	if errors.Is(err, sql.ErrNoRows) {
		err = common.NewCurrentGameNotFoundError()
		return
	}

	if err != nil {
		return
	}

	currentTimer, err := s.Database.GetCurrentTimerCommand(userId, defaultPartyId)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return
	}

	if !errors.Is(err, sql.ErrNoRows) {
		timer = toTimer(currentTimer)
		return
	}

	rollCount, err := s.WheelEffectsDatabase.GetAvailableRollsCountCommand(userId)

	if err != nil {
		return
	}

	maximumAvailableRollCountForTimer, err := s.SysParamsService.GetInt(typesysparams.ParamMaximumAvailableRollCountForTimer)

	if err != nil {
		return
	}

	if rollCount >= maximumAvailableRollCountForTimer {
		err = common.NewAvailableRollsExistConflictError()
		return
	}

	durationInS, err := s.SysParamsService.GetInt(typesysparams.ParamTimerDurationInS)

	if err != nil {
		return
	}

	_, err = s.Database.CreateCurrentTimerCommand(userId, defaultPartyId, game.Id, time.Duration(durationInS)*time.Second)

	if err != nil {
		return
	}

	currentTimer, err = s.Database.GetCurrentTimerCommand(userId, defaultPartyId)

	if err != nil {
		return
	}

	timer = toTimer(currentTimer)

	return
}

func (s *Service) StartCurrentTimer(userId int) (typetimers.Timer, error) {
	return s.actCurrentTimer(
		userId,
		typetimers.TimerStateRunning,
		[]typetimers.TimerStateType{
			typetimers.TimerStateRunning,
			typetimers.TimerStateFinished,
		})
}

func (s *Service) PauseCurrentTimer(userId int) (typetimers.Timer, error) {
	return s.actCurrentTimer(
		userId,
		typetimers.TimerStatePaused,
		[]typetimers.TimerStateType{
			typetimers.TimerStateCreated,
			typetimers.TimerStatePaused,
			typetimers.TimerStateFinished,
		})
}

func (s *Service) StopCurrentTimer(userId int) (typetimers.Timer, error) {
	return s.actCurrentTimer(
		userId,
		typetimers.TimerStateFinished,
		[]typetimers.TimerStateType{
			typetimers.TimerStateCreated,
			typetimers.TimerStateFinished,
		})
}

func (s *Service) ForceStopCurrentTimer(userId int) (timer typetimers.Timer, err error) {
	timer, err = s.actCurrentTimer(
		userId,
		typetimers.TimerStateFinished,
		[]typetimers.TimerStateType{
			typetimers.TimerStateFinished,
		})

	var notFoundError *common.NotFoundError
	if err != nil && !errors.As(err, &notFoundError) {
		return
	}

	err = nil
	return
}

func (s *Service) actCurrentTimer(
	userId int,
	timerState typetimers.TimerStateType,
	incorrectStates []typetimers.TimerStateType) (timer typetimers.Timer, err error) {

	currentTimer, err := s.Database.GetCurrentTimerCommand(userId, defaultPartyId)

	if errors.Is(err, sql.ErrNoRows) {
		err = common.NewCurrentTimerNotFoundError()
		return
	}

	if err != nil {
		return
	}

	timer = toTimer(currentTimer)

	for _, state := range incorrectStates {
		if timer.State == state {
			err = common.NewCurrentTimerIncorrectStateConflictError(timer.State)
			return
		}
	}

	// get_timer already includes the elapsed running time in TimeSpent.
	timeSpent := min(currentTimer.TimeSpent, currentTimer.Duration)

	err = s.Database.ActTimerCommand(timer.Id, timerState, timeSpent)

	if err != nil {
		return
	}

	currentTimer, err = s.Database.GetCurrentTimerCommand(userId, defaultPartyId)

	if errors.Is(err, sql.ErrNoRows) {
		err = common.NewCurrentTimerNotFoundError()
		return
	}

	if err != nil {
		return
	}

	timer = toTimer(currentTimer)

	return
}

func (s *Service) StopAllCompletedTimers() error {
	endedTimers, err := s.Database.GetCompletedTimerUsersCommand()

	if err != nil {
		return err
	}

	availableRollChangeByTimer, err := s.SysParamsService.GetInt(typesysparams.ParamAvailableRollChangeByTimer)

	if err != nil {
		return err
	}

	territoryHourChangeByTimer, err := s.SysParamsService.GetInt(typesysparams.ParamTerritoryHourChangeByTimer)

	if err != nil {
		return err
	}

	experiencePointChangeByTimer, err := s.SysParamsService.GetInt(typesysparams.ParamExperiencePointChangeByTimer)

	if err != nil {
		return err
	}

	for _, endedTimer := range endedTimers {
		_, _ = s.StopCurrentTimer(endedTimer.UserId)
		_ = s.GamesDatabase.ChangeGameTimeSpentCommand(endedTimer.UserId, endedTimer.PartyId, endedTimer.GameId, endedTimer.TimeSpent, endedTimer.UserId, nil)
		_ = s.PointsService.ChangePointValueByTypeNameNoHistory(endedTimer.UserId, defaultPartyId, typepoints.PointTypeAvailableRolls, availableRollChangeByTimer)
		_ = s.PointsService.ChangePointValueByTypeNameNoHistory(endedTimer.UserId, defaultPartyId, typepoints.PointTypeTerritoryHours, territoryHourChangeByTimer)
		_ = s.PointsService.ChangePointValueByTypeNameNoHistory(endedTimer.UserId, defaultPartyId, typepoints.PointTypeExperiencePoints, experiencePointChangeByTimer)
	}

	return nil
}
