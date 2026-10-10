package srvperks_test

import (
	"FGG-Service/src/effects/typeeffects"
	"FGG-Service/src/perks/srvperks"
	"FGG-Service/src/perks/typeperks"
	"FGG-Service/tests/effects/dbeffectsmock"
	"FGG-Service/tests/perks/dbperksmock"
	"testing"

	"github.com/stretchr/testify/require"
)

// A perk always wraps one effect, so creating one checks the effect is in the catalogue before writing
// the perk.
func TestSrvPerks_CreatePerk(test *testing.T) {
	test.Run("Success_ChecksEffectInCatalogue", func(test *testing.T) {
		perksDb := new(dbperksmock.DatabaseMock)
		effectsDb := new(dbeffectsmock.DatabaseMock)

		effectsDb.On("GetActualEffectsCommand", 1).
			Return([]typeeffects.Effect{{Id: 12, PartyId: 1, Name: "haste"}}, nil)
		perksDb.On("CreatePerkCommand", 1, "swift", "moves faster", 12).
			Return(typeperks.PerkWithRemoved{Id: 4, PartyId: 1, Name: "swift", Description: "moves faster", EffectId: 12}, nil)

		sut := srvperks.Service{Database: perksDb, EffectsDatabase: effectsDb}

		perk, err := sut.CreatePerk(test.Context(), 1, "swift", "moves faster", 12)

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

		_, err := sut.CreatePerk(test.Context(), 1, "swift", "moves faster", 99)

		require.Error(test, err)
		perksDb.AssertNotCalled(test, "CreatePerkCommand")
	})
}
