package srveffects

import (
	"FGG-Service/src/changes/dbchanges"
	"FGG-Service/src/changes/srvchanges"
	"FGG-Service/src/changes/typechanges"
	"FGG-Service/src/common"
	"FGG-Service/src/effects/dbeffects"
	"FGG-Service/src/effects/typeeffects"
	"FGG-Service/src/points/srvpoints"
	"FGG-Service/src/sysparams/srvsysparams"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-co-op/gocron/v2"
)

type IService interface {
	GetEffects(ctx context.Context, partyId int) ([]typeeffects.Effect, error)
	GetRemovedEffects(ctx context.Context, partyId int) ([]typeeffects.Effect, error)
	CreateEffect(ctx context.Context, partyId int, name string, description string, useCount int, duration *time.Duration, entries []typechanges.ChangeEntryInput) (typeeffects.Effect, error)
	RemoveEffect(ctx context.Context, partyId int, name string) error
	GetUserEffects(ctx context.Context, userId int, partyId int) ([]typeeffects.UserEffectDetail, error)
	GetUserEffectViews(ctx context.Context, userId int, partyId int) ([]typeeffects.UserEffectView, error)
	UseEffect(ctx context.Context, actorUserId int, userId int, partyId int, effectName string) error
	EndUserEffect(ctx context.Context, actorUserId int, userId int, partyId int, effectName string) error
	GetEffectHistory(ctx context.Context, userId int, partyId int) ([]typeeffects.EffectHistory, error)
	StopEndedUserEffects(ctx context.Context) error
}

type Service struct {
	Database              dbeffects.IDatabase
	ChangesDatabase       dbchanges.IDatabase
	ChangesService        srvchanges.IService
	SysParamsService      srvsysparams.IService
	PointsService         *srvpoints.Service
	EndedEffectsScheduler gocron.Scheduler
}

func NewService() *Service {
	s := &Service{
		Database:         new(dbeffects.Database),
		ChangesDatabase:  new(dbchanges.Database),
		ChangesService:   srvchanges.NewService(),
		SysParamsService: srvsysparams.NewService(),
		PointsService:    srvpoints.NewService(),
	}

	s.StartEndedEffectsScheduler()

	return s
}

func (s *Service) GetEffects(ctx context.Context, partyId int) (effects []typeeffects.Effect, err error) {
	return s.Database.GetActualEffectsCommand(ctx, partyId)
}

func (s *Service) GetRemovedEffects(ctx context.Context, partyId int) (effects []typeeffects.Effect, err error) {
	return s.Database.GetRemovedEffectsCommand(ctx, partyId)
}

// CreateEffect adds an effect to the party catalogue. The entries describe what using it grants,
// separately from the point modifiers it applies passively while it is active.
func (s *Service) CreateEffect(
	ctx context.Context,
	partyId int,
	name string,
	description string,
	useCount int,
	duration *time.Duration,
	entries []typechanges.ChangeEntryInput) (effect typeeffects.Effect, err error) {
	resolved, err := s.ChangesService.ResolveChangeEntries(ctx, partyId, entries)

	if err != nil {
		return
	}

	created, err := s.Database.CreateEffectCommand(ctx, partyId, name, description, useCount, duration, typechanges.Change{Entries: resolved})

	if err != nil {
		return
	}

	effect = typeeffects.Effect{
		Id:          created.Id,
		PartyId:     created.PartyId,
		Name:        created.Name,
		Description: created.Description,
		UseCount:    created.UseCount,
		Duration:    created.Duration,
	}

	return
}

func (s *Service) RemoveEffect(ctx context.Context, partyId int, name string) (err error) {
	effect, err := s.effectByName(ctx, partyId, name)

	if err != nil {
		return
	}

	return s.Database.RemoveEffectCommand(ctx, partyId, effect.Id)
}

// effectByName resolves an effect from the name the API addresses it by.
func (s *Service) effectByName(ctx context.Context, partyId int, name string) (effect typeeffects.Effect, err error) {
	effects, err := s.Database.GetActualEffectsCommand(ctx, partyId)

	if err != nil {
		return
	}

	for _, candidate := range effects {
		if candidate.Name == name {
			return candidate, nil
		}
	}

	err = common.NewEffectNotFoundError(name)

	return
}

func (s *Service) GetUserEffects(ctx context.Context, userId int, partyId int) (effects []typeeffects.UserEffectDetail, err error) {
	return s.Database.GetUserEffectsCommand(ctx, userId, partyId)
}

func (s *Service) GetEffectHistory(ctx context.Context, userId int, partyId int) (history []typeeffects.EffectHistory, err error) {
	return s.Database.GetEffectHistoryCommand(ctx, userId, partyId)
}

// UseEffect spends one use of an active effect and grants what it carries to its holder. The effect
// history row it produces is the source event of the resulting grants. Another member may use the
// effect, so everything recorded names the actor, not the holder.
func (s *Service) UseEffect(ctx context.Context, actorUserId int, userId int, partyId int, effectName string) (err error) {
	effect, err := s.effectByName(ctx, partyId, effectName)

	if err != nil {
		return
	}

	userEffect, err := s.Database.GetUserEffectCommand(ctx, userId, partyId, effect.Id)

	if errors.Is(err, sql.ErrNoRows) {
		return common.NewEffectNotActiveConflictError(effectName)
	}

	if err != nil {
		return
	}

	if userEffect.UsesLeft < 1 {
		return common.NewEffectUsedUpConflictError(effectName)
	}

	withChange, err := s.Database.GetEffectCommand(ctx, partyId, effect.Id)

	if err != nil {
		return
	}

	usesLeft := userEffect.UsesLeft - 1

	historyEventId, err := s.Database.ChangeUserEffectUsesLeftCommand(ctx, userId, partyId, effect.Id, usesLeft, actorUserId, nil)

	if err != nil {
		return
	}

	if usesLeft == 0 {
		err = s.Database.DeleteUserEffectCommand(ctx, userId, partyId, effect.Id, actorUserId, &historyEventId)

		if err != nil {
			return
		}
	}

	return s.applyEffectChange(ctx, actorUserId, userId, partyId, withChange.Change.Entries, historyEventId)
}

// applyEffectChange stamps the effect's change template onto the user and applies it.
func (s *Service) applyEffectChange(ctx context.Context, actorUserId int, userId int, partyId int, templateEntries []typechanges.ChangeEntry, sourceEventId int) (err error) {
	if len(templateEntries) == 0 {
		return
	}

	targetedEntries := make([]typechanges.ChangeEntry, 0, len(templateEntries))

	for _, entry := range templateEntries {
		entry.EntryId = nil
		entry.UserId = &userId
		targetedEntries = append(targetedEntries, entry)
	}

	userChange, err := s.ChangesDatabase.CreateUserChangeFromJsonbCommand(ctx, partyId, targetedEntries)

	if err != nil {
		return
	}

	return s.ChangesService.ApplyChangeEntries(ctx, partyId, userChange.Entries, actorUserId, sourceEventId)
}

// EndUserEffect ends an active effect before it runs out on its own.
func (s *Service) EndUserEffect(ctx context.Context, actorUserId int, userId int, partyId int, effectName string) (err error) {
	effect, err := s.effectByName(ctx, partyId, effectName)

	if err != nil {
		return
	}

	_, err = s.Database.GetUserEffectCommand(ctx, userId, partyId, effect.Id)

	if errors.Is(err, sql.ErrNoRows) {
		return common.NewEffectNotActiveConflictError(effectName)
	}

	if err != nil {
		return
	}

	return s.Database.DeleteUserEffectCommand(ctx, userId, partyId, effect.Id, actorUserId, nil)
}

// StopEndedUserEffects clears the effects whose duration has run out. Nothing else calls the sweep,
// so it runs on a schedule the way completed timers are stopped.
func (s *Service) StopEndedUserEffects(ctx context.Context) error {
	_, err := s.Database.DeleteEndedUserEffectsCommand(ctx)

	return err
}

// StartEndedEffectsScheduler clears effects whose duration has run out, the way completed timers are
// stopped. Nothing else calls the sweep, so without this an expired effect would stay active.
func (s *Service) StartEndedEffectsScheduler() {
	scheduler, err := gocron.NewScheduler()

	if err != nil {
		panic(err)
	}

	_, err = scheduler.NewJob(
		gocron.DurationJob(common.SchedulerInterval),
		gocron.NewTask(s.StopEndedUserEffects),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)

	if err != nil {
		panic(err)
	}

	s.EndedEffectsScheduler = scheduler

	scheduler.Start()
}

// GetUserEffectViews lists a user's active effects with their passive point modifiers named, since
// the modifiers come back from the schema as point type ids.
func (s *Service) GetUserEffectViews(ctx context.Context, userId int, partyId int) (views []typeeffects.UserEffectView, err error) {
	details, err := s.Database.GetUserEffectsCommand(ctx, userId, partyId)

	if err != nil {
		return
	}

	pointTypes, err := s.PointsService.GetPointTypes(ctx, partyId)

	if err != nil {
		return
	}

	namesByPointTypeId := make(map[int]string, len(pointTypes))
	for _, pointType := range pointTypes {
		namesByPointTypeId[pointType.Id] = pointType.Name
	}

	views = make([]typeeffects.UserEffectView, 0, len(details))

	for _, detail := range details {
		modifiers := make([]typeeffects.PointModifierView, 0, len(detail.Modifiers))

		for _, modifier := range detail.Modifiers {
			modifiers = append(modifiers, typeeffects.PointModifierView{
				PointTypeName: namesByPointTypeId[modifier.PointTypeId],
				Amount:        modifier.Amount,
			})
		}

		views = append(views, typeeffects.UserEffectView{
			Name:        detail.Name,
			Description: detail.Description,
			UseCount:    detail.UseCount,
			UsesLeft:    detail.UsesLeft,
			Duration:    detail.Duration,
			StartedDate: detail.StartedDate,
			Modifiers:   modifiers,
		})
	}

	return
}
