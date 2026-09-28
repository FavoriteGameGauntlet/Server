package srvtimers_test

import (
	"FGG-Service/src/common"
	"FGG-Service/src/timers/service"
	"FGG-Service/src/timers/types"
	"FGG-Service/tests/timers/mock/dbtimers"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

// An existing timer is returned with its derived remaining time.
func TestSrvTimers_GetCurrentTimer_ExistingTimer(test *testing.T) {
	// Arrange
	timerDb := new(dbtimermock.DatabaseMock)
	timerDb.On("GetCurrentTimerCommand", 1, 1).Return(existingCurrentTimer, nil)

	sut := srvtimers.Service{Database: timerDb}

	// Act
	timer, err := sut.GetCurrentTimer(1)

	// Assert
	require.NoError(test, err)
	require.Equal(test, existingTimer, timer)
}

// Without a timer the CURRENT_TIMER_NOT_FOUND error returns and nothing is created.
func TestSrvTimers_GetCurrentTimer_NoTimer_NotFound(test *testing.T) {
	// Arrange
	timerDb := new(dbtimermock.DatabaseMock)
	timerDb.On("GetCurrentTimerCommand", 1, 1).Return(typetimers.CurrentTimer{}, sql.ErrNoRows)

	sut := srvtimers.Service{Database: timerDb}

	// Act
	_, err := sut.GetCurrentTimer(1)

	// Assert
	var notFound *common.NotFoundError
	require.ErrorAs(test, err, &notFound)
	require.Equal(test, "CURRENT_TIMER_NOT_FOUND", notFound.GetCode())
	timerDb.AssertNotCalled(test, "CreateCurrentTimerCommand")
}

// A database error returns as-is.
func TestSrvTimers_GetCurrentTimer_DatabaseError(test *testing.T) {
	// Arrange
	timerDb := new(dbtimermock.DatabaseMock)
	timerDb.On("GetCurrentTimerCommand", 1, 1).Return(typetimers.CurrentTimer{}, dbError)

	sut := srvtimers.Service{Database: timerDb}

	// Act
	_, err := sut.GetCurrentTimer(1)

	// Assert
	require.ErrorIs(test, err, dbError)
}
