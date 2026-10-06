package srvpoints_test

import (
	"FGG-Service/src/common"
	"FGG-Service/src/points/service"
	"FGG-Service/tests/points/mock"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

// A direct change that does not fit the INTEGER columns it is recorded in is rejected as
// unprocessable before anything is looked up or written.
func TestSrvPoints_ChangeUserPointByTypeName_AmountOutOfIntegerRange_Unprocessable(test *testing.T) {
	// Arrange
	databaseMock := new(dbpointsmock.DatabaseMock)
	sut := srvpoints.Service{Database: databaseMock}

	// Act
	_, err := sut.ChangeUserPointByTypeName(test.Context(), 9, 2, 1, "Rolls", math.MaxInt32+1)

	// Assert
	var unprocessable *common.UnprocessableError
	require.ErrorAs(test, err, &unprocessable)
	databaseMock.AssertNotCalled(test, "GetPointTypeByNameCommand")
}

func TestSrvPoints_ChangePartyPointByTypeName_AmountOutOfIntegerRange_Unprocessable(test *testing.T) {
	// Arrange
	databaseMock := new(dbpointsmock.DatabaseMock)
	sut := srvpoints.Service{Database: databaseMock}

	// Act
	_, err := sut.ChangePartyPointByTypeName(test.Context(), 9, 1, "Rolls", math.MinInt32-1)

	// Assert
	var unprocessable *common.UnprocessableError
	require.ErrorAs(test, err, &unprocessable)
	databaseMock.AssertNotCalled(test, "GetPointTypeByNameCommand")
}
