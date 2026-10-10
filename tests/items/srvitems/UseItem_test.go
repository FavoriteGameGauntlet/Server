package srvitems_test

import (
	"FGG-Service/src/changes/typechanges"
	"FGG-Service/src/common"
	"FGG-Service/src/items/srvitems"
	"FGG-Service/src/items/typeitems"
	"FGG-Service/tests/changes/dbchangesmock"
	"FGG-Service/tests/changes/srvchangesmock"
	"FGG-Service/tests/items/dbitemsmock"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var dbError = errors.New("database connection lost")

var potion = typeitems.Item{Id: 4, PartyId: 1, Name: "potion", Description: "heals", UseCount: 3, ChangeId: 30}

func ptr[T any](value T) *T {
	return &value
}

// heldPotion is a copy of the potion held by user 7. The copy has an id of its own, apart from the
// id of the catalogue item it is a copy of.
func heldPotion(userItemId int, usesLeft int) typeitems.UserItemWithChangeId {
	return typeitems.UserItemWithChangeId{
		UserItem: typeitems.UserItem{Id: userItemId, UserId: 7, PartyId: 1, ItemId: potion.Id, UsesLeft: usesLeft},
		ChangeId: potion.ChangeId,
	}
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

		itemsDb.On("GetUserItemCommand", 7, 1, 11).Return(heldPotion(11, 2), nil)
		changesDb.On("GetChangeEntriesJsonbCommand", 1, potion.ChangeId).
			Return([]typechanges.ChangeEntry{templateEntry}, nil)
		itemsDb.On("ChangeUserItemUsesLeftCommand", 7, 1, 11, 1, 9, (*int)(nil)).Return(55, nil)
		changesDb.On("CreateUserChangeFromJsonbCommand", 1, []typechanges.ChangeEntry{targetedEntry}).
			Return(typechanges.UserChange{Entries: []typechanges.ChangeEntry{targetedEntry}}, nil)
		changesSvc.On("ApplyChangeEntries", 1, []typechanges.ChangeEntry{targetedEntry}, 9, 55).Return(nil)

		sut := srvitems.Service{Database: itemsDb, ChangesDatabase: changesDb, ChangesService: changesSvc}

		err := sut.UseItem(test.Context(), 9, 7, 1, 11)

		require.NoError(test, err)
		itemsDb.AssertExpectations(test)
		itemsDb.AssertNotCalled(test, "DeleteUserItemCommand")
		changesDb.AssertExpectations(test)
		changesSvc.AssertExpectations(test)
	})

	test.Run("LastUse_RemovesItemAttributedToUse", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)

		itemsDb.On("GetUserItemCommand", 7, 1, 11).Return(heldPotion(11, 1), nil)
		changesDb.On("GetChangeEntriesJsonbCommand", 1, potion.ChangeId).Return([]typechanges.ChangeEntry(nil), nil)
		itemsDb.On("ChangeUserItemUsesLeftCommand", 7, 1, 11, 0, 9, (*int)(nil)).Return(55, nil)
		itemsDb.On("DeleteUserItemCommand", 7, 1, 11, 9, ptr(55)).Return(nil)

		sut := srvitems.Service{Database: itemsDb, ChangesDatabase: changesDb}

		err := sut.UseItem(test.Context(), 9, 7, 1, 11)

		require.NoError(test, err)
		itemsDb.AssertExpectations(test)
	})

	// A user can hold several copies of the same item. Using one copy changes that copy only.
	test.Run("SeveralCopies_UsesOnlyTheNamedCopy", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)

		itemsDb.On("GetUserItemCommand", 7, 1, 12).Return(heldPotion(12, 3), nil)
		changesDb.On("GetChangeEntriesJsonbCommand", 1, potion.ChangeId).Return([]typechanges.ChangeEntry(nil), nil)
		itemsDb.On("ChangeUserItemUsesLeftCommand", 7, 1, 12, 2, 9, (*int)(nil)).Return(55, nil)

		sut := srvitems.Service{Database: itemsDb, ChangesDatabase: changesDb}

		err := sut.UseItem(test.Context(), 9, 7, 1, 12)

		require.NoError(test, err)
		itemsDb.AssertExpectations(test)
		itemsDb.AssertNotCalled(test, "GetUserItemCommand", 7, 1, 11)
		itemsDb.AssertNotCalled(test, "ChangeUserItemUsesLeftCommand", 7, 1, 11, mock.Anything, mock.Anything, mock.Anything)
		itemsDb.AssertNotCalled(test, "DeleteUserItemCommand")
	})

	// The copy carries what using it grants, so it stays usable once its item leaves the catalogue.
	test.Run("RemovedFromCatalogue_StillUsable", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)

		itemsDb.On("GetUserItemCommand", 7, 1, 11).Return(heldPotion(11, 2), nil)
		changesDb.On("GetChangeEntriesJsonbCommand", 1, potion.ChangeId).Return([]typechanges.ChangeEntry(nil), nil)
		itemsDb.On("ChangeUserItemUsesLeftCommand", 7, 1, 11, 1, 9, (*int)(nil)).Return(55, nil)

		sut := srvitems.Service{Database: itemsDb, ChangesDatabase: changesDb}

		err := sut.UseItem(test.Context(), 9, 7, 1, 11)

		require.NoError(test, err)
		itemsDb.AssertNotCalled(test, "GetActualItemsCommand", mock.Anything)
	})

	test.Run("NoUsesLeft_Rejected", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		itemsDb.On("GetUserItemCommand", 7, 1, 11).Return(heldPotion(11, 0), nil)

		sut := srvitems.Service{Database: itemsDb}

		err := sut.UseItem(test.Context(), 9, 7, 1, 11)

		var conflict *common.ConflictError
		require.ErrorAs(test, err, &conflict)
		itemsDb.AssertNotCalled(test, "ChangeUserItemUsesLeftCommand")
	})

	test.Run("NotOwned_NotFound", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		itemsDb.On("GetUserItemCommand", 7, 1, 11).Return(typeitems.UserItemWithChangeId{}, sql.ErrNoRows)

		sut := srvitems.Service{Database: itemsDb}

		err := sut.UseItem(test.Context(), 9, 7, 1, 11)

		var notFound *common.NotFoundError
		require.ErrorAs(test, err, &notFound)
		itemsDb.AssertNotCalled(test, "ChangeUserItemUsesLeftCommand")
	})

	test.Run("UserItemRead_DatabaseError", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		itemsDb.On("GetUserItemCommand", 7, 1, 11).Return(typeitems.UserItemWithChangeId{}, dbError)

		sut := srvitems.Service{Database: itemsDb}

		err := sut.UseItem(test.Context(), 9, 7, 1, 11)

		require.ErrorIs(test, err, dbError)
	})

	test.Run("ChangeRead_DatabaseError", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)

		itemsDb.On("GetUserItemCommand", 7, 1, 11).Return(heldPotion(11, 2), nil)
		changesDb.On("GetChangeEntriesJsonbCommand", 1, potion.ChangeId).Return([]typechanges.ChangeEntry(nil), dbError)

		sut := srvitems.Service{Database: itemsDb, ChangesDatabase: changesDb}

		err := sut.UseItem(test.Context(), 9, 7, 1, 11)

		require.ErrorIs(test, err, dbError)
		itemsDb.AssertNotCalled(test, "ChangeUserItemUsesLeftCommand")
	})
}

// Discarding an item removes one copy of it from its holder, and the removal is recorded as the
// discarding user's action.
func TestSrvItems_DiscardUserItem(test *testing.T) {
	test.Run("Success_RecordedAsActor", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)

		itemsDb.On("GetUserItemCommand", 7, 1, 11).Return(heldPotion(11, 2), nil)
		itemsDb.On("DeleteUserItemCommand", 7, 1, 11, 9, (*int)(nil)).Return(nil)

		sut := srvitems.Service{Database: itemsDb}

		err := sut.DiscardUserItem(test.Context(), 9, 7, 1, 11)

		require.NoError(test, err)
		itemsDb.AssertExpectations(test)
	})

	test.Run("NotOwned_NotFound", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		itemsDb.On("GetUserItemCommand", 7, 1, 11).Return(typeitems.UserItemWithChangeId{}, sql.ErrNoRows)

		sut := srvitems.Service{Database: itemsDb}

		err := sut.DiscardUserItem(test.Context(), 9, 7, 1, 11)

		var notFound *common.NotFoundError
		require.ErrorAs(test, err, &notFound)
		itemsDb.AssertNotCalled(test, "DeleteUserItemCommand")
	})
}

// Removing an item from the catalogue is addressed by id and only reaches an item that hasn't been
// removed already.
func TestSrvItems_RemoveItem(test *testing.T) {
	test.Run("Success", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		itemsDb.On("GetActualItemsCommand", 1).Return([]typeitems.Item{potion}, nil)
		itemsDb.On("RemoveItemCommand", 1, potion.Id).Return(nil)

		sut := srvitems.Service{Database: itemsDb}

		err := sut.RemoveItem(test.Context(), 1, potion.Id)

		require.NoError(test, err)
		itemsDb.AssertExpectations(test)
	})

	test.Run("AlreadyRemoved_NotFound", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		itemsDb.On("GetActualItemsCommand", 1).Return([]typeitems.Item{}, nil)

		sut := srvitems.Service{Database: itemsDb}

		err := sut.RemoveItem(test.Context(), 1, potion.Id)

		var notFound *common.NotFoundError
		require.ErrorAs(test, err, &notFound)
		itemsDb.AssertNotCalled(test, "RemoveItemCommand")
	})

	test.Run("CatalogueRead_DatabaseError", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		itemsDb.On("GetActualItemsCommand", 1).Return([]typeitems.Item{}, dbError)

		sut := srvitems.Service{Database: itemsDb}

		err := sut.RemoveItem(test.Context(), 1, potion.Id)

		require.ErrorIs(test, err, dbError)
	})
}
