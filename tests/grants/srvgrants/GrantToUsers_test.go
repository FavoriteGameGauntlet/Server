package srvgrants_test

import (
	"FGG-Service/src/changes/typechanges"
	"FGG-Service/src/grants/srvgrants"
	"FGG-Service/src/grants/typegrants"
	"FGG-Service/tests/changes/dbchangesmock"
	"FGG-Service/tests/changes/srvchangesmock"
	"FGG-Service/tests/grants/dbgrantsmock"
	"testing"

	"github.com/stretchr/testify/require"
)

func ptr[T any](value T) *T {
	return &value
}

// Each user granted to gets their own manual history entry, so the entries behind one history row
// are only that user's and the source event recorded against their grants is their own.
func TestSrvGrants_GrantToUsers(test *testing.T) {
	test.Run("Success_RecordsAndAppliesPerUser", func(test *testing.T) {
		grantsDb := new(dbgrantsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)
		changesSvc := new(srvchangesmock.ServiceMock)

		input := typechanges.NamedChangeEntry{PointTypeName: ptr("grantedPoints"), Amount: 5}
		resolved := typechanges.ChangeEntry{Amount: 5, PointTypeId: ptr(2)}

		forFirst := typechanges.ChangeEntry{Amount: 5, PointTypeId: ptr(2), UserId: ptr(3)}
		forSecond := typechanges.ChangeEntry{Amount: 5, PointTypeId: ptr(2), UserId: ptr(4)}

		changesSvc.On("ResolveChangeEntries", 1, []typechanges.NamedChangeEntry{input}).
			Return([]typechanges.ChangeEntry{resolved}, nil)

		createdForFirst := typechanges.ChangeEntry{EntryId: ptr(31), Amount: 5, PointTypeId: ptr(2), UserId: ptr(3)}
		createdForSecond := typechanges.ChangeEntry{EntryId: ptr(32), Amount: 5, PointTypeId: ptr(2), UserId: ptr(4)}

		changesDb.On("CreateUserChangeFromJsonbCommand", 1, []typechanges.ChangeEntry{forFirst}).
			Return(typechanges.UserChange{ChangeId: ptr(21), Entries: []typechanges.ChangeEntry{createdForFirst}}, nil)
		changesDb.On("CreateUserChangeFromJsonbCommand", 1, []typechanges.ChangeEntry{forSecond}).
			Return(typechanges.UserChange{ChangeId: ptr(22), Entries: []typechanges.ChangeEntry{createdForSecond}}, nil)

		grantsDb.On("CreateManualHistoryCommand", 3, 1, 21, 9, (*int)(nil)).
			Return(typegrants.ManualHistoryEntry{Id: 11, UserId: 3, ChangeId: 21}, nil)
		grantsDb.On("CreateManualHistoryCommand", 4, 1, 22, 9, (*int)(nil)).
			Return(typegrants.ManualHistoryEntry{Id: 12, UserId: 4, ChangeId: 22}, nil)

		changesSvc.On("ApplyChangeEntries", 1, []typechanges.ChangeEntry{createdForFirst}, 9, 11).Return(nil)
		changesSvc.On("ApplyChangeEntries", 1, []typechanges.ChangeEntry{createdForSecond}, 9, 12).Return(nil)

		sut := srvgrants.Service{Database: grantsDb, ChangesDatabase: changesDb, ChangesService: changesSvc}

		err := sut.GrantToUsers(test.Context(), 9, 1, []int{3, 4}, []typechanges.NamedChangeEntry{input})

		require.NoError(test, err)
		grantsDb.AssertExpectations(test)
		changesDb.AssertExpectations(test)
		changesSvc.AssertExpectations(test)
	})

	test.Run("UnknownName_RejectedBeforeRecording", func(test *testing.T) {
		grantsDb := new(dbgrantsmock.DatabaseMock)
		changesSvc := new(srvchangesmock.ServiceMock)

		input := typechanges.NamedChangeEntry{ItemId: ptr(99), Amount: 1}

		changesSvc.On("ResolveChangeEntries", 1, []typechanges.NamedChangeEntry{input}).
			Return([]typechanges.ChangeEntry{}, assertError)

		sut := srvgrants.Service{Database: grantsDb, ChangesService: changesSvc}

		err := sut.GrantToUsers(test.Context(), 9, 1, []int{3}, []typechanges.NamedChangeEntry{input})

		require.Error(test, err)
		grantsDb.AssertNotCalled(test, "CreateManualHistoryCommand")
	})
}
