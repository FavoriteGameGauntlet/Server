package srvitems_test

import (
	typechanges "FGG-Service/src/changes/types"
	srvitems "FGG-Service/src/items/service"
	typeitems "FGG-Service/src/items/types"
	dbchangesmock "FGG-Service/tests/changes/mock"
	srvchangesmock "FGG-Service/tests/changes/srvmock"
	dbitemsmock "FGG-Service/tests/items/mock"
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
		itemsDb.On("ChangeUserItemUsesLeftCommand", 7, 1, potion.Id, 1, 7, (*int)(nil)).Return(55, nil)
		changesDb.On("CreateUserChangeFromJsonbCommand", 1, []typechanges.ChangeEntry{targetedEntry}).
			Return(typechanges.UserChange{Entries: []typechanges.ChangeEntry{targetedEntry}}, nil)
		changesSvc.On("ApplyChangeEntries", 1, []typechanges.ChangeEntry{targetedEntry}, 7, 55).Return(nil)

		sut := srvitems.Service{Database: itemsDb, ChangesDatabase: changesDb, ChangesService: changesSvc}

		err := sut.UseItem(7, 1, potion.Name)

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
		itemsDb.On("ChangeUserItemUsesLeftCommand", 7, 1, potion.Id, 0, 7, (*int)(nil)).Return(55, nil)
		itemsDb.On("DeleteUserItemCommand", 7, 1, potion.Id, 7, ptr(55)).Return(nil)

		sut := srvitems.Service{Database: itemsDb}

		err := sut.UseItem(7, 1, potion.Name)

		require.NoError(test, err)
		itemsDb.AssertExpectations(test)
	})

	test.Run("NoUsesLeft_Rejected", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		itemsDb.On("GetActualItemsCommand", 1).Return([]typeitems.Item{potion}, nil)
		itemsDb.On("GetUserItemCommand", 7, 1, potion.Id).Return(typeitems.UserItem{UsesLeft: 0}, nil)

		sut := srvitems.Service{Database: itemsDb}

		err := sut.UseItem(7, 1, potion.Name)

		require.Error(test, err)
		itemsDb.AssertNotCalled(test, "ChangeUserItemUsesLeftCommand")
	})

	test.Run("NotOwned_Rejected", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		itemsDb.On("GetActualItemsCommand", 1).Return([]typeitems.Item{potion}, nil)
		itemsDb.On("GetUserItemCommand", 7, 1, potion.Id).Return(typeitems.UserItem{}, sql.ErrNoRows)

		sut := srvitems.Service{Database: itemsDb}

		err := sut.UseItem(7, 1, potion.Name)

		require.Error(test, err)
		itemsDb.AssertNotCalled(test, "ChangeUserItemUsesLeftCommand")
	})

	test.Run("UnknownItem_Rejected", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		itemsDb.On("GetActualItemsCommand", 1).Return([]typeitems.Item{}, nil)

		sut := srvitems.Service{Database: itemsDb}

		err := sut.UseItem(7, 1, "unknown")

		require.Error(test, err)
		itemsDb.AssertNotCalled(test, "GetUserItemCommand")
	})

	test.Run("CatalogueRead_DatabaseError", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		itemsDb.On("GetActualItemsCommand", 1).Return([]typeitems.Item{}, dbError)

		sut := srvitems.Service{Database: itemsDb}

		err := sut.UseItem(7, 1, potion.Name)

		require.ErrorIs(test, err, dbError)
	})
}