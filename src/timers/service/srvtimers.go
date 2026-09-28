package srvtimers

import (
	"FGG-Service/src/changes/service"
	"FGG-Service/src/changes/types"
	"FGG-Service/src/common"
	"FGG-Service/src/games/database"
	"FGG-Service/src/sysparams/service"
	"FGG-Service/src/sysparams/types"
	"FGG-Service/src/timers/database"
	"FGG-Service/src/timers/types"
	"FGG-Service/src/wheeleffects/database"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
)

// defaultPartyId is a stopgap until real party-context resolution exists (see project plan).
const defaultPartyId = 1

type IService interface {
	GetCurrentTimerTimeSpent(userId int) (time.Duration, error)
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
	ChangesService         srvchanges.IService
	WheelEffectsDatabase   dbwheeleffects.IDatabase
	SysParamsService       srvsysparams.IService
	TimerFinisherScheduler gocron.Scheduler
}

func NewService() *Service {
	db := new(dbtimers.Database)
	gdb := new(dbgames.Database)
	cs := srvchanges.NewService()
	wedb := new(dbwheeleffects.Database)
	sp := srvsysparams.NewService()

	s := &Service{
		Database:             db,
		GamesDatabase:        gdb,
		ChangesService:       cs,
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

// GetCurrentTimerTimeSpent returns how much of the current timer has passed, or zero without a timer.
func (s *Service) GetCurrentTimerTimeSpent(userId int) (time.Duration, error) {
	currentTimer, err := s.Database.GetCurrentTimerCommand(userId, defaultPartyId)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	return min(currentTimer.TimeSpent, currentTimer.Duration), nil
}

// ForceStopCurrentTimer deletes the current timer, if any, and adds the time already spent on it to
// its game. It grants none of the rewards of a completed timer.
func (s *Service) ForceStopCurrentTimer(userId int) (timer typetimers.Timer, err error) {
	deletedTimer, err := s.Database.DeleteCurrentTimerCommand(userId, defaultPartyId)

	if errors.Is(err, sql.ErrNoRows) {
		err = nil
		return
	}

	if err != nil {
		return
	}

	if deletedTimer.TimeSpent > 0 {
		err = s.GamesDatabase.ChangeGameTimeSpentCommand(userId, deletedTimer.PartyId, deletedTimer.GameId, deletedTimer.TimeSpent, userId, nil)

		if err != nil {
			return
		}
	}

	timer = toTimer(typetimers.CurrentTimer{
		Id:             deletedTimer.Id,
		GameId:         deletedTimer.GameId,
		State:          deletedTimer.State,
		Duration:       deletedTimer.Duration,
		LastActionDate: deletedTimer.LastActionDate,
		TimeSpent:      deletedTimer.TimeSpent,
	})

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

	for _, endedTimer := range endedTimers {
		err = s.completeTimer(endedTimer)

		if err != nil {
			slog.Error("CompleteTimer", "timerId", endedTimer.Id, "userId", endedTimer.UserId, "error", err)
		}
	}

	return nil
}

// completeTimer records a completed timer as a timer history event, which is then the source of the
// time it adds to its game and of the party's timer reward. A party without a reward still gets the
// event and the game time.
func (s *Service) completeTimer(timer typetimers.EndedTimer) (err error) {
	var rewardId *int
	var rewardEntries []typechanges.ChangeEntry

	reward, err := s.Database.GetTimerRewardCommand(timer.PartyId)

	if err == nil {
		rewardId = &reward.Id
		rewardEntries = reward.Change.Entries
	} else if !errors.Is(err, sql.ErrNoRows) {
		return
	}

	history, err := s.Database.CreateTimerHistoryCommand(timer.UserId, timer.PartyId, rewardId, timer.UserId)

	if err != nil {
		return
	}

	err = s.GamesDatabase.ChangeGameTimeSpentCommand(timer.UserId, timer.PartyId, timer.GameId, timer.TimeSpent, timer.UserId, &history.Id)

	if err != nil {
		return
	}

	for i := range rewardEntries {
		rewardEntries[i].UserId = &timer.UserId
	}

	return s.ChangesService.ApplyChangeEntries(timer.PartyId, rewardEntries, timer.UserId, history.Id)
}

func (s *Service) GetTimerReward(partyId int) ([]typechanges.ChangeEntryInput, error) {
	return s.Database.GetTimerRewardEntriesCommand(partyId)
}

// SetTimerReward replaces the party's timer reward. The previous one is kept as removed, so the
// history of timers completed under it still points at what they granted.
func (s *Service) SetTimerReward(partyId int, entries []typechanges.ChangeEntryInput) (reward []typechanges.ChangeEntryInput, err error) {
	resolved, err := s.ChangesService.ResolveChangeEntries(partyId, entries)

	if err != nil {
		return
	}

	_, err = s.Database.SetTimerRewardCommand(partyId, typechanges.Change{Entries: resolved})

	if err != nil {
		return
	}

	return s.Database.GetTimerRewardEntriesCommand(partyId)
}

func (s *Service) RemoveTimerReward(partyId int) error {
	return s.Database.RemoveTimerRewardCommand(partyId)
}
