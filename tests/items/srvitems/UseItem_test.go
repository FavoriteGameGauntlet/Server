package srvitems_test

import (
	"FGG-Service/src/changes/typechanges"
	"FGG-Service/src/items/srvitems"
	"FGG-Service/src/items/typeitems"
	"FGG-Service/tests/changes/dbchangesmock"
	"FGG-Service/tests/changes/srvchangesmock"
	"FGG-Service/tests/items/dbitemsmock"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

var dbError = errors.New("database connection lost")

var potion = typeitems.Item{Id: 4, PartyId: 1, Name: "potion", Description: "heals", UseCount: 3}

func ptr[T any](value T) *T {
	return &value
}

// Using an item spends one use and grants what the item carries. The history event of spending the
// use is the source event of the grants, so they can be traced back to the use that caused them.
// The item belongs to user 7 and is used by user 9: the grants go to the holder, but everything
// recorded names the user who used it.
func TestSrvItems_UseItem(test *testing.T) {
	test.Run("Success_SpendsUseThenGrants", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)
		changesSvc := new(srvchangesmock.ServiceMock)

		templateEntry := typechanges.ChangeEntry{EntryId: ptr(9), Amount: 5, PointTypeId: ptr(2)}
		targetedEntry := typechanges.ChangeEntry{Amount: 5, PointTypeId: ptr(2), UserId: ptr(7)}

		itemsDb.On("GetActualItemsCommand", 1).Return([]typeitems.Item{potion}, nil)
		itemsDb.On("GetUserItemCommand", 7, 1, potion.Id).
			Return(typeitems.UserItem{Id: 1, UserId: 7, PartyId: 1, ItemId: potion.Id, UsesLeft: 2}, nil)
		itemsDb.On("GetItemCommand", 1, potion.Id).
			Return(typeitems.ItemWithChange{Id: potion.Id, Change: typechanges.Change{Entries: []typechanges.ChangeEntry{templateEntry}}}, nil)
		itemsDb.On("ChangeUserItemUsesLeftCommand", 7, 1, potion.Id, 1, 9, (*int)(nil)).Return(55, nil)
		changesDb.On("CreateUserChangeFromJsonbCommand", 1, []typechanges.ChangeEntry{targetedEntry}).
			Return(typechanges.UserChange{Entries: []typechanges.ChangeEntry{targetedEntry}}, nil)
		changesSvc.On("ApplyChangeEntries", 1, []typechanges.ChangeEntry{targetedEntry}, 9, 55).Return(nil)

		sut := srvitems.Service{Database: itemsDb, ChangesDatabase: changesDb, ChangesService: changesSvc}

		err := sut.UseItem(test.Context(), 9, 7, 1, potion.Name)

		require.NoError(test, err)
		itemsDb.AssertExpectations(test)
		itemsDb.AssertNotCalled(test, "DeleteUserItemCommand")
		changesDb.AssertExpectations(test)
		changesSvc.AssertExpectations(test)
	})

	test.Run("LastUse_RemovesItemAttributedToUse", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)

		itemsDb.On("GetActualItemsCommand", 1).Return([]typeitems.Item{potion}, nil)
		itemsDb.On("GetUserItemCommand", 7, 1, potion.Id).
			Return(typeitems.UserItem{Id: 1, UserId: 7, PartyId: 1, ItemId: potion.Id, UsesLeft: 1}, nil)
		itemsDb.On("GetItemCommand", 1, potion.Id).Return(typeitems.ItemWithChange{Id: potion.Id}, nil)
		itemsDb.On("ChangeUserItemUsesLeftCommand", 7, 1, potion.Id, 0, 9, (*int)(nil)).Return(55, nil)
		itemsDb.On("DeleteUserItemCommand", 7, 1, potion.Id, 9, ptr(55)).Return(nil)

		sut := srvitems.Service{Database: itemsDb}

		err := sut.UseItem(test.Context(), 9, 7, 1, potion.Name)

		require.NoError(test, err)
		itemsDb.AssertExpectations(test)
	})

	test.Run("NoUsesLeft_Rejected", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		itemsDb.On("GetActualItemsCommand", 1).Return([]typeitems.Item{potion}, nil)
		itemsDb.On("GetUserItemCommand", 7, 1, potion.Id).Return(typeitems.UserItem{UsesLeft: 0}, nil)

		sut := srvitems.Service{Database: itemsDb}

		err := sut.UseItem(test.Context(), 9, 7, 1, potion.Name)

		require.Error(test, err)
		itemsDb.AssertNotCalled(test, "ChangeUserItemUsesLeftCommand")
	})

	test.Run("NotOwned_Rejected", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		itemsDb.On("GetActualItemsCommand", 1).Return([]typeitems.Item{potion}, nil)
		itemsDb.On("GetUserItemCommand", 7, 1, potion.Id).Return(typeitems.UserItem{}, sql.ErrNoRows)

		sut := srvitems.Service{Database: itemsDb}

		err := sut.UseItem(test.Context(), 9, 7, 1, potion.Name)

		require.Error(test, err)
		itemsDb.AssertNotCalled(test, "ChangeUserItemUsesLeftCommand")
	})

	test.Run("UnknownItem_Rejected", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		itemsDb.On("GetActualItemsCommand", 1).Return([]typeitems.Item{}, nil)

		sut := srvitems.Service{Database: itemsDb}

		err := sut.UseItem(test.Context(), 9, 7, 1, "unknown")

		require.Error(test, err)
		itemsDb.AssertNotCalled(test, "GetUserItemCommand")
	})

	test.Run("CatalogueRead_DatabaseError", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		itemsDb.On("GetActualItemsCommand", 1).Return([]typeitems.Item{}, dbError)

		sut := srvitems.Service{Database: itemsDb}

		err := sut.UseItem(test.Context(), 9, 7, 1, potion.Name)

		require.ErrorIs(test, err, dbError)
	})
}

// Discarding an item removes it from its holder, and the removal is recorded as the discarding
// user's action.
func TestSrvItems_DiscardUserItem(test *testing.T) {
	test.Run("Success_RecordedAsActor", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)

		itemsDb.On("GetActualItemsCommand", 1).Return([]typeitems.Item{potion}, nil)
		itemsDb.On("GetUserItemCommand", 7, 1, potion.Id).
			Return(typeitems.UserItem{Id: 1, UserId: 7, PartyId: 1, ItemId: potion.Id, UsesLeft: 2}, nil)
		itemsDb.On("DeleteUserItemCommand", 7, 1, potion.Id, 9, (*int)(nil)).Return(nil)

		sut := srvitems.Service{Database: itemsDb}

		err := sut.DiscardUserItem(test.Context(), 9, 7, 1, potion.Name)

		require.NoError(test, err)
		itemsDb.AssertExpectations(test)
	})
}
