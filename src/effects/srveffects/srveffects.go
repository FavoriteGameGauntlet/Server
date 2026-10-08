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
	"FGG-Service/src/validator"
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
)

type IService interface {
	GetEffects(ctx context.Context, partyId int) ([]typeeffects.NamedEffect, error)
	GetRemovedEffects(ctx context.Context, partyId int) ([]typeeffects.NamedEffect, error)
	CreateEffect(ctx context.Context, partyId int, name string, description string, useCount *int, duration *time.Duration, entries []typechanges.NamedChangeEntry, modifiers []typeeffects.NamedPointModifier) (typeeffects.NamedEffect, error)
	RemoveEffect(ctx context.Context, partyId int, name string) error
	GetUserEffects(ctx context.Context, userId int, partyId int) ([]typeeffects.UserEffectDetail, error)
	GetNamedUserEffects(ctx context.Context, userId int, partyId int) ([]typeeffects.NamedUserEffect, error)
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

func (s *Service) GetEffects(ctx context.Context, partyId int) (named []typeeffects.NamedEffect, err error) {
	effects, err := s.Database.GetActualEffectsCommand(ctx, partyId)

	if err != nil {
		return
	}

	return s.nameEffects(ctx, partyId, effects)
}

func (s *Service) GetRemovedEffects(ctx context.Context, partyId int) (named []typeeffects.NamedEffect, err error) {
	effects, err := s.Database.GetRemovedEffectsCommand(ctx, partyId)

	if err != nil {
		return
	}

	return s.nameEffects(ctx, partyId, effects)
}

// nameEffects names the passive point modifiers of catalogue effects, since they come back from the
// schema as point type ids.
func (s *Service) nameEffects(ctx context.Context, partyId int, effects []typeeffects.Effect) (named []typeeffects.NamedEffect, err error) {
	namesByPointTypeId, err := s.fetchPointTypeNamesById(ctx, partyId)

	if err != nil {
		return
	}

	named = make([]typeeffects.NamedEffect, 0, len(effects))

	for _, effect := range effects {
		named = append(named, typeeffects.NamedEffect{
			Name:        effect.Name,
			Description: effect.Description,
			UseCount:    effect.UseCount,
			Duration:    effect.Duration,
			Modifiers:   nameModifiers(effect.Modifiers, namesByPointTypeId),
		})
	}

	return
}

// nameModifiers swaps the point type ids of modifiers for the names the API uses.
func nameModifiers(modifiers []typeeffects.PointModifier, namesByPointTypeId map[int]string) []typeeffects.NamedPointModifier {
	named := make([]typeeffects.NamedPointModifier, 0, len(modifiers))

	for _, modifier := range modifiers {
		named = append(named, typeeffects.NamedPointModifier{
			PointTypeName: namesByPointTypeId[modifier.PointTypeId],
			Amount:        modifier.Amount,
		})
	}

	return named
}

// fetchPointTypeNamesById maps the party's point type ids to the names the API uses.
func (s *Service) fetchPointTypeNamesById(ctx context.Context, partyId int) (namesByPointTypeId map[int]string, err error) {
	pointTypes, err := s.PointsService.GetPointTypes(ctx, partyId)

	if err != nil {
		return
	}

	namesByPointTypeId = make(map[int]string, len(pointTypes))
	for _, pointType := range pointTypes {
		namesByPointTypeId[pointType.Id] = pointType.Name
	}

	return
}

// CreateEffect adds an effect to the party catalogue. The entries describe what using it grants,
// separately from the modifiers, the point changes it applies passively while it is active. A nil
// useCount makes an effect that can be used without limit.
func (s *Service) CreateEffect(
	ctx context.Context,
	partyId int,
	name string,
	description string,
	useCount *int,
	duration *time.Duration,
	entries []typechanges.NamedChangeEntry,
	modifiers []typeeffects.NamedPointModifier) (effect typeeffects.NamedEffect, err error) {
	resolved, err := s.ChangesService.ResolveChangeEntries(ctx, partyId, entries)

	if err != nil {
		return
	}

	resolvedModifiers, err := s.resolveModifiers(ctx, partyId, modifiers)

	if err != nil {
		return
	}

	created, err := s.Database.CreateEffectCommand(ctx, partyId, name, description, useCount, duration, typechanges.Change{Entries: resolved}, resolvedModifiers)

	if err != nil {
		return
	}

	effect = typeeffects.NamedEffect{
		Name:        created.Name,
		Description: created.Description,
		UseCount:    created.UseCount,
		Duration:    created.Duration,
		Modifiers:   modifiers,
	}

	return
}

// resolveModifiers turns the point type names of the modifiers into the ids the schema stores. A point
// type can carry only one modifier per effect.
func (s *Service) resolveModifiers(ctx context.Context, partyId int, modifiers []typeeffects.NamedPointModifier) ([]typeeffects.PointModifier, error) {
	resolved := make([]typeeffects.PointModifier, 0, len(modifiers))
	seenPointTypeIds := make(map[int]bool, len(modifiers))

	for _, modifier := range modifiers {
		err := validator.ValidateEffectModifierAmount(modifier.Amount)

		if err != nil {
			return nil, err
		}

		pointType, err := s.PointsService.GetPointTypeByName(ctx, partyId, modifier.PointTypeName)

		if err != nil {
			return nil, err
		}

		if seenPointTypeIds[pointType.Id] {
			return nil, common.NewEffectModifierDuplicateUnprocessableError(modifier.PointTypeName)
		}

		seenPointTypeIds[pointType.Id] = true

		resolved = append(resolved, typeeffects.PointModifier{PointTypeId: pointType.Id, Amount: modifier.Amount})
	}

	return resolved, nil
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
// effect, so everything recorded names the actor, not the holder. An effect without a use limit has
// nothing to spend: its history row carries no count and it never runs out.
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

	if userEffect.UsesLeft != nil && *userEffect.UsesLeft < 1 {
		return common.NewEffectUsedUpConflictError(effectName)
	}

	withChange, err := s.Database.GetEffectCommand(ctx, partyId, effect.Id)

	if err != nil {
		return
	}

	var usesLeft *int
	if userEffect.UsesLeft != nil {
		spent := *userEffect.UsesLeft - 1
		usesLeft = &spent
	}

	historyEventId, err := s.Database.ChangeUserEffectUsesLeftCommand(ctx, userId, partyId, effect.Id, usesLeft, actorUserId, nil)

	if err != nil {
		return
	}

	if usesLeft != nil && *usesLeft == 0 {
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
// so it runs on a schedule the way completed timers are stopped. An effect that ran out on its own
// has no actor, so its holder is recorded as the one who ended it, the way a finished timer is.
func (s *Service) StopEndedUserEffects(ctx context.Context) error {
	endedEffects, err := s.Database.GetEndedUserEffectsCommand(ctx)

	if err != nil {
		return err
	}

	for _, ended := range endedEffects {
		err = s.Database.DeleteUserEffectCommand(ctx, ended.UserId, ended.PartyId, ended.EffectId, ended.UserId, nil)

		if err != nil {
			slog.Error("StopEndedUserEffect", "userEffectId", ended.Id, "userId", ended.UserId, "error", err)
		}
	}

	return nil
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

// GetNamedUserEffects lists a user's active effects with their passive point modifiers named, since
// the modifiers come back from the schema as point type ids.
func (s *Service) GetNamedUserEffects(ctx context.Context, userId int, partyId int) (named []typeeffects.NamedUserEffect, err error) {
	details, err := s.Database.GetUserEffectsCommand(ctx, userId, partyId)

	if err != nil {
		return
	}

	namesByPointTypeId, err := s.fetchPointTypeNamesById(ctx, partyId)

	if err != nil {
		return
	}

	named = make([]typeeffects.NamedUserEffect, 0, len(details))

	for _, detail := range details {
		named = append(named, typeeffects.NamedUserEffect{
			Name:        detail.Name,
			Description: detail.Description,
			UseCount:    detail.UseCount,
			UsesLeft:    detail.UsesLeft,
			Duration:    detail.Duration,
			StartedDate: detail.StartedDate,
			Modifiers:   nameModifiers(detail.Modifiers, namesByPointTypeId),
		})
	}

	return
}
