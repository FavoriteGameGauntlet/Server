package srvitems

import (
	"FGG-Service/src/changes/database"
	srvchanges "FGG-Service/src/changes/service"
	"FGG-Service/src/changes/types"
	"FGG-Service/src/common"
	"FGG-Service/src/items/database"
	"FGG-Service/src/items/types"
	"database/sql"
	"errors"
)

type IService interface {
	GetItems(partyId int) ([]typeitems.Item, error)
	GetRemovedItems(partyId int) ([]typeitems.Item, error)
	CreateItem(partyId int, name string, description string, useCount int, entries []typechanges.ChangeEntryInput) (typeitems.Item, error)
	RemoveItem(partyId int, name string) error
	GetUserItems(userId int, partyId int) ([]typeitems.UserItemDetail, error)
	UseItem(userId int, partyId int, itemName string) error
	DiscardUserItem(userId int, partyId int, itemName string) error
	GetItemHistory(userId int, partyId int) ([]typeitems.ItemHistoryDetail, error)
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

func (s *Service) GetItems(partyId int) (items []typeitems.Item, err error) {
	return s.Database.GetActualItemsCommand(partyId)
}

func (s *Service) GetRemovedItems(partyId int) (items []typeitems.Item, err error) {
	return s.Database.GetRemovedItemsCommand(partyId)
}

// CreateItem adds an item to the party catalogue. The entries describe what using it grants. Names
// identify items across the API, so a name already taken by an actual item is rejected.
func (s *Service) CreateItem(
	partyId int,
	name string,
	description string,
	useCount int,
	entries []typechanges.ChangeEntryInput) (item typeitems.Item, err error) {
	actual, err := s.Database.GetActualItemsCommand(partyId)

	if err != nil {
		return
	}

	for _, candidate := range actual {
		if candidate.Name == name {
			err = common.NewItemAlreadyExistsConflictError(name)
			return
		}
	}

	resolved, err := s.ChangesService.ResolveChangeEntries(partyId, entries)

	if err != nil {
		return
	}

	created, err := s.Database.CreateItemCommand(partyId, name, description, useCount, typechanges.Change{Entries: resolved})

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

func (s *Service) RemoveItem(partyId int, name string) (err error) {
	item, err := s.itemByName(partyId, name)

	if err != nil {
		return
	}

	return s.Database.RemoveItemCommand(partyId, item.Id)
}

// itemByName resolves an item from the name the API addresses it by.
func (s *Service) itemByName(partyId int, name string) (item typeitems.Item, err error) {
	items, err := s.Database.GetActualItemsCommand(partyId)

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
// GetUserItems lists what the user holds, named from the party catalogue.
func (s *Service) GetUserItems(userId int, partyId int) (details []typeitems.UserItemDetail, err error) {
	userItems, err := s.Database.GetUserItemsCommand(userId, partyId)

	if err != nil {
		return
	}

	itemsById, err := s.itemsById(partyId)

	if err != nil {
		return
	}

	details = make([]typeitems.UserItemDetail, 0, len(userItems))

	for _, userItem := range userItems {
		item := itemsById[userItem.ItemId]

		details = append(details, typeitems.UserItemDetail{
			Name:         item.Name,
			Description:  item.Description,
			UsesLeft:     userItem.UsesLeft,
			ReceivedDate: userItem.ReceivedDate,
		})
	}

	return
}

// UseItem spends one use of an item the user holds and grants what it carries. The history event of
// spending the use is the event the resulting grants are attributed to.
func (s *Service) UseItem(userId int, partyId int, itemName string) (err error) {
	item, err := s.itemByName(partyId, itemName)

	if err != nil {
		return
	}

	userItem, err := s.Database.GetUserItemCommand(userId, partyId, item.Id)

	if errors.Is(err, sql.ErrNoRows) {
		return common.NewItemNotOwnedConflictError(itemName)
	}

	if err != nil {
		return
	}

	if userItem.UsesLeft < 1 {
		return common.NewItemUsedUpConflictError(itemName)
	}

	withChange, err := s.Database.GetItemCommand(partyId, item.Id)

	if err != nil {
		return
	}

	usesLeft := userItem.UsesLeft - 1

	historyEventId, err := s.Database.ChangeUserItemUsesLeftCommand(userId, partyId, item.Id, usesLeft, userId, nil)

	if err != nil {
		return
	}

	if usesLeft == 0 {
		err = s.Database.DeleteUserItemCommand(userId, partyId, item.Id, userId, &historyEventId)

		if err != nil {
			return
		}
	}

	return s.applyItemChange(userId, partyId, withChange.Change.Entries, historyEventId)
}

// applyItemChange stamps the item's change template onto the user and applies it, the same way a
// rolled wheel row is applied.
func (s *Service) applyItemChange(userId int, partyId int, templateEntries []typechanges.ChangeEntry, sourceEventId int) (err error) {
	if len(templateEntries) == 0 {
		return
	}

	targetedEntries := make([]typechanges.ChangeEntry, 0, len(templateEntries))

	for _, entry := range templateEntries {
		entry.EntryId = nil
		entry.UserId = &userId
		targetedEntries = append(targetedEntries, entry)
	}

	userChange, err := s.ChangesDatabase.CreateUserChangeFromJsonbCommand(partyId, targetedEntries)

	if err != nil {
		return
	}

	return s.ChangesService.ApplyChangeEntries(partyId, userChange.Entries, userId, sourceEventId)
}

// DiscardUserItem drops an item the user holds without using it.
func (s *Service) DiscardUserItem(userId int, partyId int, itemName string) (err error) {
	item, err := s.itemByName(partyId, itemName)

	if err != nil {
		return
	}

	_, err = s.Database.GetUserItemCommand(userId, partyId, item.Id)

	if errors.Is(err, sql.ErrNoRows) {
		return common.NewItemNotOwnedConflictError(itemName)
	}

	if err != nil {
		return
	}

	return s.Database.DeleteUserItemCommand(userId, partyId, item.Id, userId, nil)
}

// GetItemHistory lists the recorded item events of a user, named from the party catalogue.
func (s *Service) GetItemHistory(userId int, partyId int) (history []typeitems.ItemHistoryDetail, err error) {
	entries, err := s.Database.GetItemHistoryCommand(userId, partyId)

	if err != nil {
		return
	}

	itemsById, err := s.itemsById(partyId)

	if err != nil {
		return
	}

	history = make([]typeitems.ItemHistoryDetail, 0, len(entries))

	for _, entry := range entries {
		history = append(history, typeitems.ItemHistoryDetail{
			Name:        itemsById[entry.ItemId].Name,
			Action:      entry.Action,
			UsesLeft:    entry.UsesLeft,
			CreatedDate: entry.CreatedDate,
		})
	}

	return
}

// itemsById indexes the party catalogue, so reads that name many items resolve them in one query.
func (s *Service) itemsById(partyId int) (itemsById map[int]typeitems.Item, err error) {
	items, err := s.Database.GetActualItemsCommand(partyId)

	if err != nil {
		return
	}

	removed, err := s.Database.GetRemovedItemsCommand(partyId)

	if err != nil {
		return
	}

	itemsById = make(map[int]typeitems.Item, len(items)+len(removed))

	for _, item := range append(items, removed...) {
		itemsById[item.Id] = item
	}

	return
}