package srvtimers_test

import (
	"FGG-Service/src/changes/types"
	"FGG-Service/src/timers/service"
	"FGG-Service/tests/changes/srvmock"
	"FGG-Service/tests/timers/mock/dbtimers"
	"testing"

	"github.com/stretchr/testify/require"
)

// Setting a reward stores the entries resolved from their names and returns the reward as stored.
func TestSrvTimers_SetTimerReward_StoresResolvedEntries(test *testing.T) {
	// Arrange
	timerDb := new(dbtimermock.DatabaseMock)
	changes := new(srvchangesmock.ServiceMock)

	pointTypeName := "Rolls"
	pointTypeId := 4
	inputs := []typechanges.ChangeEntryInput{{PointTypeName: &pointTypeName, Amount: 2}}
	resolved := []typechanges.ChangeEntry{{PointTypeId: &pointTypeId, Amount: 2}}

	changes.On("ResolveChangeEntries", 1, inputs).Return(resolved, nil)
	timerDb.On("SetTimerRewardCommand", 1, typechanges.Change{Entries: resolved}).Return(3, nil)
	timerDb.On("GetTimerRewardEntriesCommand", 1).Return(inputs, nil)

	sut := srvtimers.Service{Database: timerDb, ChangesService: changes}

	// Act
	reward, err := sut.SetTimerReward(1, inputs)

	// Assert
	require.NoError(test, err)
	require.Equal(test, inputs, reward)
	timerDb.AssertExpectations(test)
}

// An entry naming something that does not exist leaves the current reward in place.
func TestSrvTimers_SetTimerReward_UnresolvableEntry_KeepsCurrentReward(test *testing.T) {
	// Arrange
	timerDb := new(dbtimermock.DatabaseMock)
	changes := new(srvchangesmock.ServiceMock)

	itemName := "Missing"
	inputs := []typechanges.ChangeEntryInput{{ItemName: &itemName, Amount: 1}}

	changes.On("ResolveChangeEntries", 1, inputs).Return([]typechanges.ChangeEntry(nil), dbError)

	sut := srvtimers.Service{Database: timerDb, ChangesService: changes}

	// Act
	_, err := sut.SetTimerReward(1, inputs)

	// Assert
	require.ErrorIs(test, err, dbError)
	timerDb.AssertNotCalled(test, "SetTimerRewardCommand")
}
