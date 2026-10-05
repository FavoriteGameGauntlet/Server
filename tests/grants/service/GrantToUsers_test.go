package srvgrants_test

import (
	typechanges "FGG-Service/src/changes/types"
	srvgrants "FGG-Service/src/grants/service"
	typegrants "FGG-Service/src/grants/types"
	dbchangesmock "FGG-Service/tests/changes/mock"
	srvchangesmock "FGG-Service/tests/changes/srvmock"
	dbgrantsmock "FGG-Service/tests/grants/mock"
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

		input := typechanges.ChangeEntryInput{PointTypeName: ptr("freePoints"), Amount: 5}
		resolved := typechanges.ChangeEntry{Amount: 5, PointTypeId: ptr(2)}

		forFirst := typechanges.ChangeEntry{Amount: 5, PointTypeId: ptr(2), UserId: ptr(3)}
		forSecond := typechanges.ChangeEntry{Amount: 5, PointTypeId: ptr(2), UserId: ptr(4)}

		changesSvc.On("ResolveChangeEntries", 1, []typechanges.ChangeEntryInput{input}).
			Return([]typechanges.ChangeEntry{resolved}, nil)

		grantsDb.On("CreateManualHistoryCommand", 1, 9, []typechanges.ChangeEntry{forFirst}, (*int)(nil)).
			Return([]typegrants.ManualHistoryEntry{{Id: 11, UserId: 3, ChangeId: 21}}, nil)
		grantsDb.On("CreateManualHistoryCommand", 1, 9, []typechanges.ChangeEntry{forSecond}, (*int)(nil)).
			Return([]typegrants.ManualHistoryEntry{{Id: 12, UserId: 4, ChangeId: 22}}, nil)

		changesDb.On("GetChangeEntriesJsonbCommand", 1, 21).Return([]typechanges.ChangeEntry{forFirst}, nil)
		changesDb.On("GetChangeEntriesJsonbCommand", 1, 22).Return([]typechanges.ChangeEntry{forSecond}, nil)

		changesSvc.On("ApplyChangeEntries", 1, []typechanges.ChangeEntry{forFirst}, 9, 11).Return(nil)
		changesSvc.On("ApplyChangeEntries", 1, []typechanges.ChangeEntry{forSecond}, 9, 12).Return(nil)

		sut := srvgrants.Service{Database: grantsDb, ChangesDatabase: changesDb, ChangesService: changesSvc}

		err := sut.GrantToUsers(9, 1, []int{3, 4}, []typechanges.ChangeEntryInput{input})

		require.NoError(test, err)
		grantsDb.AssertExpectations(test)
		changesDb.AssertExpectations(test)
		changesSvc.AssertExpectations(test)
	})

	test.Run("UnknownName_RejectedBeforeRecording", func(test *testing.T) {
		grantsDb := new(dbgrantsmock.DatabaseMock)
		changesSvc := new(srvchangesmock.ServiceMock)

		input := typechanges.ChangeEntryInput{ItemName: ptr("missing"), Amount: 1}

		changesSvc.On("ResolveChangeEntries", 1, []typechanges.ChangeEntryInput{input}).
			Return([]typechanges.ChangeEntry{}, assertError)

		sut := srvgrants.Service{Database: grantsDb, ChangesService: changesSvc}

		err := sut.GrantToUsers(9, 1, []int{3}, []typechanges.ChangeEntryInput{input})

		require.Error(test, err)
		grantsDb.AssertNotCalled(test, "CreateManualHistoryCommand")
	})
}