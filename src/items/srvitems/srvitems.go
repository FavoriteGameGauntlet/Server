package srvitems

import (
	"FGG-Service/src/changes/dbchanges"
	"FGG-Service/src/changes/srvchanges"
	"FGG-Service/src/changes/typechanges"
	"FGG-Service/src/common"
	"FGG-Service/src/items/dbitems"
	"FGG-Service/src/items/typeitems"
	"context"
	"database/sql"
	"errors"
	"slices"
)

type IService interface {
	GetItems(ctx context.Context, partyId int) ([]typeitems.ItemWithEntries, error)
	GetRemovedItems(ctx context.Context, partyId int) ([]typeitems.ItemWithEntries, error)
	CreateItem(ctx context.Context, partyId int, name string, description string, useCount int, entries []typechanges.NamedChangeEntry) (typeitems.ItemWithEntries, error)
	RemoveItem(ctx context.Context, partyId int, itemId int) error
	GetUserItems(ctx context.Context, userId int, partyId int) ([]typeitems.UserItemDetail, error)
	UseItem(ctx context.Context, actorUserId int, userId int, partyId int, userItemId int) error
	DiscardUserItem(ctx context.Context, actorUserId int, userId int, partyId int, userItemId int) error
	GetItemHistory(ctx context.Context, userId int, partyId int) ([]typeitems.ItemHistory, error)
}

type Service struct {
	Database        dbitems.IDatabase
	ChangesDatabase dbchanges.IDatabase
	ChangesService  srvchanges.IService
}

func NewService() *Service {
	return &Service{
		Database:        new(dbitems.Database),
		ChangesDatabase: new(dbchanges.Database),
		ChangesService:  srvchanges.NewService(),
	}
}

func (s *Service) GetItems(ctx context.Context, partyId int) (items []typeitems.ItemWithEntries, err error) {
	actual, err := s.Database.GetActualItemsCommand(ctx, partyId)

	if err != nil {
		return
	}

	return s.withEntries(ctx, partyId, actual)
}

func (s *Service) GetRemovedItems(ctx context.Context, partyId int) (items []typeitems.ItemWithEntries, err error) {
	removed, err := s.Database.GetRemovedItemsCommand(ctx, partyId)

	if err != nil {
		return
	}

	return s.withEntries(ctx, partyId, removed)
}

// withEntries attaches to each item what using it grants.
func (s *Service) withEntries(ctx context.Context, partyId int, items []typeitems.Item) (detailed []typeitems.ItemWithEntries, err error) {
	detailed = make([]typeitems.ItemWithEntries, len(items))

	for i, item := range items {
		detailed[i].Item = item
		detailed[i].Entries, err = s.ChangesDatabase.GetNamedChangeEntriesCommand(ctx, partyId, item.ChangeId)

		if err != nil {
			return nil, err
		}
	}

	return
}

// CreateItem adds an item to the party catalogue. The entries describe what using it grants. A name
// already taken by an actual item is rejected.
func (s *Service) CreateItem(
	ctx context.Context,
	partyId int,
	name string,
	description string,
	useCount int,
	entries []typechanges.NamedChangeEntry) (item typeitems.ItemWithEntries, err error) {
	actual, err := s.Database.GetActualItemsCommand(ctx, partyId)

	if err != nil {
		return
	}

	for _, candidate := range actual {
		if candidate.Name == name {
			err = common.NewItemAlreadyExistsConflictError(name)
			return
		}
	}

	resolved, err := s.ChangesService.ResolveChangeEntries(ctx, partyId, entries)

	if err != nil {
		return
	}

	created, err := s.Database.CreateItemCommand(ctx, partyId, name, description, useCount, typechanges.Change{Entries: resolved})

	if err != nil {
		return
	}

	item.Item = typeitems.Item{
		Id:          created.Id,
		PartyId:     created.PartyId,
		Name:        created.Name,
		Description: created.Description,
		UseCount:    created.UseCount,
		ChangeId:    *created.Change.ChangeId,
	}

	item.Entries, err = s.ChangesDatabase.GetNamedChangeEntriesCommand(ctx, partyId, item.ChangeId)

	return
}

// RemoveItem removes an item from the party catalogue. Copies users already hold stay usable.
func (s *Service) RemoveItem(ctx context.Context, partyId int, itemId int) (err error) {
	items, err := s.Database.GetActualItemsCommand(ctx, partyId)

	if err != nil {
		return
	}

	isActual := slices.ContainsFunc(items, func(item typeitems.Item) bool {
		return item.Id == itemId
	})

	if !isActual {
		return common.NewItemNotFoundError(itemId)
	}

	return s.Database.RemoveItemCommand(ctx, partyId, itemId)
}

// GetUserItems lists what the user holds, each item along with what using it grants.
func (s *Service) GetUserItems(ctx context.Context, userId int, partyId int) (details []typeitems.UserItemDetail, err error) {
	details, err = s.Database.GetUserItemsCommand(ctx, userId, partyId)

	if err != nil {
		return
	}

	for i := range details {
		details[i].Entries, err = s.ChangesDatabase.GetNamedChangeEntriesCommand(ctx, partyId, details[i].ChangeId)

		if err != nil {
			return nil, err
		}
	}

	return
}

// UseItem spends one use of a copy of an item the user holds and grants what it carries to that user.
// The history event of spending the use is the event the resulting grants are attributed to. Another
// member may use the item, so everything recorded names the actor, not the holder. The copy stays
// usable after its item is removed from the catalogue.
func (s *Service) UseItem(ctx context.Context, actorUserId int, userId int, partyId int, userItemId int) (err error) {
	userItem, err := s.Database.GetUserItemCommand(ctx, userId, partyId, userItemId)

	if errors.Is(err, sql.ErrNoRows) {
		return common.NewUserItemNotFoundError(userItemId)
	}

	if err != nil {
		return
	}

	if userItem.UsesLeft < 1 {
		return common.NewItemUsedUpConflictError(userItemId)
	}

	templateEntries, err := s.ChangesDatabase.GetChangeEntriesJsonbCommand(ctx, partyId, userItem.ChangeId)

	if err != nil {
		return
	}

	usesLeft := userItem.UsesLeft - 1

	historyEventId, err := s.Database.ChangeUserItemUsesLeftCommand(ctx, userId, partyId, userItemId, usesLeft, actorUserId, nil)

	if err != nil {
		return
	}

	if usesLeft == 0 {
		err = s.Database.DeleteUserItemCommand(ctx, userId, partyId, userItemId, actorUserId, &historyEventId)

		if err != nil {
			return
		}
	}

	return s.applyItemChange(ctx, actorUserId, userId, partyId, templateEntries, historyEventId)
}

// applyItemChange stamps the item's change template onto the user and applies it.
func (s *Service) applyItemChange(ctx context.Context, actorUserId int, userId int, partyId int, templateEntries []typechanges.ChangeEntry, sourceEventId int) (err error) {
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

// DiscardUserItem drops one copy of an item the user holds without using it.
func (s *Service) DiscardUserItem(ctx context.Context, actorUserId int, userId int, partyId int, userItemId int) (err error) {
	_, err = s.Database.GetUserItemCommand(ctx, userId, partyId, userItemId)

	if errors.Is(err, sql.ErrNoRows) {
		return common.NewUserItemNotFoundError(userItemId)
	}

	if err != nil {
		return
	}

	return s.Database.DeleteUserItemCommand(ctx, userId, partyId, userItemId, actorUserId, nil)
}

func (s *Service) GetItemHistory(ctx context.Context, userId int, partyId int) (history []typeitems.ItemHistory, err error) {
	return s.Database.GetItemHistoryCommand(ctx, userId, partyId)
}
