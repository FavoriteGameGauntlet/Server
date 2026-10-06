package srvparties_test

import (
	"FGG-Service/src/common"
	srvparties "FGG-Service/src/parties/service"
	typeparties "FGG-Service/src/parties/types"
	dbpartiesmock "FGG-Service/tests/parties/mock"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

var requireAccessDbError = errors.New("database connection lost")

// Anyone who isn't a current member of a party, including a member who has left, gets a not-found
// error that cannot be told apart from the one for a party that doesn't exist, so the answer doesn't
// reveal which parties exist.
func TestSrvParties_RequireMember(test *testing.T) {
	test.Run("Member_Allowed", func(test *testing.T) {
		db := new(dbpartiesmock.DatabaseMock)

		db.On("GetMemberCommand", 5, 8).Return(memberJoined(5, 10, false), nil)

		sut := srvparties.Service{Database: db}

		err := sut.RequireMember(test.Context(), 5, 8)

		require.NoError(test, err)
	})

	test.Run("Admin_Allowed", func(test *testing.T) {
		db := new(dbpartiesmock.DatabaseMock)

		db.On("GetMemberCommand", 5, 8).Return(memberJoined(5, 10, true), nil)

		sut := srvparties.Service{Database: db}

		err := sut.RequireMember(test.Context(), 5, 8)

		require.NoError(test, err)
	})

	test.Run("NotAMember_PartyNotFound", func(test *testing.T) {
		db := new(dbpartiesmock.DatabaseMock)

		db.On("GetMemberCommand", 5, 8).Return(typeparties.MemberWithLogin{}, sql.ErrNoRows)

		sut := srvparties.Service{Database: db}

		err := sut.RequireMember(test.Context(), 5, 8)

		requirePartyNotFound(test, err)
	})

	test.Run("MemberWhoLeft_PartyNotFound", func(test *testing.T) {
		db := new(dbpartiesmock.DatabaseMock)

		db.On("GetMemberCommand", 5, 8).Return(memberWhoLeft(5, 10, false), nil)

		sut := srvparties.Service{Database: db}

		err := sut.RequireMember(test.Context(), 5, 8)

		requirePartyNotFound(test, err)
	})

	test.Run("DatabaseFails_ErrorReturned", func(test *testing.T) {
		db := new(dbpartiesmock.DatabaseMock)

		db.On("GetMemberCommand", 5, 8).Return(typeparties.MemberWithLogin{}, requireAccessDbError)

		sut := srvparties.Service{Database: db}

		err := sut.RequireMember(test.Context(), 5, 8)

		require.ErrorIs(test, err, requireAccessDbError)
	})
}

// Admin rights belong to the party in question, not to some other party the user administers.
func TestSrvParties_RequireAdmin(test *testing.T) {
	test.Run("Admin_Allowed", func(test *testing.T) {
		db := new(dbpartiesmock.DatabaseMock)

		db.On("GetMemberCommand", 5, 8).Return(memberJoined(5, 10, true), nil)

		sut := srvparties.Service{Database: db}

		err := sut.RequireAdmin(test.Context(), 5, 8)

		require.NoError(test, err)
	})

	test.Run("MemberWhoIsNotAdmin_NotAdmin", func(test *testing.T) {
		db := new(dbpartiesmock.DatabaseMock)

		db.On("GetMemberCommand", 5, 8).Return(memberJoined(5, 10, false), nil)

		sut := srvparties.Service{Database: db}

		err := sut.RequireAdmin(test.Context(), 5, 8)

		var unauthorized *common.UnauthorizedError
		require.ErrorAs(test, err, &unauthorized)
		require.Equal(test, "NOT_ADMIN", unauthorized.GetCode())
	})

	test.Run("NotAMember_PartyNotFound", func(test *testing.T) {
		db := new(dbpartiesmock.DatabaseMock)

		db.On("GetMemberCommand", 5, 8).Return(typeparties.MemberWithLogin{}, sql.ErrNoRows)

		sut := srvparties.Service{Database: db}

		err := sut.RequireAdmin(test.Context(), 5, 8)

		requirePartyNotFound(test, err)
	})

	test.Run("AdminWhoLeft_PartyNotFound", func(test *testing.T) {
		db := new(dbpartiesmock.DatabaseMock)

		db.On("GetMemberCommand", 5, 8).Return(memberWhoLeft(5, 10, true), nil)

		sut := srvparties.Service{Database: db}

		err := sut.RequireAdmin(test.Context(), 5, 8)

		requirePartyNotFound(test, err)
	})

	test.Run("DatabaseFails_ErrorReturned", func(test *testing.T) {
		db := new(dbpartiesmock.DatabaseMock)

		db.On("GetMemberCommand", 5, 8).Return(typeparties.MemberWithLogin{}, requireAccessDbError)

		sut := srvparties.Service{Database: db}

		err := sut.RequireAdmin(test.Context(), 5, 8)

		require.ErrorIs(test, err, requireAccessDbError)
	})
}

func requirePartyNotFound(test *testing.T, err error) {
	test.Helper()

	var notFound *common.NotFoundError
	require.ErrorAs(test, err, &notFound)
	require.Equal(test, "PARTY_NOT_FOUND", notFound.GetCode())
}
