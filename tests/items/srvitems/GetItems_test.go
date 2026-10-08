package srvitems_test

import (
	"FGG-Service/src/changes/typechanges"
	"FGG-Service/src/items/srvitems"
	"FGG-Service/src/items/typeitems"
	"FGG-Service/tests/changes/dbchangesmock"
	"FGG-Service/tests/changes/srvchangesmock"
	"FGG-Service/tests/items/dbitemsmock"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var potionWithChange = typeitems.Item{Id: 4, PartyId: 1, Name: "potion", Description: "heals", UseCount: 3, ChangeId: 11}

var elixirWithChange = typeitems.Item{Id: 5, PartyId: 1, Name: "elixir", Description: "revives", UseCount: 1, ChangeId: 12}

var potionEntries = []typechanges.NamedChangeEntry{
	{PointTypeName: ptr("health"), Amount: 5},
	{PerkName: ptr("tough"), Amount: 1},
}

// Every item of the catalogue comes with what using it grants, named the way the API addresses it.
func TestSrvItems_GetItems(test *testing.T) {
	test.Run("Success_EntriesAttachedPerItem", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)

		itemsDb.On("GetActualItemsCommand", 1).Return([]typeitems.Item{potionWithChange, elixirWithChange}, nil)
		changesDb.On("GetNamedChangeEntriesCommand", 1, potionWithChange.ChangeId).Return(potionEntries, nil)
		changesDb.On("GetNamedChangeEntriesCommand", 1, elixirWithChange.ChangeId).Return([]typechanges.NamedChangeEntry{}, nil)

		sut := srvitems.Service{Database: itemsDb, ChangesDatabase: changesDb}

		items, err := sut.GetItems(test.Context(), 1)

		require.NoError(test, err)
		require.Equal(test, []typeitems.ItemWithEntries{
			{Item: potionWithChange, Entries: potionEntries},
			{Item: elixirWithChange, Entries: []typechanges.NamedChangeEntry{}},
		}, items)
		changesDb.AssertExpectations(test)
	})

	test.Run("EntriesRead_DatabaseError", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)

		itemsDb.On("GetActualItemsCommand", 1).Return([]typeitems.Item{potionWithChange}, nil)
		changesDb.On("GetNamedChangeEntriesCommand", 1, potionWithChange.ChangeId).Return([]typechanges.NamedChangeEntry{}, dbError)

		sut := srvitems.Service{Database: itemsDb, ChangesDatabase: changesDb}

		_, err := sut.GetItems(test.Context(), 1)

		require.ErrorIs(test, err, dbError)
	})

	test.Run("CatalogueRead_DatabaseError", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)

		itemsDb.On("GetActualItemsCommand", 1).Return([]typeitems.Item{}, dbError)

		sut := srvitems.Service{Database: itemsDb, ChangesDatabase: changesDb}

		_, err := sut.GetItems(test.Context(), 1)

		require.ErrorIs(test, err, dbError)
		changesDb.AssertNotCalled(test, "GetNamedChangeEntriesCommand")
	})
}

// Items removed from the catalogue keep reporting what they grant.
func TestSrvItems_GetRemovedItems(test *testing.T) {
	test.Run("Success_EntriesAttachedPerItem", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)

		itemsDb.On("GetRemovedItemsCommand", 1).Return([]typeitems.Item{potionWithChange}, nil)
		changesDb.On("GetNamedChangeEntriesCommand", 1, potionWithChange.ChangeId).Return(potionEntries, nil)

		sut := srvitems.Service{Database: itemsDb, ChangesDatabase: changesDb}

		items, err := sut.GetRemovedItems(test.Context(), 1)

		require.NoError(test, err)
		require.Equal(test, []typeitems.ItemWithEntries{{Item: potionWithChange, Entries: potionEntries}}, items)
	})
}

// A created item answers with the entries the catalogue stored for it.
func TestSrvItems_CreateItem(test *testing.T) {
	test.Run("Success_ReturnsStoredEntries", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)
		changesSvc := new(srvchangesmock.ServiceMock)

		resolved := []typechanges.ChangeEntry{{Amount: 5, PointTypeId: ptr(2)}, {Amount: 1, PerkId: ptr(3)}}
		change := typechanges.Change{ChangeId: ptr(11), Entries: resolved}

		itemsDb.On("GetActualItemsCommand", 1).Return([]typeitems.Item{}, nil)
		changesSvc.On("ResolveChangeEntries", 1, potionEntries).Return(resolved, nil)
		itemsDb.On("CreateItemCommand", 1, "potion", "heals", 3, typechanges.Change{Entries: resolved}).
			Return(typeitems.ItemWithChange{Id: 4, PartyId: 1, Name: "potion", Description: "heals", UseCount: 3, Change: change}, nil)
		changesDb.On("GetNamedChangeEntriesCommand", 1, 11).Return(potionEntries, nil)

		sut := srvitems.Service{Database: itemsDb, ChangesDatabase: changesDb, ChangesService: changesSvc}

		item, err := sut.CreateItem(test.Context(), 1, "potion", "heals", 3, potionEntries)

		require.NoError(test, err)
		require.Equal(test, typeitems.ItemWithEntries{Item: potionWithChange, Entries: potionEntries}, item)
		changesDb.AssertExpectations(test)
	})

	test.Run("EntriesRead_DatabaseError", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)
		changesSvc := new(srvchangesmock.ServiceMock)

		resolved := []typechanges.ChangeEntry{{Amount: 5, PointTypeId: ptr(2)}}

		itemsDb.On("GetActualItemsCommand", 1).Return([]typeitems.Item{}, nil)
		changesSvc.On("ResolveChangeEntries", 1, potionEntries).Return(resolved, nil)
		itemsDb.On("CreateItemCommand", 1, "potion", "heals", 3, typechanges.Change{Entries: resolved}).
			Return(typeitems.ItemWithChange{Id: 4, PartyId: 1, Change: typechanges.Change{ChangeId: ptr(11)}}, nil)
		changesDb.On("GetNamedChangeEntriesCommand", 1, 11).Return([]typechanges.NamedChangeEntry{}, dbError)

		sut := srvitems.Service{Database: itemsDb, ChangesDatabase: changesDb, ChangesService: changesSvc}

		_, err := sut.CreateItem(test.Context(), 1, "potion", "heals", 3, potionEntries)

		require.ErrorIs(test, err, dbError)
	})
}

// The items a user holds come with what using them grants.
func TestSrvItems_GetUserItems(test *testing.T) {
	received := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)

	test.Run("Success_EntriesAttachedPerItem", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)

		held := typeitems.UserItemDetail{Name: "potion", Description: "heals", UsesLeft: 2, ReceivedDate: received, ChangeId: 11}

		itemsDb.On("GetUserItemsCommand", 7, 1).Return([]typeitems.UserItemDetail{held}, nil)
		changesDb.On("GetNamedChangeEntriesCommand", 1, 11).Return(potionEntries, nil)

		sut := srvitems.Service{Database: itemsDb, ChangesDatabase: changesDb}

		details, err := sut.GetUserItems(test.Context(), 7, 1)

		held.Entries = potionEntries

		require.NoError(test, err)
		require.Equal(test, []typeitems.UserItemDetail{held}, details)
	})

	test.Run("EntriesRead_DatabaseError", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)

		itemsDb.On("GetUserItemsCommand", 7, 1).Return([]typeitems.UserItemDetail{{ChangeId: 11}}, nil)
		changesDb.On("GetNamedChangeEntriesCommand", 1, 11).Return([]typechanges.NamedChangeEntry{}, dbError)

		sut := srvitems.Service{Database: itemsDb, ChangesDatabase: changesDb}

		_, err := sut.GetUserItems(test.Context(), 7, 1)

		require.ErrorIs(test, err, dbError)
	})
}
