package srvperks_test

import (
	typeeffects "FGG-Service/src/effects/types"
	srvperks "FGG-Service/src/perks/service"
	typeperks "FGG-Service/src/perks/types"
	dbeffectsmock "FGG-Service/tests/effects/mock"
	dbperksmock "FGG-Service/tests/perks/mock"
	"testing"

	"github.com/stretchr/testify/require"
)

// A perk always wraps one effect, so creating one resolves the named effect before writing the perk.
func TestSrvPerks_CreatePerk(test *testing.T) {
	test.Run("Success_ResolvesEffectByName", func(test *testing.T) {
		perksDb := new(dbperksmock.DatabaseMock)
		effectsDb := new(dbeffectsmock.DatabaseMock)

		effectsDb.On("GetActualEffectsCommand", 1).
			Return([]typeeffects.Effect{{Id: 12, PartyId: 1, Name: "haste"}}, nil)
		perksDb.On("CreatePerkCommand", 1, "swift", "moves faster", 12).
			Return(typeperks.PerkWithRemoved{Id: 4, PartyId: 1, Name: "swift", Description: "moves faster", EffectId: 12}, nil)

		sut := srvperks.Service{Database: perksDb, EffectsDatabase: effectsDb}

		perk, err := sut.CreatePerk(1, "swift", "moves faster", "haste")

		require.NoError(test, err)
		require.Equal(test, 12, perk.EffectId)
		perksDb.AssertExpectations(test)
		effectsDb.AssertExpectations(test)
	})

	test.Run("UnknownEffect_Rejected", func(test *testing.T) {
		perksDb := new(dbperksmock.DatabaseMock)
		effectsDb := new(dbeffectsmock.DatabaseMock)

		effectsDb.On("GetActualEffectsCommand", 1).Return([]typeeffects.Effect{}, nil)

		sut := srvperks.Service{Database: perksDb, EffectsDatabase: effectsDb}

		_, err := sut.CreatePerk(1, "swift", "moves faster", "missing")

		require.Error(test, err)
		perksDb.AssertNotCalled(test, "CreatePerkCommand")
	})
}