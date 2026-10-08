package srveffects_test

import (
	"FGG-Service/src/changes/typechanges"
	"FGG-Service/src/effects/srveffects"
	"FGG-Service/src/effects/typeeffects"
	"FGG-Service/tests/changes/dbchangesmock"
	"FGG-Service/tests/changes/srvchangesmock"
	"FGG-Service/tests/effects/dbeffectsmock"
	"database/sql"
	"testing"

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

		effectsDb.On("GetActualEffectsCommand", 1).Return([]typeeffects.Effect{haste}, nil)
		effectsDb.On("GetUserEffectCommand", 7, 1, haste.Id).
			Return(typeeffects.UserEffectDetail{Id: 21, UserId: 7, PartyId: 1, EffectId: haste.Id, UsesLeft: ptr(2)}, nil)
		effectsDb.On("GetEffectCommand", 1, haste.Id).
			Return(typeeffects.EffectWithChange{Id: haste.Id, Change: typechanges.Change{Entries: []typechanges.ChangeEntry{templateEntry}}}, nil)
		effectsDb.On("ChangeUserEffectUsesLeftCommand", 7, 1, haste.Id, ptr(1), 9, (*int)(nil)).Return(60, nil)
		changesDb.On("CreateUserChangeFromJsonbCommand", 1, []typechanges.ChangeEntry{targetedEntry}).
			Return(typechanges.UserChange{Entries: []typechanges.ChangeEntry{targetedEntry}}, nil)
		changesSvc.On("ApplyChangeEntries", 1, []typechanges.ChangeEntry{targetedEntry}, 9, 60).Return(nil)

		sut := srveffects.Service{Database: effectsDb, ChangesDatabase: changesDb, ChangesService: changesSvc}

		err := sut.UseEffect(test.Context(), 9, 7, 1, haste.Name)

		require.NoError(test, err)
		effectsDb.AssertExpectations(test)
		effectsDb.AssertNotCalled(test, "DeleteUserEffectCommand")
		changesDb.AssertExpectations(test)
		changesSvc.AssertExpectations(test)
	})

	test.Run("LastUse_RemovesEffectAttributedToUse", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)

		effectsDb.On("GetActualEffectsCommand", 1).Return([]typeeffects.Effect{haste}, nil)
		effectsDb.On("GetUserEffectCommand", 7, 1, haste.Id).
			Return(typeeffects.UserEffectDetail{Id: 21, UserId: 7, PartyId: 1, EffectId: haste.Id, UsesLeft: ptr(1)}, nil)
		effectsDb.On("GetEffectCommand", 1, haste.Id).Return(typeeffects.EffectWithChange{Id: haste.Id}, nil)
		effectsDb.On("ChangeUserEffectUsesLeftCommand", 7, 1, haste.Id, ptr(0), 9, (*int)(nil)).Return(60, nil)
		effectsDb.On("DeleteUserEffectCommand", 7, 1, haste.Id, 9, ptr(60)).Return(nil)

		sut := srveffects.Service{Database: effectsDb}

		err := sut.UseEffect(test.Context(), 9, 7, 1, haste.Name)

		require.NoError(test, err)
		effectsDb.AssertExpectations(test)
	})

	test.Run("Unlimited_RecordsUseAndKeepsEffect", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		unlimited := typeeffects.Effect{Id: 4, PartyId: 1, Name: "aura", Description: "always on"}

		effectsDb.On("GetActualEffectsCommand", 1).Return([]typeeffects.Effect{unlimited}, nil)
		effectsDb.On("GetUserEffectCommand", 7, 1, unlimited.Id).
			Return(typeeffects.UserEffectDetail{Id: 22, UserId: 7, PartyId: 1, EffectId: unlimited.Id}, nil)
		effectsDb.On("GetEffectCommand", 1, unlimited.Id).Return(typeeffects.EffectWithChange{Id: unlimited.Id}, nil)
		effectsDb.On("ChangeUserEffectUsesLeftCommand", 7, 1, unlimited.Id, (*int)(nil), 9, (*int)(nil)).Return(61, nil)

		sut := srveffects.Service{Database: effectsDb}

		err := sut.UseEffect(test.Context(), 9, 7, 1, unlimited.Name)

		require.NoError(test, err)
		effectsDb.AssertExpectations(test)
		effectsDb.AssertNotCalled(test, "DeleteUserEffectCommand")
	})

	test.Run("NoUsesLeft_Rejected", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		effectsDb.On("GetActualEffectsCommand", 1).Return([]typeeffects.Effect{haste}, nil)
		effectsDb.On("GetUserEffectCommand", 7, 1, haste.Id).Return(typeeffects.UserEffectDetail{UsesLeft: ptr(0)}, nil)

		sut := srveffects.Service{Database: effectsDb}

		err := sut.UseEffect(test.Context(), 9, 7, 1, haste.Name)

		require.Error(test, err)
		effectsDb.AssertNotCalled(test, "ChangeUserEffectUsesLeftCommand")
	})

	test.Run("NotActive_Rejected", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		effectsDb.On("GetActualEffectsCommand", 1).Return([]typeeffects.Effect{haste}, nil)
		effectsDb.On("GetUserEffectCommand", 7, 1, haste.Id).Return(typeeffects.UserEffectDetail{}, sql.ErrNoRows)

		sut := srveffects.Service{Database: effectsDb}

		err := sut.UseEffect(test.Context(), 9, 7, 1, haste.Name)

		require.Error(test, err)
		effectsDb.AssertNotCalled(test, "ChangeUserEffectUsesLeftCommand")
	})
}

// Ending an effect removes it from its holder, and the removal is recorded as the ending user's action.
func TestSrvEffects_EndUserEffect(test *testing.T) {
	test.Run("Success_RecordedAsActor", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)

		effectsDb.On("GetActualEffectsCommand", 1).Return([]typeeffects.Effect{haste}, nil)
		effectsDb.On("GetUserEffectCommand", 7, 1, haste.Id).
			Return(typeeffects.UserEffectDetail{Id: 21, UserId: 7, PartyId: 1, EffectId: haste.Id, UsesLeft: ptr(2)}, nil)
		effectsDb.On("DeleteUserEffectCommand", 7, 1, haste.Id, 9, (*int)(nil)).Return(nil)

		sut := srveffects.Service{Database: effectsDb}

		err := sut.EndUserEffect(test.Context(), 9, 7, 1, haste.Name)

		require.NoError(test, err)
		effectsDb.AssertExpectations(test)
	})
}
