package srveffects_test

import (
	"FGG-Service/src/changes/typechanges"
	"FGG-Service/src/common"
	"FGG-Service/src/effects/srveffects"
	"FGG-Service/src/effects/typeeffects"
	"FGG-Service/tests/changes/dbchangesmock"
	"FGG-Service/tests/changes/srvchangesmock"
	"FGG-Service/tests/effects/dbeffectsmock"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var haste = typeeffects.Effect{Id: 3, PartyId: 1, Name: "haste", Description: "faster", UseCount: ptr(2)}

func ptr[T any](value T) *T {
	return &value
}

// Using an effect spends one use and grants what the effect carries.
// The effect belongs to user 7 and is used by user 9: the grants go to the holder, but everything
// recorded names the user who used it.
func TestSrvEffects_UseEffect(test *testing.T) {
	test.Run("Success_SpendsUseThenGrants", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)
		changesSvc := new(srvchangesmock.ServiceMock)

		templateEntry := typechanges.ChangeEntry{EntryId: ptr(3), Amount: 2, ItemId: ptr(8)}
		targetedEntry := typechanges.ChangeEntry{Amount: 2, ItemId: ptr(8), UserId: ptr(7)}

		effectsDb.On("GetUserEffectCommand", 7, 1, 21).
			Return(typeeffects.UserEffectDetail{Id: 21, UserId: 7, PartyId: 1, EffectId: haste.Id, UsesLeft: ptr(2)}, nil)
		effectsDb.On("GetEffectCommand", 1, haste.Id).
			Return(typeeffects.EffectWithChange{Id: haste.Id, Change: typechanges.Change{Entries: []typechanges.ChangeEntry{templateEntry}}}, nil)
		effectsDb.On("ChangeUserEffectUsesLeftCommand", 7, 1, 21, ptr(1), 9, (*int)(nil)).Return(60, nil)
		changesDb.On("CreateUserChangeFromJsonbCommand", 1, []typechanges.ChangeEntry{targetedEntry}).
			Return(typechanges.UserChange{Entries: []typechanges.ChangeEntry{targetedEntry}}, nil)
		changesSvc.On("ApplyChangeEntries", 1, []typechanges.ChangeEntry{targetedEntry}, 9, 60).Return(nil)

		sut := srveffects.Service{Database: effectsDb, ChangesDatabase: changesDb, ChangesService: changesSvc}

		err := sut.UseEffect(test.Context(), 9, 7, 1, 21)

		require.NoError(test, err)
		effectsDb.AssertExpectations(test)
		effectsDb.AssertNotCalled(test, "DeleteUserEffectCommand")
		changesDb.AssertExpectations(test)
		changesSvc.AssertExpectations(test)
	})

	test.Run("LastUse_RemovesEffectAttributedToUse", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)

		effectsDb.On("GetUserEffectCommand", 7, 1, 21).
			Return(typeeffects.UserEffectDetail{Id: 21, UserId: 7, PartyId: 1, EffectId: haste.Id, UsesLeft: ptr(1)}, nil)
		effectsDb.On("GetEffectCommand", 1, haste.Id).Return(typeeffects.EffectWithChange{Id: haste.Id}, nil)
		effectsDb.On("ChangeUserEffectUsesLeftCommand", 7, 1, 21, ptr(0), 9, (*int)(nil)).Return(60, nil)
		effectsDb.On("DeleteUserEffectCommand", 7, 1, 21, 9, ptr(60)).Return(nil)

		sut := srveffects.Service{Database: effectsDb}

		err := sut.UseEffect(test.Context(), 9, 7, 1, 21)

		require.NoError(test, err)
		effectsDb.AssertExpectations(test)
	})

	// A user can hold several copies of the same effect. Using one copy changes that copy only.
	test.Run("SeveralCopies_UsesOnlyTheNamedCopy", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)

		effectsDb.On("GetUserEffectCommand", 7, 1, 22).
			Return(typeeffects.UserEffectDetail{Id: 22, UserId: 7, PartyId: 1, EffectId: haste.Id, UsesLeft: ptr(2)}, nil)
		effectsDb.On("GetEffectCommand", 1, haste.Id).Return(typeeffects.EffectWithChange{Id: haste.Id}, nil)
		effectsDb.On("ChangeUserEffectUsesLeftCommand", 7, 1, 22, ptr(1), 9, (*int)(nil)).Return(60, nil)

		sut := srveffects.Service{Database: effectsDb}

		err := sut.UseEffect(test.Context(), 9, 7, 1, 22)

		require.NoError(test, err)
		effectsDb.AssertExpectations(test)
		effectsDb.AssertNotCalled(test, "ChangeUserEffectUsesLeftCommand", 7, 1, 21, mock.Anything, mock.Anything, mock.Anything)
		effectsDb.AssertNotCalled(test, "DeleteUserEffectCommand")
	})

	test.Run("Unlimited_RecordsUseAndKeepsEffect", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		unlimited := typeeffects.Effect{Id: 4, PartyId: 1, Name: "aura", Description: "always on"}

		effectsDb.On("GetUserEffectCommand", 7, 1, 23).
			Return(typeeffects.UserEffectDetail{Id: 23, UserId: 7, PartyId: 1, EffectId: unlimited.Id}, nil)
		effectsDb.On("GetEffectCommand", 1, unlimited.Id).Return(typeeffects.EffectWithChange{Id: unlimited.Id}, nil)
		effectsDb.On("ChangeUserEffectUsesLeftCommand", 7, 1, 23, (*int)(nil), 9, (*int)(nil)).Return(61, nil)

		sut := srveffects.Service{Database: effectsDb}

		err := sut.UseEffect(test.Context(), 9, 7, 1, 23)

		require.NoError(test, err)
		effectsDb.AssertExpectations(test)
		effectsDb.AssertNotCalled(test, "DeleteUserEffectCommand")
	})

	test.Run("NoUsesLeft_Rejected", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		effectsDb.On("GetUserEffectCommand", 7, 1, 21).Return(typeeffects.UserEffectDetail{UsesLeft: ptr(0)}, nil)

		sut := srveffects.Service{Database: effectsDb}

		err := sut.UseEffect(test.Context(), 9, 7, 1, 21)

		var conflict *common.ConflictError
		require.ErrorAs(test, err, &conflict)
		effectsDb.AssertNotCalled(test, "ChangeUserEffectUsesLeftCommand")
	})

	test.Run("NotActive_NotFound", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		effectsDb.On("GetUserEffectCommand", 7, 1, 21).Return(typeeffects.UserEffectDetail{}, sql.ErrNoRows)

		sut := srveffects.Service{Database: effectsDb}

		err := sut.UseEffect(test.Context(), 9, 7, 1, 21)

		var notFound *common.NotFoundError
		require.ErrorAs(test, err, &notFound)
		effectsDb.AssertNotCalled(test, "ChangeUserEffectUsesLeftCommand")
	})
}

// Ending an effect removes one copy of it from its holder, and the removal is recorded as the ending
// user's action.
func TestSrvEffects_EndUserEffect(test *testing.T) {
	test.Run("Success_RecordedAsActor", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)

		effectsDb.On("GetUserEffectCommand", 7, 1, 21).
			Return(typeeffects.UserEffectDetail{Id: 21, UserId: 7, PartyId: 1, EffectId: haste.Id, UsesLeft: ptr(2)}, nil)
		effectsDb.On("DeleteUserEffectCommand", 7, 1, 21, 9, (*int)(nil)).Return(nil)

		sut := srveffects.Service{Database: effectsDb}

		err := sut.EndUserEffect(test.Context(), 9, 7, 1, 21)

		require.NoError(test, err)
		effectsDb.AssertExpectations(test)
	})

	test.Run("NotActive_NotFound", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		effectsDb.On("GetUserEffectCommand", 7, 1, 21).Return(typeeffects.UserEffectDetail{}, sql.ErrNoRows)

		sut := srveffects.Service{Database: effectsDb}

		err := sut.EndUserEffect(test.Context(), 9, 7, 1, 21)

		var notFound *common.NotFoundError
		require.ErrorAs(test, err, &notFound)
		effectsDb.AssertNotCalled(test, "DeleteUserEffectCommand")
	})
}

// Removing an effect from the catalogue is addressed by id and only reaches an effect that hasn't
// been removed already.
func TestSrvEffects_RemoveEffect(test *testing.T) {
	test.Run("Success", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		effectsDb.On("GetActualEffectsCommand", 1).Return([]typeeffects.Effect{haste}, nil)
		effectsDb.On("RemoveEffectCommand", 1, haste.Id).Return(nil)

		sut := srveffects.Service{Database: effectsDb}

		err := sut.RemoveEffect(test.Context(), 1, haste.Id)

		require.NoError(test, err)
		effectsDb.AssertExpectations(test)
	})

	test.Run("AlreadyRemoved_NotFound", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		effectsDb.On("GetActualEffectsCommand", 1).Return([]typeeffects.Effect{}, nil)

		sut := srveffects.Service{Database: effectsDb}

		err := sut.RemoveEffect(test.Context(), 1, haste.Id)

		var notFound *common.NotFoundError
		require.ErrorAs(test, err, &notFound)
		effectsDb.AssertNotCalled(test, "RemoveEffectCommand")
	})
}
