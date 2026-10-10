package srvchanges_test

import (
	"FGG-Service/src/changes/srvchanges"
	"FGG-Service/src/changes/typechanges"
	"FGG-Service/src/common"
	"FGG-Service/src/effects/typeeffects"
	"FGG-Service/src/items/typeitems"
	"FGG-Service/src/points/srvpoints"
	"FGG-Service/tests/effects/dbeffectsmock"
	"FGG-Service/tests/items/dbitemsmock"
	"FGG-Service/tests/points/dbpointsmock"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

// An amount that does not fit the INTEGER column it is stored in is rejected as unprocessable
// before the entry's target is even looked up.
func TestSrvChanges_ResolveChangeEntries_AmountOutOfIntegerRange_Unprocessable(test *testing.T) {
	for _, amount := range []int{math.MaxInt32 + 1, math.MinInt32 - 1} {
		// Arrange
		pointsDb := new(dbpointsmock.DatabaseMock)
		sut := srvchanges.Service{PointsService: &srvpoints.Service{Database: pointsDb}}

		name := "Rolls"

		// Act
		_, err := sut.ResolveChangeEntries(test.Context(), 1, []typechanges.NamedChangeEntry{{PointTypeName: &name, Amount: amount}})

		// Assert
		var unprocessable *common.UnprocessableError
		require.ErrorAs(test, err, &unprocessable)
		pointsDb.AssertNotCalled(test, "GetPointTypeByNameCommand")
	}
}

// An item or an effect is named by the id of its catalogue entry. The entry has to exist and not be
// removed, otherwise the change would grant something that can't be told apart from a typo.
func TestSrvChanges_ResolveChangeEntries_ItemById(test *testing.T) {
	itemId := 4

	test.Run("Actual_Resolved", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		itemsDb.On("GetActualItemsCommand", 1).Return([]typeitems.Item{{Id: itemId, PartyId: 1, Name: "potion"}}, nil)
		sut := srvchanges.Service{ItemsDatabase: itemsDb}

		entries, err := sut.ResolveChangeEntries(test.Context(), 1, []typechanges.NamedChangeEntry{{ItemId: &itemId, Amount: 2}})

		require.NoError(test, err)
		require.Equal(test, []typechanges.ChangeEntry{{ItemId: &itemId, Amount: 2}}, entries)
	})

	test.Run("RemovedOrMissing_NotFound", func(test *testing.T) {
		itemsDb := new(dbitemsmock.DatabaseMock)
		itemsDb.On("GetActualItemsCommand", 1).Return([]typeitems.Item{}, nil)
		sut := srvchanges.Service{ItemsDatabase: itemsDb}

		_, err := sut.ResolveChangeEntries(test.Context(), 1, []typechanges.NamedChangeEntry{{ItemId: &itemId, Amount: 2}})

		var notFound *common.NotFoundError
		require.ErrorAs(test, err, &notFound)
	})
}

func TestSrvChanges_ResolveChangeEntries_EffectById(test *testing.T) {
	effectId := 3

	test.Run("Actual_Resolved", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		effectsDb.On("GetActualEffectsCommand", 1).Return([]typeeffects.Effect{{Id: effectId, PartyId: 1, Name: "haste"}}, nil)
		sut := srvchanges.Service{EffectsDatabase: effectsDb}

		entries, err := sut.ResolveChangeEntries(test.Context(), 1, []typechanges.NamedChangeEntry{{EffectId: &effectId, Amount: 1}})

		require.NoError(test, err)
		require.Equal(test, []typechanges.ChangeEntry{{EffectId: &effectId, Amount: 1}}, entries)
	})

	test.Run("RemovedOrMissing_NotFound", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		effectsDb.On("GetActualEffectsCommand", 1).Return([]typeeffects.Effect{}, nil)
		sut := srvchanges.Service{EffectsDatabase: effectsDb}

		_, err := sut.ResolveChangeEntries(test.Context(), 1, []typechanges.NamedChangeEntry{{EffectId: &effectId, Amount: 1}})

		var notFound *common.NotFoundError
		require.ErrorAs(test, err, &notFound)
	})
}
