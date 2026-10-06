package srvwheeleffects

import (
	"FGG-Service/src/changes/database"
	srvchanges "FGG-Service/src/changes/service"
	"FGG-Service/src/changes/types"
	"FGG-Service/src/common"
	"FGG-Service/src/sysparams/service"
	"FGG-Service/src/sysparams/types"
	"FGG-Service/src/wheeleffects/database"
	"FGG-Service/src/wheeleffects/types"
	"context"
	"database/sql"
	"errors"
	"math/rand/v2"
)

// defaultCollectionId is a stopgap until wheel-collection selection exists (see project plan) — every
// roll is scoped to this one hardcoded collection.
const defaultCollectionId = 1

type IService interface {
	GetAvailableWheelRows(ctx context.Context, userId int, partyId int) (rows []typewheeleffects.WheelRow, err error)
	MakeEffectRoll(ctx context.Context, userId int, partyId int, isReroll bool) (rolled []typewheeleffects.LastWheelRow, err error)
	ClearLastWheelEffects(ctx context.Context, userId int, partyId int) error
	GetLastRolledWheelEffects(ctx context.Context, userId int, partyId int) (rows []typewheeleffects.LastWheelRow, err error)
	ApplyWheelEffectRoll(ctx context.Context, userId int, partyId int, wheelRowName string, targetUserIds []int) error
	GetLastWheelRowByName(ctx context.Context, userId int, partyId int, wheelRowName string) (row typewheeleffects.LastWheelRow, err error)
	GetEffectHistory(ctx context.Context, userId int, partyId int) (history []typewheeleffects.WheelRowHistory, err error)
	GetEffectHistoryByEffectName(ctx context.Context, userId int, partyId int, effectName string) (history *typewheeleffects.WheelRowHistory, err error)
}

type Service struct {
	Database         dbwheeleffects.IDatabase
	ChangesDatabase  dbchanges.IDatabase
	ChangesService   srvchanges.IService
	SysParamsService srvsysparams.IService
}

func NewService() *Service {
	return &Service{
		Database:         new(dbwheeleffects.Database),
		ChangesDatabase:  new(dbchanges.Database),
		ChangesService:   srvchanges.NewService(),
		SysParamsService: srvsysparams.NewService(),
	}
}

func (s *Service) GetAvailableWheelRows(ctx context.Context, userId int, partyId int) (rows []typewheeleffects.WheelRow, err error) {
	return s.Database.GetAvailableWheelRowsCommand(ctx, userId, partyId, defaultCollectionId)
}

func (s *Service) MakeEffectRoll(ctx context.Context, userId int, partyId int, isReroll bool) (rolled []typewheeleffects.LastWheelRow, err error) {
	candidates, err := s.Database.GetAvailableWheelRowsCommand(ctx, userId, partyId, defaultCollectionId)

	if err != nil {
		return
	}

	sampleSize, err := s.SysParamsService.GetInt(ctx, partyId, typesysparams.ParamMinimumAvailableWheelEffectsForRoll)

	if err != nil {
		return
	}

	if len(candidates) < sampleSize {
		err = common.NewNotEnoughAvailableWheelEffectsConflictError()
		return
	}

	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	err = s.Database.ClearLastWheelEffectsCommand(ctx, userId, partyId)

	if err != nil {
		return
	}

	inputs := make([]typewheeleffects.RolledWheelRowInput, sampleSize)
	for i := range inputs {
		inputs[i] = typewheeleffects.RolledWheelRowInput{
			WheelRowId: candidates[i].Id,
			Position:   i + 1,
		}
	}

	_, err = s.Database.AddLastRolledWheelEffectsCommand(ctx, userId, partyId, inputs)

	if err != nil {
		return
	}

	rolled, err = s.Database.GetLastRolledWheelEffectsCommand(ctx, userId, partyId)

	return
}

func (s *Service) ClearLastWheelEffects(ctx context.Context, userId int, partyId int) error {
	return s.Database.ClearLastWheelEffectsCommand(ctx, userId, partyId)
}

func (s *Service) GetLastRolledWheelEffects(ctx context.Context, userId int, partyId int) (rows []typewheeleffects.LastWheelRow, err error) {
	rows, err = s.Database.GetLastRolledWheelEffectsCommand(ctx, userId, partyId)

	if err != nil {
		return
	}

	if len(rows) == 0 {
		err = common.NewLastWheelEffectsNotFoundError()
	}

	return
}

// ApplyWheelEffectRoll applies the named last-rolled wheel row's Change to every user in
// targetUserIds. This preserves the old feature's ability to apply a roll to users other than the
// roller: the party-level Change template (WheelRow.ChangeId) has no user of its own, so its entries
// are copied into a fresh users.Change with each entry stamped to one of targetUserIds before the
// entries are actually applied.
func (s *Service) ApplyWheelEffectRoll(ctx context.Context, userId int, partyId int, wheelRowName string, targetUserIds []int) (err error) {
	lastRow, err := s.GetLastWheelRowByName(ctx, userId, partyId, wheelRowName)

	if err != nil {
		return
	}

	wheelRow, err := s.Database.GetWheelRowCommand(ctx, partyId, lastRow.Id)

	if err != nil {
		return
	}

	history, err := s.Database.AddWheelEffectHistoryCommand(ctx, userId, partyId, wheelRow.Id, userId, nil)

	if err != nil {
		return
	}

	templateEntries, err := s.ChangesDatabase.GetChangeEntriesJsonbCommand(ctx, partyId, wheelRow.ChangeId)

	if err != nil {
		return
	}

	targetedEntries := make([]typechanges.ChangeEntry, 0, len(templateEntries)*len(targetUserIds))
	for _, targetUserId := range targetUserIds {
		for _, entry := range templateEntries {
			entry.EntryId = nil
			entry.UserId = &targetUserId
			targetedEntries = append(targetedEntries, entry)
		}
	}

	userChange, err := s.ChangesDatabase.CreateUserChangeFromJsonbCommand(ctx, partyId, targetedEntries)

	if err != nil {
		return
	}

	return s.ChangesService.ApplyChangeEntries(ctx, partyId, userChange.Entries, userId, history.Id)
}

func (s *Service) GetEffectHistory(ctx context.Context, userId int, partyId int) (history []typewheeleffects.WheelRowHistory, err error) {
	return s.Database.GetEffectHistoryCommand(ctx, userId, partyId)
}

func (s *Service) GetEffectHistoryByEffectName(ctx context.Context, userId int, partyId int, effectName string) (history *typewheeleffects.WheelRowHistory, err error) {
	notNilHistory, err := s.Database.GetEffectHistoryByEffectNameCommand(ctx, userId, partyId, effectName)

	if errors.Is(err, sql.ErrNoRows) {
		err = nil
		return
	}

	if err != nil {
		return
	}

	history = &notNilHistory

	return
}

// GetLastWheelRowByName finds the named row among the user's currently rolled-but-unapplied rows,
// and errors if it was already applied (a WheelRowHistory row exists for it).
func (s *Service) GetLastWheelRowByName(ctx context.Context, userId int, partyId int, wheelRowName string) (row typewheeleffects.LastWheelRow, err error) {
	lastRows, err := s.Database.GetLastRolledWheelEffectsCommand(ctx, userId, partyId)

	if err != nil {
		return
	}

	found := false
	for _, lastRow := range lastRows {
		if lastRow.Name == wheelRowName {
			row = lastRow
			found = true
			break
		}
	}

	if !found {
		err = common.NewWheelEffectNameNotFoundError()
		return
	}

	_, err = s.Database.GetEffectHistoryByEffectNameCommand(ctx, userId, partyId, wheelRowName)

	if err == nil {
		err = common.NewWheelEffectRollAlreadyAppliedConflictError()
		return
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return
	}

	err = nil

	return
}
