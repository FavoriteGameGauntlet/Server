package srvgrants

import (
	"FGG-Service/src/changes/database"
	srvchanges "FGG-Service/src/changes/service"
	"FGG-Service/src/changes/types"
	"FGG-Service/src/grants/database"
)

type IService interface {
	GrantToUsers(actorUserId int, partyId int, targetUserIds []int, inputs []typechanges.ChangeEntryInput) error
}

type Service struct {
	Database        dbgrants.IDatabase
	ChangesDatabase dbchanges.IDatabase
	ChangesService  srvchanges.IService
}

func NewService() *Service {
	return &Service{
		Database:        new(dbgrants.Database),
		ChangesDatabase: new(dbchanges.Database),
		ChangesService:  srvchanges.NewService(),
	}
}

// GrantToUsers gives each named user what the entries describe, on an administrator's say-so rather
// than as the outcome of a game action. Each user gets their own manual history entry, whose id is
// the source event their grants are recorded against, so a grant can be traced back to the
// administrator who made it.
func (s *Service) GrantToUsers(actorUserId int, partyId int, targetUserIds []int, inputs []typechanges.ChangeEntryInput) (err error) {
	resolved, err := s.ChangesService.ResolveChangeEntries(partyId, inputs)

	if err != nil {
		return
	}

	for _, targetUserId := range targetUserIds {
		targetUserId := targetUserId

		targetedEntries := make([]typechanges.ChangeEntry, 0, len(resolved))

		for _, entry := range resolved {
			entry.EntryId = nil
			entry.UserId = &targetUserId
			targetedEntries = append(targetedEntries, entry)
		}

		err = s.grantToUser(actorUserId, partyId, targetedEntries)

		if err != nil {
			return
		}
	}

	return
}

// grantToUser records one user's grant and applies it. The manual history is written per user so
// that the entries behind each history row are only that user's.
func (s *Service) grantToUser(actorUserId int, partyId int, targetedEntries []typechanges.ChangeEntry) (err error) {
	if len(targetedEntries) == 0 {
		return
	}

	created, err := s.Database.CreateManualHistoryCommand(partyId, actorUserId, targetedEntries, nil)

	if err != nil {
		return
	}

	for _, entry := range created {
		var persistedEntries []typechanges.ChangeEntry
		persistedEntries, err = s.ChangesDatabase.GetChangeEntriesJsonbCommand(partyId, entry.ChangeId)

		if err != nil {
			return
		}

		err = s.ChangesService.ApplyChangeEntries(partyId, persistedEntries, actorUserId, entry.Id)

		if err != nil {
			return
		}
	}

	return
}