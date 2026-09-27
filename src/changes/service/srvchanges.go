package srvchanges

import (
	"FGG-Service/src/changes/database"
	"FGG-Service/src/changes/types"
	"FGG-Service/src/effects/database"
	"FGG-Service/src/items/database"
	"FGG-Service/src/perks/database"
	"FGG-Service/src/points/service"
	"fmt"
)

// IService applies the entries of a Change (see typechanges.ChangeEntry) to the point/item/perk/effect
// values of the users they target. Nothing in the SQL schema does this dispatch itself — every
// history-recording function (create_manual_history, create_wheel_row_history, create_exchange_history)
// is record-only, so it happens here.
type IService interface {
	ApplyChangeEntries(partyId int, entries []typechanges.ChangeEntry, actorUserId int, sourceEventId int) error
	ResolveChangeEntries(partyId int, inputs []typechanges.ChangeEntryInput) ([]typechanges.ChangeEntry, error)
}

type Service struct {
	Database        dbchanges.IDatabase
	ItemsDatabase   dbitems.IDatabase
	PerksDatabase   dbperks.IDatabase
	EffectsDatabase dbeffects.IDatabase
	PointsService   *srvpoints.Service
}

func NewService() *Service {
	return &Service{
		Database:        new(dbchanges.Database),
		ItemsDatabase:   new(dbitems.Database),
		PerksDatabase:   new(dbperks.Database),
		EffectsDatabase: new(dbeffects.Database),
		PointsService:   srvpoints.NewService(),
	}
}

// ApplyChangeEntries grants each entry to its UserId. sourceEventId is the HistoryEvents row that
// caused this (e.g. a wheel roll's WheelRowHistory.Id or a manual grant's ManualHistory.Id) and is
// recorded as the origin of every resulting point/item/perk/effect history row.
func (s *Service) ApplyChangeEntries(partyId int, entries []typechanges.ChangeEntry, actorUserId int, sourceEventId int) error {
	for _, entry := range entries {
		if entry.UserId == nil {
			return fmt.Errorf("change entry has no target user id")
		}

		userId := *entry.UserId

		var err error

		switch {
		case entry.PointTypeId != nil:
			err = s.PointsService.ChangeUserPointValueClamped(userId, partyId, *entry.PointTypeId, entry.Amount, actorUserId, sourceEventId)
		case entry.ItemId != nil:
			_, err = s.ItemsDatabase.CreateUserItemCommand(userId, partyId, *entry.ItemId, actorUserId, sourceEventId)
		case entry.EffectId != nil:
			_, err = s.EffectsDatabase.CreateUserEffectCommand(userId, partyId, *entry.EffectId, actorUserId, sourceEventId)
		case entry.PerkId != nil:
			err = s.applyPerkEntry(userId, partyId, *entry.PerkId, actorUserId, sourceEventId)
		default:
			err = fmt.Errorf("change entry has no point/item/perk/effect target")
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// applyPerkEntry grants a perk by first granting its underlying Effect (party.Perks.EffectId is
// mandatory), then granting the perk itself. Both record the originating event as their source.
func (s *Service) applyPerkEntry(userId int, partyId int, perkId int, actorUserId int, sourceEventId int) error {
	perk, err := s.PerksDatabase.GetPerkCommand(partyId, perkId)

	if err != nil {
		return err
	}

	_, err = s.EffectsDatabase.CreateUserEffectCommand(userId, partyId, perk.EffectId, actorUserId, sourceEventId)

	if err != nil {
		return err
	}

	_, err = s.PerksDatabase.CreateUserPerkCommand(userId, partyId, perkId, actorUserId, sourceEventId)

	return err
}
