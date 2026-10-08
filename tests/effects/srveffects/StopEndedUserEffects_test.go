package srveffects_test

import (
	"FGG-Service/src/effects/srveffects"
	"FGG-Service/src/effects/typeeffects"
	"FGG-Service/tests/effects/dbeffectsmock"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// An effect that ran out on its own is removed and recorded, with its holder as the one who ended it.
func TestSrvEffects_StopEndedUserEffects(test *testing.T) {
	test.Run("Success_RemovesEachEffectAsItsHolder", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)

		effectsDb.On("GetEndedUserEffectsCommand").Return([]typeeffects.EndedUserEffect{
			{Id: 21, UserId: 7, PartyId: 1, EffectId: 3},
			{Id: 22, UserId: 8, PartyId: 1, EffectId: 4, UsesLeft: ptr(2)},
		}, nil)
		effectsDb.On("DeleteUserEffectCommand", 7, 1, 3, 7, (*int)(nil)).Return(nil)
		effectsDb.On("DeleteUserEffectCommand", 8, 1, 4, 8, (*int)(nil)).Return(nil)

		sut := srveffects.Service{Database: effectsDb}

		err := sut.StopEndedUserEffects(test.Context())

		require.NoError(test, err)
		effectsDb.AssertExpectations(test)
	})

	test.Run("NothingEnded_NothingRemoved", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		effectsDb.On("GetEndedUserEffectsCommand").Return([]typeeffects.EndedUserEffect{}, nil)

		sut := srveffects.Service{Database: effectsDb}

		err := sut.StopEndedUserEffects(test.Context())

		require.NoError(test, err)
		effectsDb.AssertNotCalled(test, "DeleteUserEffectCommand")
	})

	test.Run("RemovalFails_OthersStillRemoved", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)

		effectsDb.On("GetEndedUserEffectsCommand").Return([]typeeffects.EndedUserEffect{
			{Id: 21, UserId: 7, PartyId: 1, EffectId: 3},
			{Id: 22, UserId: 8, PartyId: 1, EffectId: 4},
		}, nil)
		effectsDb.On("DeleteUserEffectCommand", 7, 1, 3, 7, (*int)(nil)).Return(errors.New("db is down"))
		effectsDb.On("DeleteUserEffectCommand", 8, 1, 4, 8, (*int)(nil)).Return(nil)

		sut := srveffects.Service{Database: effectsDb}

		err := sut.StopEndedUserEffects(test.Context())

		require.NoError(test, err)
		effectsDb.AssertExpectations(test)
	})

	test.Run("LookupFails_ErrorReturned", func(test *testing.T) {
		effectsDb := new(dbeffectsmock.DatabaseMock)
		lookupError := errors.New("db is down")
		effectsDb.On("GetEndedUserEffectsCommand").Return([]typeeffects.EndedUserEffect(nil), lookupError)

		sut := srveffects.Service{Database: effectsDb}

		err := sut.StopEndedUserEffects(test.Context())

		require.ErrorIs(test, err, lookupError)
		effectsDb.AssertNotCalled(test, "DeleteUserEffectCommand")
	})
}
