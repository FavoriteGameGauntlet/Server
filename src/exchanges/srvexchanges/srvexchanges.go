package srvexchanges

import (
	"FGG-Service/src/changes/dbchanges"
	"FGG-Service/src/changes/srvchanges"
	"FGG-Service/src/changes/typechanges"
	"FGG-Service/src/common"
	"FGG-Service/src/exchanges/dbexchanges"
	"FGG-Service/src/exchanges/typeexchanges"
	"FGG-Service/src/points/srvpoints"
	"FGG-Service/src/points/typepoints"
	"context"
)

type IService interface {
	GetExchanges(ctx context.Context, partyId int) ([]typeexchanges.Exchange, error)
	GetRemovedExchanges(ctx context.Context, partyId int) ([]typeexchanges.Exchange, error)
	CreateExchange(ctx context.Context, partyId int, name string, description string, sourceEntries []typechanges.ChangeEntryInput, targetEntries []typechanges.ChangeEntryInput) (typeexchanges.Exchange, error)
	RemoveExchange(ctx context.Context, partyId int, name string) error
	UseExchange(ctx context.Context, userId int, partyId int, exchangeName string, targetUserIds []int) error
	GetExchangeHistory(ctx context.Context, userId int, partyId int) ([]typeexchanges.ExchangeHistory, error)
}

type Service struct {
	Database        dbexchanges.IDatabase
	ChangesDatabase dbchanges.IDatabase
	ChangesService  srvchanges.IService
	PointsService   *srvpoints.Service
}

func NewService() *Service {
	return &Service{
		Database:        new(dbexchanges.Database),
		ChangesDatabase: new(dbchanges.Database),
		ChangesService:  srvchanges.NewService(),
		PointsService:   srvpoints.NewService(),
	}
}

func (s *Service) GetExchanges(ctx context.Context, partyId int) (exchanges []typeexchanges.Exchange, err error) {
	return s.Database.GetActualExchangesCommand(ctx, partyId)
}

func (s *Service) GetRemovedExchanges(ctx context.Context, partyId int) (exchanges []typeexchanges.Exchange, err error) {
	return s.Database.GetRemovedExchangesCommand(ctx, partyId)
}

func (s *Service) GetExchangeHistory(ctx context.Context, userId int, partyId int) (history []typeexchanges.ExchangeHistory, err error) {
	return s.Database.GetExchangeHistoryCommand(ctx, userId, partyId)
}

// CreateExchange adds an exchange to the party catalogue. The source entries are what it costs and
// the target entries are what it gives; both carry the sign they are applied with, so a cost is
// negative.
func (s *Service) CreateExchange(
	ctx context.Context,
	partyId int,
	name string,
	description string,
	sourceEntries []typechanges.ChangeEntryInput,
	targetEntries []typechanges.ChangeEntryInput) (exchange typeexchanges.Exchange, err error) {
	resolvedSource, err := s.ChangesService.ResolveChangeEntries(ctx, partyId, sourceEntries)

	if err != nil {
		return
	}

	resolvedTarget, err := s.ChangesService.ResolveChangeEntries(ctx, partyId, targetEntries)

	if err != nil {
		return
	}

	created, err := s.Database.CreateExchangeCommand(
		ctx,
		partyId,
		name,
		description,
		typechanges.Change{Entries: resolvedSource},
		typechanges.Change{Entries: resolvedTarget})

	if err != nil {
		return
	}

	exchange = typeexchanges.Exchange{
		Id:          created.Id,
		PartyId:     partyId,
		Name:        created.Name,
		Description: created.Description,
	}

	return
}

func (s *Service) RemoveExchange(ctx context.Context, partyId int, name string) (err error) {
	exchange, err := s.exchangeByName(ctx, partyId, name)

	if err != nil {
		return
	}

	return s.Database.RemoveExchangeCommand(ctx, partyId, exchange.Id)
}

// exchangeByName resolves an exchange from the name the API addresses it by.
func (s *Service) exchangeByName(ctx context.Context, partyId int, name string) (exchange typeexchanges.Exchange, err error) {
	exchanges, err := s.Database.GetActualExchangesCommand(ctx, partyId)

	if err != nil {
		return
	}

	for _, candidate := range exchanges {
		if candidate.Name == name {
			return candidate, nil
		}
	}

	err = common.NewExchangeNotFoundError(name)

	return
}

// UseExchange spends an exchange's source change and grants its target change. Without target
// logins both sides fall on the user making the exchange. With them, the source change is what the
// named users lose and the target change is what the user making the exchange gains — this is how a
// mechanic that takes something from another player is expressed.
func (s *Service) UseExchange(ctx context.Context, userId int, partyId int, exchangeName string, targetUserIds []int) (err error) {
	exchange, err := s.exchangeByName(ctx, partyId, exchangeName)

	if err != nil {
		return
	}

	withChanges, err := s.Database.GetExchangeWithEntriesCommand(ctx, partyId, exchange.Id)

	if err != nil {
		return
	}

	payingUserIds := targetUserIds
	if len(payingUserIds) == 0 {
		payingUserIds = []int{userId}
	}

	err = s.requireAffordable(ctx, partyId, payingUserIds, withChanges.SourceChange.Entries)

	if err != nil {
		return
	}

	history, err := s.Database.CreateExchangeHistoryCommand(ctx, userId, partyId, exchange.Id, userId, nil)

	if err != nil {
		return
	}

	err = s.applyEntriesTo(ctx, partyId, payingUserIds, withChanges.SourceChange.Entries, userId, history.Id)

	if err != nil {
		return
	}

	return s.applyEntriesTo(ctx, partyId, []int{userId}, withChanges.TargetChange.Entries, userId, history.Id)
}

// requireAffordable rejects an exchange whose cost the payer cannot cover. Point changes clamp to
// the point type's minimum, so without this check an unaffordable cost would quietly take less than
// it should instead of failing.
func (s *Service) requireAffordable(ctx context.Context, partyId int, payingUserIds []int, sourceEntries []typechanges.ChangeEntry) (err error) {
	for _, entry := range sourceEntries {
		if entry.PointTypeId == nil || entry.Amount >= 0 {
			continue
		}

		var pointType typepoints.PointTypeInfo
		pointType, err = s.PointsService.GetPointTypeById(ctx, partyId, *entry.PointTypeId)

		if err != nil {
			return
		}

		for _, payingUserId := range payingUserIds {
			var value int
			value, err = s.PointsService.GetPointValueByTypeName(ctx, payingUserId, partyId, pointType.Name)

			if err != nil {
				return
			}

			if pointType.Minimum != nil && value+entry.Amount < *pointType.Minimum {
				return common.NewNotEnoughPointsConflictError(pointType.Name, -entry.Amount)
			}
		}
	}

	return
}

// applyEntriesTo stamps a change template onto each of the given users and applies it.
func (s *Service) applyEntriesTo(ctx context.Context, partyId int, targetUserIds []int, templateEntries []typechanges.ChangeEntry, actorUserId int, sourceEventId int) (err error) {
	if len(templateEntries) == 0 {
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

	return s.ChangesService.ApplyChangeEntries(ctx, partyId, userChange.Entries, actorUserId, sourceEventId)
}
