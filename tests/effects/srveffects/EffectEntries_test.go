package srveffects_test

import (
	"FGG-Service/src/changes/typechanges"
	"FGG-Service/src/effects/srveffects"
	"FGG-Service/src/effects/typeeffects"
	"FGG-Service/src/points/srvpoints"
	"FGG-Service/src/points/typepoints"
	"FGG-Service/tests/changes/dbchangesmock"
	"FGG-Service/tests/changes/srvchangesmock"
	"FGG-Service/tests/effects/dbeffectsmock"
	"FGG-Service/tests/points/dbpointsmock"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var storedChangeId = 20

func effectWithChange(effectId int, changeId int) typeeffects.EffectWithChange {
	return typeeffects.EffectWithChange{
		Id:      effectId,
		PartyId: 1,
		Change:  typechanges.Change{ChangeId: &changeId},
	}
}

var hasteEntries = []typechanges.NamedChangeEntry{
	{PointTypeName: ptr("strength"), Amount: 2},
	{ItemName: ptr("potion"), Amount: 1},
}

func TestSrvEffects_EntriesOfListedEffects(test *testing.T) {
	test.Run("Catalogue_EntriesAttachedPerEffect", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)
		pointsDb := new(dbpointsmock.DatabaseMock)

		effectsDb.On("GetActualEffectsCommand", 1).Return([]typeeffects.Effect{
			{Id: 2, PartyId: 1, Name: "haste"},
			{Id: 3, PartyId: 1, Name: "plain"},
		}, nil)
		effectsDb.On("GetEffectCommand", 1, 2).Return(effectWithChange(2, 11), nil)
		effectsDb.On("GetEffectCommand", 1, 3).Return(effectWithChange(3, 12), nil)
		changesDb.On("GetNamedChangeEntriesCommand", 1, 11).Return(hasteEntries, nil)
		changesDb.On("GetNamedChangeEntriesCommand", 1, 12).Return([]typechanges.NamedChangeEntry{}, nil)
		pointsDb.On("GetPointTypesCommand", 1).Return([]typepoints.PointTypeInfo{}, nil)

		sut := srveffects.Service{Database: effectsDb, ChangesDatabase: changesDb, PointsService: &srvpoints.Service{Database: pointsDb}}

		effects, err := sut.GetEffects(test.Context(), 1)

		require.NoError(test, err)
		require.Len(test, effects, 2)
		require.Equal(test, hasteEntries, effects[0].Entries)
		require.Empty(test, effects[1].Entries)
	})

	test.Run("Removed_EntriesAttached", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)
		pointsDb := new(dbpointsmock.DatabaseMock)

		effectsDb.On("GetRemovedEffectsCommand", 1).Return([]typeeffects.Effect{{Id: 2, PartyId: 1, Name: "haste"}}, nil)
		effectsDb.On("GetEffectCommand", 1, 2).Return(effectWithChange(2, 11), nil)
		changesDb.On("GetNamedChangeEntriesCommand", 1, 11).Return(hasteEntries, nil)
		pointsDb.On("GetPointTypesCommand", 1).Return([]typepoints.PointTypeInfo{}, nil)

		sut := srveffects.Service{Database: effectsDb, ChangesDatabase: changesDb, PointsService: &srvpoints.Service{Database: pointsDb}}

		effects, err := sut.GetRemovedEffects(test.Context(), 1)

		require.NoError(test, err)
		require.Equal(test, hasteEntries, effects[0].Entries)
	})

	test.Run("EmptyCatalogue_NothingLookedUp", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)
		pointsDb := new(dbpointsmock.DatabaseMock)

		effectsDb.On("GetActualEffectsCommand", 1).Return([]typeeffects.Effect{}, nil)
		pointsDb.On("GetPointTypesCommand", 1).Return([]typepoints.PointTypeInfo{}, nil)

		sut := srveffects.Service{Database: effectsDb, ChangesDatabase: changesDb, PointsService: &srvpoints.Service{Database: pointsDb}}

		effects, err := sut.GetEffects(test.Context(), 1)

		require.NoError(test, err)
		require.Empty(test, effects)
		effectsDb.AssertNotCalled(test, "GetEffectCommand", mock.Anything, mock.Anything)
	})

	test.Run("EffectLookupFails_ErrorReturned", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)
		pointsDb := new(dbpointsmock.DatabaseMock)

		effectsDb.On("GetActualEffectsCommand", 1).Return([]typeeffects.Effect{{Id: 2, PartyId: 1, Name: "haste"}}, nil)
		effectsDb.On("GetEffectCommand", 1, 2).Return(typeeffects.EffectWithChange{}, sql.ErrConnDone)
		pointsDb.On("GetPointTypesCommand", 1).Return([]typepoints.PointTypeInfo{}, nil)

		sut := srveffects.Service{Database: effectsDb, ChangesDatabase: changesDb, PointsService: &srvpoints.Service{Database: pointsDb}}

		_, err := sut.GetEffects(test.Context(), 1)

		require.ErrorIs(test, err, sql.ErrConnDone)
	})

	test.Run("EntriesLookupFails_ErrorReturned", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)
		pointsDb := new(dbpointsmock.DatabaseMock)

		effectsDb.On("GetActualEffectsCommand", 1).Return([]typeeffects.Effect{{Id: 2, PartyId: 1, Name: "haste"}}, nil)
		effectsDb.On("GetEffectCommand", 1, 2).Return(effectWithChange(2, 11), nil)
		changesDb.On("GetNamedChangeEntriesCommand", 1, 11).Return([]typechanges.NamedChangeEntry{}, sql.ErrConnDone)
		pointsDb.On("GetPointTypesCommand", 1).Return([]typepoints.PointTypeInfo{}, nil)

		sut := srveffects.Service{Database: effectsDb, ChangesDatabase: changesDb, PointsService: &srvpoints.Service{Database: pointsDb}}

		_, err := sut.GetEffects(test.Context(), 1)

		require.ErrorIs(test, err, sql.ErrConnDone)
	})

	test.Run("ChangeWithoutId_ErrorReturned", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)
		pointsDb := new(dbpointsmock.DatabaseMock)

		effectsDb.On("GetActualEffectsCommand", 1).Return([]typeeffects.Effect{{Id: 2, PartyId: 1, Name: "haste"}}, nil)
		effectsDb.On("GetEffectCommand", 1, 2).Return(typeeffects.EffectWithChange{Id: 2, PartyId: 1}, nil)
		pointsDb.On("GetPointTypesCommand", 1).Return([]typepoints.PointTypeInfo{}, nil)

		sut := srveffects.Service{Database: effectsDb, ChangesDatabase: changesDb, PointsService: &srvpoints.Service{Database: pointsDb}}

		_, err := sut.GetEffects(test.Context(), 1)

		require.Error(test, err)
		changesDb.AssertNotCalled(test, "GetNamedChangeEntriesCommand", mock.Anything, mock.Anything)
	})
}

func TestSrvEffects_EntriesOfUserEffects(test *testing.T) {
	test.Run("EntriesAttachedPerEffect", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)
		pointsDb := new(dbpointsmock.DatabaseMock)

		effectsDb.On("GetUserEffectsCommand", 5, 1).Return([]typeeffects.UserEffectDetail{
			{Id: 30, UserId: 5, PartyId: 1, EffectId: 2, Name: "haste"},
			{Id: 31, UserId: 5, PartyId: 1, EffectId: 3, Name: "plain"},
		}, nil)
		effectsDb.On("GetEffectCommand", 1, 2).Return(effectWithChange(2, 11), nil)
		effectsDb.On("GetEffectCommand", 1, 3).Return(effectWithChange(3, 12), nil)
		changesDb.On("GetNamedChangeEntriesCommand", 1, 11).Return(hasteEntries, nil)
		changesDb.On("GetNamedChangeEntriesCommand", 1, 12).Return([]typechanges.NamedChangeEntry{}, nil)
		pointsDb.On("GetPointTypesCommand", 1).Return([]typepoints.PointTypeInfo{}, nil)

		sut := srveffects.Service{Database: effectsDb, ChangesDatabase: changesDb, PointsService: &srvpoints.Service{Database: pointsDb}}

		userEffects, err := sut.GetNamedUserEffects(test.Context(), 5, 1)

		require.NoError(test, err)
		require.Len(test, userEffects, 2)
		require.Equal(test, hasteEntries, userEffects[0].Entries)
		require.Empty(test, userEffects[1].Entries)
	})

	test.Run("EntriesLookupFails_ErrorReturned", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)
		pointsDb := new(dbpointsmock.DatabaseMock)

		effectsDb.On("GetUserEffectsCommand", 5, 1).Return([]typeeffects.UserEffectDetail{{Id: 30, UserId: 5, PartyId: 1, EffectId: 2}}, nil)
		effectsDb.On("GetEffectCommand", 1, 2).Return(effectWithChange(2, 11), nil)
		changesDb.On("GetNamedChangeEntriesCommand", 1, 11).Return([]typechanges.NamedChangeEntry{}, sql.ErrConnDone)
		pointsDb.On("GetPointTypesCommand", 1).Return([]typepoints.PointTypeInfo{}, nil)

		sut := srveffects.Service{Database: effectsDb, ChangesDatabase: changesDb, PointsService: &srvpoints.Service{Database: pointsDb}}

		_, err := sut.GetNamedUserEffects(test.Context(), 5, 1)

		require.ErrorIs(test, err, sql.ErrConnDone)
	})
}

func TestSrvEffects_EntriesOfEffectHistory(test *testing.T) {
	test.Run("EntriesAttachedPerRow_EffectLookedUpOnce", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)

		effectsDb.On("GetEffectHistoryCommand", 5, 1).Return([]typeeffects.EffectHistory{
			{Id: 40, UserId: 5, EffectId: 2, Name: "haste", Action: "used", CreatedDate: time.Now()},
			{Id: 41, UserId: 5, EffectId: 3, Name: "plain", Action: "gained", CreatedDate: time.Now()},
			{Id: 42, UserId: 5, EffectId: 2, Name: "haste", Action: "gained", CreatedDate: time.Now()},
		}, nil)
		effectsDb.On("GetEffectCommand", 1, 2).Return(effectWithChange(2, 11), nil)
		effectsDb.On("GetEffectCommand", 1, 3).Return(effectWithChange(3, 12), nil)
		changesDb.On("GetNamedChangeEntriesCommand", 1, 11).Return(hasteEntries, nil)
		changesDb.On("GetNamedChangeEntriesCommand", 1, 12).Return([]typechanges.NamedChangeEntry{}, nil)

		sut := srveffects.Service{Database: effectsDb, ChangesDatabase: changesDb}

		history, err := sut.GetEffectHistory(test.Context(), 5, 1)

		require.NoError(test, err)
		require.Len(test, history, 3)
		require.Equal(test, hasteEntries, history[0].Entries)
		require.Empty(test, history[1].Entries)
		require.Equal(test, hasteEntries, history[2].Entries)
		effectsDb.AssertNumberOfCalls(test, "GetEffectCommand", 2)
	})

	test.Run("NoHistory_NothingLookedUp", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)

		effectsDb.On("GetEffectHistoryCommand", 5, 1).Return([]typeeffects.EffectHistory{}, nil)

		sut := srveffects.Service{Database: effectsDb, ChangesDatabase: changesDb}

		history, err := sut.GetEffectHistory(test.Context(), 5, 1)

		require.NoError(test, err)
		require.Empty(test, history)
		effectsDb.AssertNotCalled(test, "GetEffectCommand", mock.Anything, mock.Anything)
	})

	test.Run("HistoryLookupFails_ErrorReturned", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)

		effectsDb.On("GetEffectHistoryCommand", 5, 1).Return([]typeeffects.EffectHistory(nil), sql.ErrConnDone)

		sut := srveffects.Service{Database: effectsDb, ChangesDatabase: changesDb}

		_, err := sut.GetEffectHistory(test.Context(), 5, 1)

		require.ErrorIs(test, err, sql.ErrConnDone)
		effectsDb.AssertNotCalled(test, "GetEffectCommand", mock.Anything, mock.Anything)
	})

	test.Run("EntriesLookupFails_ErrorReturned", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)

		effectsDb.On("GetEffectHistoryCommand", 5, 1).Return([]typeeffects.EffectHistory{{Id: 40, UserId: 5, EffectId: 2}}, nil)
		effectsDb.On("GetEffectCommand", 1, 2).Return(effectWithChange(2, 11), nil)
		changesDb.On("GetNamedChangeEntriesCommand", 1, 11).Return([]typechanges.NamedChangeEntry{}, sql.ErrConnDone)

		sut := srveffects.Service{Database: effectsDb, ChangesDatabase: changesDb}

		history, err := sut.GetEffectHistory(test.Context(), 5, 1)

		require.ErrorIs(test, err, sql.ErrConnDone)
		require.Nil(test, history)
	})
}

func TestSrvEffects_CreateEffectReturnsStoredEntries(test *testing.T) {
	test.Run("Success_StoredEntriesNamed", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)
		changesSvc := new(srvchangesmock.ServiceMock)

		resolved := []typechanges.ChangeEntry{{Amount: 2, PointTypeId: ptr(4)}}
		changesSvc.On("ResolveChangeEntries", 1, mock.Anything).Return(resolved, nil)
		effectsDb.On("CreateEffectCommand",
			1, "haste", "faster", (*int)(nil), (*time.Duration)(nil),
			typechanges.Change{Entries: resolved}, []typeeffects.PointModifier{}).
			Return(typeeffects.EffectWithChange{
				Id:     2,
				Name:   "haste",
				Change: typechanges.Change{ChangeId: &storedChangeId, Entries: resolved},
			}, nil)
		changesDb.On("GetNamedChangeEntriesCommand", 1, storedChangeId).Return(hasteEntries, nil)

		sut := srveffects.Service{Database: effectsDb, ChangesDatabase: changesDb, ChangesService: changesSvc}

		effect, err := sut.CreateEffect(test.Context(), 1, "haste", "faster", nil, nil, []typechanges.NamedChangeEntry{
			{PointTypeName: ptr("strength"), Amount: 2},
		}, nil)

		require.NoError(test, err)
		require.Equal(test, hasteEntries, effect.Entries)
	})

	test.Run("EntriesLookupFails_ErrorReturned", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)
		changesSvc := new(srvchangesmock.ServiceMock)

		changesSvc.On("ResolveChangeEntries", 1, mock.Anything).Return([]typechanges.ChangeEntry{}, nil)
		effectsDb.On("CreateEffectCommand", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(typeeffects.EffectWithChange{Id: 2, Change: typechanges.Change{ChangeId: &storedChangeId}}, nil)
		changesDb.On("GetNamedChangeEntriesCommand", 1, storedChangeId).Return([]typechanges.NamedChangeEntry{}, sql.ErrConnDone)

		sut := srveffects.Service{Database: effectsDb, ChangesDatabase: changesDb, ChangesService: changesSvc}

		_, err := sut.CreateEffect(test.Context(), 1, "haste", "faster", nil, nil, nil, nil)

		require.ErrorIs(test, err, sql.ErrConnDone)
	})
}
