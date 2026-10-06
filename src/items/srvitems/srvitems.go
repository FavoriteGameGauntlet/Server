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
)

type IService interface {
	GetItems(ctx context.Context, partyId int) ([]typeitems.Item, error)
	GetRemovedItems(ctx context.Context, partyId int) ([]typeitems.Item, error)
	CreateItem(ctx context.Context, partyId int, name string, description string, useCount int, entries []typechanges.ChangeEntryInput) (typeitems.Item, error)
	RemoveItem(ctx context.Context, partyId int, name string) error
	GetUserItems(ctx context.Context, userId int, partyId int) ([]typeitems.UserItemDetail, error)
	UseItem(ctx context.Context, actorUserId int, userId int, partyId int, itemName string) error
	DiscardUserItem(ctx context.Context, actorUserId int, userId int, partyId int, itemName string) error
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

func (s *Service) GetItems(ctx context.Context, partyId int) (items []typeitems.Item, err error) {
	return s.Database.GetActualItemsCommand(ctx, partyId)
}

func (s *Service) GetRemovedItems(ctx context.Context, partyId int) (items []typeitems.Item, err error) {
	return s.Database.GetRemovedItemsCommand(ctx, partyId)
}

// CreateItem adds an item to the party catalogue. The entries describe what using it grants. Names
// identify items across the API, so a name already taken by an actual item is rejected.
func (s *Service) CreateItem(
	ctx context.Context,
	partyId int,
	name string,
	description string,
	useCount int,
	entries []typechanges.ChangeEntryInput) (item typeitems.Item, err error) {
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

	item = typeitems.Item{
		Id:          created.Id,
		PartyId:     created.PartyId,
		Name:        created.Name,
		Description: created.Description,
		UseCount:    created.UseCount,
	}

	return
}

func (s *Service) RemoveItem(ctx context.Context, partyId int, name string) (err error) {
	item, err := s.itemByName(ctx, partyId, name)

	if err != nil {
		return
	}

	return s.Database.RemoveItemCommand(ctx, partyId, item.Id)
}

// itemByName resolves an item from the name the API addresses it by.
func (s *Service) itemByName(ctx context.Context, partyId int, name string) (item typeitems.Item, err error) {
	items, err := s.Database.GetActualItemsCommand(ctx, partyId)

	if err != nil {
		return
	}

	for _, candidate := range items {
		if candidate.Name == name {
			return candidate, nil
		}
	}

	err = common.NewItemNotFoundError(name)

	return
}

// GetUserItems lists what the user holds.
func (s *Service) GetUserItems(ctx context.Context, userId int, partyId int) (details []typeitems.UserItemDetail, err error) {
	return s.Database.GetUserItemsCommand(ctx, userId, partyId)
}

// UseItem spends one use of an item the user holds and grants what it carries to that user. The
// history event of spending the use is the event the resulting grants are attributed to. Another
// member may use the item, so everything recorded names the actor, not the holder.
func (s *Service) UseItem(ctx context.Context, actorUserId int, userId int, partyId int, itemName string) (err error) {
	item, err := s.itemByName(ctx, partyId, itemName)

	if err != nil {
		return
	}

	userItem, err := s.Database.GetUserItemCommand(ctx, userId, partyId, item.Id)

	if errors.Is(err, sql.ErrNoRows) {
		return common.NewItemNotOwnedConflictError(itemName)
	}

	if err != nil {
		return
	}

	if userItem.UsesLeft < 1 {
		return common.NewItemUsedUpConflictError(itemName)
	}

	withChange, err := s.Database.GetItemCommand(ctx, partyId, item.Id)

	if err != nil {
		return
	}

	usesLeft := userItem.UsesLeft - 1

	historyEventId, err := s.Database.ChangeUserItemUsesLeftCommand(ctx, userId, partyId, item.Id, usesLeft, actorUserId, nil)

	if err != nil {
		return
	}

	if usesLeft == 0 {
		err = s.Database.DeleteUserItemCommand(ctx, userId, partyId, item.Id, actorUserId, &historyEventId)

		if err != nil {
			return
		}
	}

	return s.applyItemChange(ctx, actorUserId, userId, partyId, withChange.Change.Entries, historyEventId)
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

// DiscardUserItem drops an item the user holds without using it.
func (s *Service) DiscardUserItem(ctx context.Context, actorUserId int, userId int, partyId int, itemName string) (err error) {
	item, err := s.itemByName(ctx, partyId, itemName)

	if err != nil {
		return
	}

	_, err = s.Database.GetUserItemCommand(ctx, userId, partyId, item.Id)

	if errors.Is(err, sql.ErrNoRows) {
		return common.NewItemNotOwnedConflictError(itemName)
	}

	if err != nil {
		return
	}

	return s.Database.DeleteUserItemCommand(ctx, userId, partyId, item.Id, actorUserId, nil)
}

func (s *Service) GetItemHistory(ctx context.Context, userId int, partyId int) (history []typeitems.ItemHistory, err error) {
	return s.Database.GetItemHistoryCommand(ctx, userId, partyId)
}
