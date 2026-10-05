package srvperks_test

import (
	srvperks "FGG-Service/src/perks/service"
	typeperks "FGG-Service/src/perks/types"
	dbperksmock "FGG-Service/tests/perks/mock"
	"testing"

	"github.com/stretchr/testify/require"
)

// Revoking a perk takes it from its holder, and the revocation is recorded as the actor's action,
// not the holder's.
func TestSrvPerks_RevokeUserPerk(test *testing.T) {
	test.Run("Success_RecordedAsActor", func(test *testing.T) {
		perksDb := new(dbperksmock.DatabaseMock)

		perksDb.On("GetActualPerksCommand", 1).Return([]typeperks.Perk{{Id: 4, PartyId: 1, Name: "swift"}}, nil)
		perksDb.On("GetUserPerkCommand", 7, 1, 4).Return(typeperks.UserPerk{Id: 1, UserId: 7, PartyId: 1, PerkId: 4}, nil)
		perksDb.On("DeleteUserPerkCommand", 7, 1, 4, 9, (*int)(nil)).Return(nil)

		sut := srvperks.Service{Database: perksDb}

		err := sut.RevokeUserPerk(9, 7, 1, "swift")

		require.NoError(test, err)
		perksDb.AssertExpectations(test)
	})
}
