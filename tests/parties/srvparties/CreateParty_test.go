package srvparties_test

import (
	"FGG-Service/src/parties/dbparties"
	"FGG-Service/src/parties/srvparties"
	"FGG-Service/src/parties/typeparties"
	"FGG-Service/src/points/srvpoints"
	"FGG-Service/src/points/typepoints"
	"FGG-Service/tests/parties/dbpartiesmock"
	"FGG-Service/tests/points/dbpointsmock"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

var _ dbparties.IDatabase = new(dbpartiesmock.DatabaseMock)

// Admin rights are party membership data, so a new party would have nobody able to administer it
// unless the user creating it joins as an admin. Their points are seeded at the same time, since
// nothing else seeds them.
func TestSrvParties_CreateParty(test *testing.T) {
	test.Run("Success_CreatorJoinsAsAdminWithSeededPoints", func(test *testing.T) {
		partiesDb := new(dbpartiesmock.DatabaseMock)
		pointsDb := new(dbpointsmock.DatabaseMock)

		startingPoints := typepoints.PointTypeInfo{Id: 2, PartyId: 8, Name: "startingPoints", StartValue: 3}

		partiesDb.On("CreatePartyCommand", "gauntlet").Return(typeparties.Party{Id: 8, Name: "gauntlet"}, nil)
		partiesDb.On("GetMemberCommand", 5, 8).Return(typeparties.MemberWithLogin{}, sql.ErrNoRows)
		partiesDb.On("CreateMemberCommand", 5, 8, (*string)(nil), true).
			Return(typeparties.Member{Id: 1, UserId: 5, PartyId: 8, IsAdmin: true}, nil)
		pointsDb.On("GetPointTypesCommand", 8).Return([]typepoints.PointTypeInfo{startingPoints}, nil)
		pointsDb.On("CreateUserPointCommand", 5, 8, startingPoints.Id, startingPoints.StartValue).
			Return(typepoints.UserPoint{}, nil)

		sut := srvparties.Service{Database: partiesDb, PointsService: &srvpoints.Service{Database: pointsDb}}

		party, err := sut.CreateParty(test.Context(), 5, "gauntlet")

		require.NoError(test, err)
		require.Equal(test, 8, party.Id)
		partiesDb.AssertExpectations(test)
		pointsDb.AssertExpectations(test)
	})

	test.Run("AlreadyMember_Rejected", func(test *testing.T) {
		partiesDb := new(dbpartiesmock.DatabaseMock)

		partiesDb.On("GetMemberCommand", 5, 8).
			Return(typeparties.MemberWithLogin{Id: 1, UserId: 5, PartyId: 8}, nil)

		sut := srvparties.Service{Database: partiesDb}

		_, err := sut.AddMember(test.Context(), 5, 8, nil, false)

		require.Error(test, err)
		partiesDb.AssertNotCalled(test, "CreateMemberCommand")
	})
}
