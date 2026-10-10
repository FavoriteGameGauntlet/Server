package srvperks

import (
	"FGG-Service/src/common"
	"FGG-Service/src/effects/dbeffects"
	"FGG-Service/src/effects/typeeffects"
	"FGG-Service/src/perks/dbperks"
	"FGG-Service/src/perks/typeperks"
	"context"
	"database/sql"
	"errors"
	"slices"
)

type IService interface {
	GetPerks(ctx context.Context, partyId int) ([]typeperks.Perk, error)
	GetRemovedPerks(ctx context.Context, partyId int) ([]typeperks.Perk, error)
	CreatePerk(ctx context.Context, partyId int, name string, description string, effectId int) (typeperks.Perk, error)
	RemovePerk(ctx context.Context, partyId int, name string) error
	GetUserPerks(ctx context.Context, userId int, partyId int) ([]typeperks.UserPerkDetail, error)
	RevokeUserPerk(ctx context.Context, actorUserId int, userId int, partyId int, perkName string) error
	GetPerkHistory(ctx context.Context, userId int, partyId int) ([]typeperks.PerkHistoryEntry, error)
}

type Service struct {
	Database        dbperks.IDatabase
	EffectsDatabase dbeffects.IDatabase
}

func NewService() *Service {
	return &Service{
		Database:        new(dbperks.Database),
		EffectsDatabase: new(dbeffects.Database),
	}
}

func (s *Service) GetPerks(ctx context.Context, partyId int) (perks []typeperks.Perk, err error) {
	return s.Database.GetActualPerksCommand(ctx, partyId)
}

func (s *Service) GetRemovedPerks(ctx context.Context, partyId int) (perks []typeperks.Perk, err error) {
	return s.Database.GetRemovedPerksCommand(ctx, partyId)
}

// CreatePerk adds a perk to the party catalogue. A perk always wraps one effect, which is what a
// user actually receives when the perk is granted.
func (s *Service) CreatePerk(ctx context.Context, partyId int, name string, description string, effectId int) (perk typeperks.Perk, err error) {
	effects, err := s.EffectsDatabase.GetActualEffectsCommand(ctx, partyId)

	if err != nil {
		return
	}

	isEffectActual := slices.ContainsFunc(effects, func(effect typeeffects.Effect) bool {
		return effect.Id == effectId
	})

	if !isEffectActual {
		err = common.NewEffectNotFoundError(effectId)
		return
	}

	created, err := s.Database.CreatePerkCommand(ctx, partyId, name, description, effectId)

	if err != nil {
		return
	}

	perk = typeperks.Perk{
		Id:          created.Id,
		PartyId:     created.PartyId,
		Name:        created.Name,
		Description: created.Description,
		EffectId:    created.EffectId,
	}

	return
}

func (s *Service) RemovePerk(ctx context.Context, partyId int, name string) (err error) {
	perk, err := s.perkByName(ctx, partyId, name)

	if err != nil {
		return
	}

	return s.Database.RemovePerkCommand(ctx, partyId, perk.Id)
}

// perkByName resolves a perk from the name the API addresses it by.
func (s *Service) perkByName(ctx context.Context, partyId int, name string) (perk typeperks.Perk, err error) {
	perks, err := s.Database.GetActualPerksCommand(ctx, partyId)

	if err != nil {
		return
	}

	for _, candidate := range perks {
		if candidate.Name == name {
			return candidate, nil
		}
	}

	err = common.NewPerkNotFoundError(name)

	return
}

// GetUserPerks lists the perks a user holds, named from the party catalogue.
func (s *Service) GetUserPerks(ctx context.Context, userId int, partyId int) (details []typeperks.UserPerkDetail, err error) {
	userPerks, err := s.Database.GetUserPerksCommand(ctx, userId, partyId)

	if err != nil {
		return
	}

	perksById, err := s.perksById(ctx, partyId)

	if err != nil {
		return
	}

	details = make([]typeperks.UserPerkDetail, 0, len(userPerks))

	for _, userPerk := range userPerks {
		perk := perksById[userPerk.PerkId]

		details = append(details, typeperks.UserPerkDetail{
			Name:         perk.Name,
			Description:  perk.Description,
			ReceivedDate: userPerk.ReceivedDate,
		})
	}

	return
}

// RevokeUserPerk takes a perk away from a user. The revocation is recorded as the actor's, not the
// holder's.
func (s *Service) RevokeUserPerk(ctx context.Context, actorUserId int, userId int, partyId int, perkName string) (err error) {
	perk, err := s.perkByName(ctx, partyId, perkName)

	if err != nil {
		return
	}

	_, err = s.Database.GetUserPerkCommand(ctx, userId, partyId, perk.Id)

	if errors.Is(err, sql.ErrNoRows) {
		return common.NewPerkNotOwnedConflictError(perkName)
	}

	if err != nil {
		return
	}

	return s.Database.DeleteUserPerkCommand(ctx, userId, partyId, perk.Id, actorUserId, nil)
}

// GetPerkHistory lists the recorded perk events of a user, named from the party catalogue.
func (s *Service) GetPerkHistory(ctx context.Context, userId int, partyId int) (history []typeperks.PerkHistoryEntry, err error) {
	entries, err := s.Database.GetPerkHistoryCommand(ctx, userId, partyId)

	if err != nil {
		return
	}

	perksById, err := s.perksById(ctx, partyId)

	if err != nil {
		return
	}

	history = make([]typeperks.PerkHistoryEntry, 0, len(entries))

	for _, entry := range entries {
		history = append(history, typeperks.PerkHistoryEntry{
			Name:        perksById[entry.PerkId].Name,
			Action:      entry.Action,
			ActorUserId: entry.ActorUserId,
			CreatedDate: entry.CreatedDate,
		})
	}

	return
}

// perksById indexes the party catalogue, so reads that name many perks resolve them in one query.
func (s *Service) perksById(ctx context.Context, partyId int) (perksById map[int]typeperks.Perk, err error) {
	perks, err := s.Database.GetActualPerksCommand(ctx, partyId)

	if err != nil {
		return
	}

	removed, err := s.Database.GetRemovedPerksCommand(ctx, partyId)

	if err != nil {
		return
	}

	perksById = make(map[int]typeperks.Perk, len(perks)+len(removed))

	for _, perk := range append(perks, removed...) {
		perksById[perk.Id] = perk
	}

	return
}
