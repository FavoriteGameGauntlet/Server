package srveffects_test

import (
	"FGG-Service/src/changes/typechanges"
	"FGG-Service/src/common"
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

var strength = typepoints.PointTypeInfo{Id: 4, PartyId: 1, Name: "strength"}
var luck = typepoints.PointTypeInfo{Id: 5, PartyId: 1, Name: "luck"}

// The catalogue names the modifiers of each effect, and an effect without any gets none.
func TestSrvEffects_GetEffectsWithModifiers(test *testing.T) {
	test.Run("Actual_ModifiersNamedPerEffect", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		changesDb := newChangesDbWithoutEntries()
		pointsDb := new(dbpointsmock.DatabaseMock)

		effectsDb.On("GetActualEffectsCommand", 1).Return([]typeeffects.Effect{
			{Id: 2, PartyId: 1, Name: "haste", Modifiers: []typeeffects.PointModifier{
				{Id: 7, PointTypeId: strength.Id, Amount: 3},
				{Id: 8, PointTypeId: luck.Id, Amount: -1},
			}},
			{Id: 3, PartyId: 1, Name: "plain", Modifiers: []typeeffects.PointModifier{}},
		}, nil)
		effectsDb.On("GetEffectCommand", 1, 2).Return(effectWithChange(2, 11), nil)
		effectsDb.On("GetEffectCommand", 1, 3).Return(effectWithChange(3, 12), nil)
		pointsDb.On("GetPointTypesCommand", 1).Return([]typepoints.PointTypeInfo{strength, luck}, nil)

		sut := srveffects.Service{Database: effectsDb, ChangesDatabase: changesDb, PointsService: &srvpoints.Service{Database: pointsDb}}

		effects, err := sut.GetEffects(test.Context(), 1)

		require.NoError(test, err)
		require.Len(test, effects, 2)
		require.Equal(test, []typeeffects.NamedPointModifier{
			{PointTypeName: "strength", Amount: 3},
			{PointTypeName: "luck", Amount: -1},
		}, effects[0].Modifiers)
		require.Empty(test, effects[1].Modifiers)
	})

	test.Run("Removed_ModifiersNamed", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		changesDb := newChangesDbWithoutEntries()
		pointsDb := new(dbpointsmock.DatabaseMock)

		effectsDb.On("GetRemovedEffectsCommand", 1).Return([]typeeffects.Effect{
			{Id: 2, PartyId: 1, Name: "haste", Modifiers: []typeeffects.PointModifier{{Id: 7, PointTypeId: strength.Id, Amount: 3}}},
		}, nil)
		effectsDb.On("GetEffectCommand", 1, 2).Return(effectWithChange(2, 11), nil)
		pointsDb.On("GetPointTypesCommand", 1).Return([]typepoints.PointTypeInfo{strength}, nil)

		sut := srveffects.Service{Database: effectsDb, ChangesDatabase: changesDb, PointsService: &srvpoints.Service{Database: pointsDb}}

		effects, err := sut.GetRemovedEffects(test.Context(), 1)

		require.NoError(test, err)
		require.Equal(test, []typeeffects.NamedPointModifier{{PointTypeName: "strength", Amount: 3}}, effects[0].Modifiers)
	})

	test.Run("PointTypeLookupFails_ErrorReturned", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		pointsDb := new(dbpointsmock.DatabaseMock)

		effectsDb.On("GetActualEffectsCommand", 1).Return([]typeeffects.Effect{{Id: 2, PartyId: 1, Name: "haste"}}, nil)
		pointsDb.On("GetPointTypesCommand", 1).Return([]typepoints.PointTypeInfo(nil), sql.ErrConnDone)

		sut := srveffects.Service{Database: effectsDb, PointsService: &srvpoints.Service{Database: pointsDb}}

		_, err := sut.GetEffects(test.Context(), 1)

		require.ErrorIs(test, err, sql.ErrConnDone)
	})
}

// newChangesDbWithoutEntries answers every entries lookup with a change that grants nothing, for tests
// that are not about what an effect grants.
func newChangesDbWithoutEntries() *dbchangesmock.DatabaseMock {
	changesDb := new(dbchangesmock.DatabaseMock)
	changesDb.On("GetNamedChangeEntriesCommand", 1, mock.Anything).Return([]typechanges.NamedChangeEntry{}, nil)

	return changesDb
}

// A new effect stores its modifiers by point type id and answers with them named.
func TestSrvEffects_CreateEffectWithModifiers(test *testing.T) {
	newService := func(effectsDb *dbeffectsmock.DatabaseMock, pointsDb *dbpointsmock.DatabaseMock) srveffects.Service {
		changesSvc := new(srvchangesmock.ServiceMock)
		changesSvc.On("ResolveChangeEntries", 1, mock.Anything).Return([]typechanges.ChangeEntry{}, nil)

		return srveffects.Service{
			Database:        effectsDb,
			ChangesDatabase: newChangesDbWithoutEntries(),
			ChangesService:  changesSvc,
			PointsService:   &srvpoints.Service{Database: pointsDb},
		}
	}

	test.Run("Success_StoresResolvedModifiersAndNamesThem", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		pointsDb := new(dbpointsmock.DatabaseMock)

		pointsDb.On("GetPointTypeByNameCommand", 1, "strength").Return(strength, nil)
		pointsDb.On("GetPointTypeByNameCommand", 1, "luck").Return(luck, nil)
		effectsDb.On("CreateEffectCommand",
			1, "haste", "faster", (*int)(nil), (*time.Duration)(nil),
			typechanges.Change{Entries: []typechanges.ChangeEntry{}},
			[]typeeffects.PointModifier{{PointTypeId: strength.Id, Amount: 3}, {PointTypeId: luck.Id, Amount: -1}}).
			Return(typeeffects.EffectWithChange{Id: 2, PartyId: 1, Name: "haste", Description: "faster", Change: typechanges.Change{ChangeId: &storedChangeId}}, nil)

		sut := newService(effectsDb, pointsDb)

		effect, err := sut.CreateEffect(test.Context(), 1, "haste", "faster", nil, nil, nil, []typeeffects.NamedPointModifier{
			{PointTypeName: "strength", Amount: 3},
			{PointTypeName: "luck", Amount: -1},
		})

		require.NoError(test, err)
		require.Equal(test, []typeeffects.NamedPointModifier{
			{PointTypeName: "strength", Amount: 3},
			{PointTypeName: "luck", Amount: -1},
		}, effect.Modifiers)
		effectsDb.AssertExpectations(test)
	})

	test.Run("NoModifiers_StoresAnEmptyList", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		pointsDb := new(dbpointsmock.DatabaseMock)

		effectsDb.On("CreateEffectCommand",
			1, "plain", "nothing", (*int)(nil), (*time.Duration)(nil),
			typechanges.Change{Entries: []typechanges.ChangeEntry{}},
			[]typeeffects.PointModifier{}).
			Return(typeeffects.EffectWithChange{Id: 3, PartyId: 1, Name: "plain", Description: "nothing", Change: typechanges.Change{ChangeId: &storedChangeId}}, nil)

		sut := newService(effectsDb, pointsDb)

		effect, err := sut.CreateEffect(test.Context(), 1, "plain", "nothing", nil, nil, nil, nil)

		require.NoError(test, err)
		require.Empty(test, effect.Modifiers)
		effectsDb.AssertExpectations(test)
	})

	test.Run("ZeroAmount_RejectedBeforeStoring", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		pointsDb := new(dbpointsmock.DatabaseMock)

		sut := newService(effectsDb, pointsDb)

		_, err := sut.CreateEffect(test.Context(), 1, "haste", "faster", nil, nil, nil, []typeeffects.NamedPointModifier{
			{PointTypeName: "strength", Amount: 0},
		})

		require.Equal(test, common.NewEffectModifierAmountUnprocessableError(0, -2147483648, 2147483647), err)
		effectsDb.AssertNotCalled(test, "CreateEffectCommand", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	test.Run("UnknownPointType_RejectedBeforeStoring", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		pointsDb := new(dbpointsmock.DatabaseMock)

		pointsDb.On("GetPointTypeByNameCommand", 1, "ghost").Return(typepoints.PointTypeInfo{}, sql.ErrNoRows)

		sut := newService(effectsDb, pointsDb)

		_, err := sut.CreateEffect(test.Context(), 1, "haste", "faster", nil, nil, nil, []typeeffects.NamedPointModifier{
			{PointTypeName: "ghost", Amount: 1},
		})

		require.Equal(test, common.NewPointTypeNotFoundError("ghost"), err)
		effectsDb.AssertNotCalled(test, "CreateEffectCommand", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	test.Run("SamePointTypeTwice_RejectedBeforeStoring", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		pointsDb := new(dbpointsmock.DatabaseMock)

		pointsDb.On("GetPointTypeByNameCommand", 1, "strength").Return(strength, nil)

		sut := newService(effectsDb, pointsDb)

		_, err := sut.CreateEffect(test.Context(), 1, "haste", "faster", nil, nil, nil, []typeeffects.NamedPointModifier{
			{PointTypeName: "strength", Amount: 3},
			{PointTypeName: "strength", Amount: 1},
		})

		require.Equal(test, common.NewEffectModifierDuplicateUnprocessableError("strength"), err)
		effectsDb.AssertNotCalled(test, "CreateEffectCommand", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})
}
