package srveffects_test

import (
	typechanges "FGG-Service/src/changes/types"
	srveffects "FGG-Service/src/effects/service"
	typeeffects "FGG-Service/src/effects/types"
	dbchangesmock "FGG-Service/tests/changes/mock"
	srvchangesmock "FGG-Service/tests/changes/srvmock"
	dbeffectsmock "FGG-Service/tests/effects/mock"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

var haste = typeeffects.Effect{Id: 3, PartyId: 1, Name: "haste", Description: "faster", UseCount: 2}

func ptr[T any](value T) *T {
	return &value
}

// Using an effect spends one use and grants what the effect carries, the same way using an item does.
func TestSrvEffects_UseEffect(test *testing.T) {
	test.Run("Success_SpendsUseThenGrants", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)
		changesSvc := new(srvchangesmock.ServiceMock)

		templateEntry := typechanges.ChangeEntry{EntryId: ptr(3), Amount: 2, ItemId: ptr(8)}
		targetedEntry := typechanges.ChangeEntry{Amount: 2, ItemId: ptr(8), UserId: ptr(7)}

		effectsDb.On("GetActualEffectsCommand", 1).Return([]typeeffects.Effect{haste}, nil)
		effectsDb.On("GetUserEffectCommand", 7, 1, haste.Id).
			Return(typeeffects.UserEffectDetail{Id: 21, UserId: 7, PartyId: 1, EffectId: haste.Id, UsesLeft: 2}, nil)
		effectsDb.On("GetEffectCommand", 1, haste.Id).
			Return(typeeffects.EffectWithChange{Id: haste.Id, Change: typechanges.Change{Entries: []typechanges.ChangeEntry{templateEntry}}}, nil)
		effectsDb.On("ChangeUserEffectUsesLeftCommand", 7, 1, haste.Id, 1, 7, (*int)(nil)).Return(nil)
		changesDb.On("CreateUserChangeFromJsonbCommand", 1, []typechanges.ChangeEntry{targetedEntry}).
			Return(typechanges.UserChange{Entries: []typechanges.ChangeEntry{targetedEntry}}, nil)
		changesSvc.On("ApplyChangeEntries", 1, []typechanges.ChangeEntry{targetedEntry}, 7, 21).Return(nil)

		sut := srveffects.Service{Database: effectsDb, ChangesDatabase: changesDb, ChangesService: changesSvc}

		err := sut.UseEffect(7, 1, haste.Name)

		require.NoError(test, err)
		effectsDb.AssertExpectations(test)
		changesDb.AssertExpectations(test)
		changesSvc.AssertExpectations(test)
	})

	test.Run("NoUsesLeft_Rejected", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		effectsDb.On("GetActualEffectsCommand", 1).Return([]typeeffects.Effect{haste}, nil)
		effectsDb.On("GetUserEffectCommand", 7, 1, haste.Id).Return(typeeffects.UserEffectDetail{UsesLeft: 0}, nil)

		sut := srveffects.Service{Database: effectsDb}

		err := sut.UseEffect(7, 1, haste.Name)

		require.Error(test, err)
		effectsDb.AssertNotCalled(test, "ChangeUserEffectUsesLeftCommand")
	})

	test.Run("NotActive_Rejected", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		effectsDb.On("GetActualEffectsCommand", 1).Return([]typeeffects.Effect{haste}, nil)
		effectsDb.On("GetUserEffectCommand", 7, 1, haste.Id).Return(typeeffects.UserEffectDetail{}, sql.ErrNoRows)

		sut := srveffects.Service{Database: effectsDb}

		err := sut.UseEffect(7, 1, haste.Name)

		require.Error(test, err)
		effectsDb.AssertNotCalled(test, "ChangeUserEffectUsesLeftCommand")
	})
}