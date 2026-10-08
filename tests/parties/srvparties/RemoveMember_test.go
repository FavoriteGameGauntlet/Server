package srvparties_test

import (
	"FGG-Service/src/common"
	"FGG-Service/src/parties/srvparties"
	"FGG-Service/src/parties/typeparties"
	"FGG-Service/tests/parties/dbpartiesmock"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var removeMemberDbError = errors.New("database connection lost")

func memberJoined(userId int, daysAgo int, isAdmin bool) typeparties.MemberWithLogin {
	return typeparties.MemberWithLogin{
		UserId:     userId,
		PartyId:    8,
		IsAdmin:    isAdmin,
		JoinedDate: time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -daysAgo),
	}
}

func memberWhoLeft(userId int, daysAgo int, isAdmin bool) typeparties.MemberWithLogin {
	member := memberJoined(userId, daysAgo, isAdmin)
	leftDate := member.JoinedDate.AddDate(0, 0, 1)
	member.LeftDate = &leftDate

	return member
}

// A party has to keep an admin, so when the last one leaves the member who joined earliest takes
// over. Members who already left are not candidates, and the promotion happens before the removal
// so a failed removal can only leave an extra admin, never none.
func TestSrvParties_RemoveMember(test *testing.T) {
	test.Run("LastAdminLeaves_EarliestJoinedMemberPromoted", func(test *testing.T) {
		db := new(dbpartiesmock.DatabaseMock)
		var calls []string

		db.On("GetMemberCommand", 1, 8).Return(typeparties.MemberWithLogin{UserId: 1, IsAdmin: true}, nil)
		db.On("GetMembersCommand", 8).Return([]typeparties.MemberWithLogin{
			memberJoined(1, 30, true),
			memberJoined(2, 5, false),
			memberJoined(3, 20, false),
			memberJoined(4, 10, false),
		}, nil)
		db.On("ChangeMemberAdminStatusCommand", 3, 8, true).
			Run(func(mock.Arguments) { calls = append(calls, "promote") }).
			Return(nil)
		db.On("RemoveMemberCommand", 1, 8).
			Run(func(mock.Arguments) { calls = append(calls, "remove") }).
			Return(nil)

		sut := srvparties.Service{Database: db}

		err := sut.RemoveMember(test.Context(), 1, 8)

		require.NoError(test, err)
		require.Equal(test, []string{"promote", "remove"}, calls)
		db.AssertExpectations(test)
	})

	test.Run("LastAdminLeaves_MembersWhoLeftAreSkipped", func(test *testing.T) {
		db := new(dbpartiesmock.DatabaseMock)

		db.On("GetMemberCommand", 1, 8).Return(typeparties.MemberWithLogin{UserId: 1, IsAdmin: true}, nil)
		db.On("GetMembersCommand", 8).Return([]typeparties.MemberWithLogin{
			memberJoined(1, 30, true),
			memberWhoLeft(2, 25, false),
			memberJoined(3, 5, false),
		}, nil)
		db.On("ChangeMemberAdminStatusCommand", 3, 8, true).Return(nil)
		db.On("RemoveMemberCommand", 1, 8).Return(nil)

		sut := srvparties.Service{Database: db}

		err := sut.RemoveMember(test.Context(), 1, 8)

		require.NoError(test, err)
		db.AssertExpectations(test)
	})

	test.Run("LastAdminLeaves_AdminWhoLeftDoesNotCount", func(test *testing.T) {
		db := new(dbpartiesmock.DatabaseMock)

		db.On("GetMemberCommand", 1, 8).Return(typeparties.MemberWithLogin{UserId: 1, IsAdmin: true}, nil)
		db.On("GetMembersCommand", 8).Return([]typeparties.MemberWithLogin{
			memberJoined(1, 30, true),
			memberWhoLeft(2, 25, true),
			memberJoined(3, 5, false),
		}, nil)
		db.On("ChangeMemberAdminStatusCommand", 3, 8, true).Return(nil)
		db.On("RemoveMemberCommand", 1, 8).Return(nil)

		sut := srvparties.Service{Database: db}

		err := sut.RemoveMember(test.Context(), 1, 8)

		require.NoError(test, err)
		db.AssertExpectations(test)
	})

	test.Run("AnotherAdminRemains_NobodyPromoted", func(test *testing.T) {
		db := new(dbpartiesmock.DatabaseMock)

		db.On("GetMemberCommand", 1, 8).Return(typeparties.MemberWithLogin{UserId: 1, IsAdmin: true}, nil)
		// The non-admin comes before the admin, so the admin is only found after a candidate was picked.
		db.On("GetMembersCommand", 8).Return([]typeparties.MemberWithLogin{
			memberJoined(1, 30, true),
			memberJoined(2, 20, false),
			memberJoined(3, 5, true),
		}, nil)
		db.On("RemoveMemberCommand", 1, 8).Return(nil)

		sut := srvparties.Service{Database: db}

		err := sut.RemoveMember(test.Context(), 1, 8)

		require.NoError(test, err)
		db.AssertNotCalled(test, "ChangeMemberAdminStatusCommand", mock.Anything, mock.Anything, mock.Anything)
		db.AssertExpectations(test)
	})

	test.Run("NonAdminLeaves_NobodyPromoted", func(test *testing.T) {
		db := new(dbpartiesmock.DatabaseMock)

		db.On("GetMemberCommand", 2, 8).Return(typeparties.MemberWithLogin{UserId: 2}, nil)
		db.On("GetMembersCommand", 8).Return([]typeparties.MemberWithLogin{
			memberJoined(1, 30, true),
			memberJoined(2, 20, false),
		}, nil).Maybe()
		db.On("RemoveMemberCommand", 2, 8).Return(nil)

		sut := srvparties.Service{Database: db}

		err := sut.RemoveMember(test.Context(), 2, 8)

		require.NoError(test, err)
		db.AssertNotCalled(test, "ChangeMemberAdminStatusCommand", mock.Anything, mock.Anything, mock.Anything)
		db.AssertExpectations(test)
	})

	test.Run("OnlyMemberLeaves_NobodyToPromote", func(test *testing.T) {
		db := new(dbpartiesmock.DatabaseMock)

		db.On("GetMemberCommand", 1, 8).Return(typeparties.MemberWithLogin{UserId: 1, IsAdmin: true}, nil)
		db.On("GetMembersCommand", 8).Return([]typeparties.MemberWithLogin{
			memberJoined(1, 30, true),
			memberWhoLeft(2, 20, false),
		}, nil)
		db.On("RemoveMemberCommand", 1, 8).Return(nil)

		sut := srvparties.Service{Database: db}

		err := sut.RemoveMember(test.Context(), 1, 8)

		require.NoError(test, err)
		db.AssertNotCalled(test, "ChangeMemberAdminStatusCommand", mock.Anything, mock.Anything, mock.Anything)
		db.AssertExpectations(test)
	})

	test.Run("NotAMember_Rejected", func(test *testing.T) {
		db := new(dbpartiesmock.DatabaseMock)

		db.On("GetMemberCommand", 1, 8).Return(typeparties.MemberWithLogin{}, sql.ErrNoRows)

		sut := srvparties.Service{Database: db}

		err := sut.RemoveMember(test.Context(), 1, 8)

		var notFound *common.NotFoundError
		require.ErrorAs(test, err, &notFound)
		db.AssertNotCalled(test, "GetMembersCommand", mock.Anything)
		db.AssertNotCalled(test, "RemoveMemberCommand", mock.Anything, mock.Anything)
	})

	test.Run("GetMembersFails_NothingChanged", func(test *testing.T) {
		db := new(dbpartiesmock.DatabaseMock)

		db.On("GetMemberCommand", 1, 8).Return(typeparties.MemberWithLogin{UserId: 1, IsAdmin: true}, nil)
		db.On("GetMembersCommand", 8).Return([]typeparties.MemberWithLogin(nil), removeMemberDbError)

		sut := srvparties.Service{Database: db}

		err := sut.RemoveMember(test.Context(), 1, 8)

		require.ErrorIs(test, err, removeMemberDbError)
		db.AssertNotCalled(test, "ChangeMemberAdminStatusCommand", mock.Anything, mock.Anything, mock.Anything)
		db.AssertNotCalled(test, "RemoveMemberCommand", mock.Anything, mock.Anything)
	})

	test.Run("PromotionFails_MemberNotRemoved", func(test *testing.T) {
		db := new(dbpartiesmock.DatabaseMock)

		db.On("GetMemberCommand", 1, 8).Return(typeparties.MemberWithLogin{UserId: 1, IsAdmin: true}, nil)
		db.On("GetMembersCommand", 8).Return([]typeparties.MemberWithLogin{
			memberJoined(1, 30, true),
			memberJoined(2, 5, false),
		}, nil)
		db.On("ChangeMemberAdminStatusCommand", 2, 8, true).Return(removeMemberDbError)

		sut := srvparties.Service{Database: db}

		err := sut.RemoveMember(test.Context(), 1, 8)

		require.ErrorIs(test, err, removeMemberDbError)
		db.AssertNotCalled(test, "RemoveMemberCommand", mock.Anything, mock.Anything)
	})

	test.Run("RemoveFails_ErrorReturned", func(test *testing.T) {
		db := new(dbpartiesmock.DatabaseMock)

		db.On("GetMemberCommand", 2, 8).Return(typeparties.MemberWithLogin{UserId: 2}, nil)
		db.On("GetMembersCommand", 8).Return([]typeparties.MemberWithLogin{
			memberJoined(1, 30, true),
			memberJoined(2, 20, false),
		}, nil).Maybe()
		db.On("RemoveMemberCommand", 2, 8).Return(removeMemberDbError)

		sut := srvparties.Service{Database: db}

		err := sut.RemoveMember(test.Context(), 2, 8)

		require.ErrorIs(test, err, removeMemberDbError)
	})
}
