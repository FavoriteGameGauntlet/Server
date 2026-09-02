package srvexchanges

import (
	"FGG-Service/src/changes/database"
	srvchanges "FGG-Service/src/changes/service"
	"FGG-Service/src/changes/types"
	"FGG-Service/src/common"
	"FGG-Service/src/exchanges/database"
	"FGG-Service/src/exchanges/types"
	srvpoints "FGG-Service/src/points/service"
	typepoints "FGG-Service/src/points/type"
)

type IService interface {
	GetExchanges(partyId int) ([]typeexchanges.Exchange, error)
	GetRemovedExchanges(partyId int) ([]typeexchanges.Exchange, error)
	CreateExchange(partyId int, name string, description string, sourceEntries []typechanges.ChangeEntryInput, targetEntries []typechanges.ChangeEntryInput) (typeexchanges.Exchange, error)
	RemoveExchange(partyId int, name string) error
	UseExchange(userId int, partyId int, exchangeName string, targetUserIds []int) error
	GetExchangeHistory(userId int, partyId int) ([]typeexchanges.ExchangeHistory, error)
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

func (s *Service) GetExchanges(partyId int) (exchanges []typeexchanges.Exchange, err error) {
	return s.Database.GetActualExchangesCommand(partyId)
}

func (s *Service) GetRemovedExchanges(partyId int) (exchanges []typeexchanges.Exchange, err error) {
	return s.Database.GetRemovedExchangesCommand(partyId)
}

func (s *Service) GetExchangeHistory(userId int, partyId int) (history []typeexchanges.ExchangeHistory, err error) {
	return s.Database.GetExchangeHistoryCommand(userId, partyId)
}

// CreateExchange adds an exchange to the party catalogue. The source entries are what it costs and
// the target entries are what it gives; both carry the sign they are applied with, so a cost is
// negative.
func (s *Service) CreateExchange(
	partyId int,
	name string,
	description string,
	sourceEntries []typechanges.ChangeEntryInput,
	targetEntries []typechanges.ChangeEntryInput) (exchange typeexchanges.Exchange, err error) {
	resolvedSource, err := s.ChangesService.ResolveChangeEntries(partyId, sourceEntries)

	if err != nil {
		return
	}

	resolvedTarget, err := s.ChangesService.ResolveChangeEntries(partyId, targetEntries)

	if err != nil {
		return
	}

	created, err := s.Database.CreateExchangeCommand(
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

func (s *Service) RemoveExchange(partyId int, name string) (err error) {
	exchange, err := s.exchangeByName(partyId, name)

	if err != nil {
		return
	}

	return s.Database.RemoveExchangeCommand(partyId, exchange.Id)
}

// exchangeByName resolves an exchange from the name the API addresses it by.
func (s *Service) exchangeByName(partyId int, name string) (exchange typeexchanges.Exchange, err error) {
	exchanges, err := s.Database.GetActualExchangesCommand(partyId)

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
func (s *Service) UseExchange(userId int, partyId int, exchangeName string, targetUserIds []int) (err error) {
	exchange, err := s.exchangeByName(partyId, exchangeName)

	if err != nil {
		return
	}

	withChanges, err := s.Database.GetExchangeWithEntriesCommand(partyId, exchange.Id)

	if err != nil {
		return
	}

	payingUserIds := targetUserIds
	if len(payingUserIds) == 0 {
		payingUserIds = []int{userId}
	}

	err = s.requireAffordable(partyId, payingUserIds, withChanges.SourceChange.Entries)

	if err != nil {
		return
	}

	history, err := s.Database.CreateExchangeHistoryCommand(userId, partyId, exchange.Id, nil)

	if err != nil {
		return
	}

	err = s.applyEntriesTo(partyId, payingUserIds, withChanges.SourceChange.Entries, history.Id)

	if err != nil {
		return
	}

	return s.applyEntriesTo(partyId, []int{userId}, withChanges.TargetChange.Entries, history.Id)
}

// requireAffordable rejects an exchange whose cost the payer cannot cover. Point changes clamp to
// the point type's minimum, so without this check an unaffordable cost would quietly take less than
// it should instead of failing.
func (s *Service) requireAffordable(partyId int, payingUserIds []int, sourceEntries []typechanges.ChangeEntry) (err error) {
	for _, entry := range sourceEntries {
		if entry.PointTypeId == nil || entry.Amount >= 0 {
			continue
		}

		var pointType typepoints.PointTypeInfo
		pointType, err = s.PointsService.GetPointTypeById(partyId, *entry.PointTypeId)

		if err != nil {
			return
		}

		for _, payingUserId := range payingUserIds {
			var value int
			value, err = s.PointsService.GetPointValueByTypeName(payingUserId, partyId, pointType.Name)

			if err != nil {
				return
			}

			if value+entry.Amount < pointType.Minimum {
				return common.NewNotEnoughPointsConflictError(pointType.Name, -entry.Amount)
			}
		}
	}

	return
}

// applyEntriesTo stamps a change template onto each of the given users and applies it.
func (s *Service) applyEntriesTo(partyId int, targetUserIds []int, templateEntries []typechanges.ChangeEntry, sourceEventId int) (err error) {
	if len(templateEntries) == 0 {
		return
	}

	targetedEntries := make([]typechanges.ChangeEntry, 0, len(templateEntries)*len(targetUserIds))

	for _, targetUserId := range targetUserIds {
		targetUserId := targetUserId

		for _, entry := range templateEntries {
			entry.EntryId = nil
			entry.UserId = &targetUserId
			targetedEntries = append(targetedEntries, entry)
		}
	}

	userChange, err := s.ChangesDatabase.CreateUserChangeFromJsonbCommand(partyId, targetedEntries)

	if err != nil {
		return
	}

	return s.ChangesService.ApplyChangeEntries(partyId, userChange.Entries, sourceEventId)
}