package srvchanges_test

import (
	srvchanges "FGG-Service/src/changes/service"
	"FGG-Service/src/changes/types"
	"FGG-Service/src/common"
	srvpoints "FGG-Service/src/points/service"
	dbpointsmock "FGG-Service/tests/points/mock"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

// An amount that does not fit the INTEGER column it is stored in is rejected as unprocessable
// before the entry's target is even looked up.
func TestSrvChanges_ResolveChangeEntries_AmountOutOfIntegerRange_Unprocessable(test *testing.T) {
	for _, amount := range []int{math.MaxInt32 + 1, math.MinInt32 - 1} {
		// Arrange
		pointsDb := new(dbpointsmock.DatabaseMock)
		sut := srvchanges.Service{PointsService: &srvpoints.Service{Database: pointsDb}}

		name := "Rolls"

		// Act
		_, err := sut.ResolveChangeEntries(1, []typechanges.ChangeEntryInput{{PointTypeName: &name, Amount: amount}})

		// Assert
		var unprocessable *common.UnprocessableError
		require.ErrorAs(test, err, &unprocessable)
		pointsDb.AssertNotCalled(test, "GetPointTypeByNameCommand")
	}
}
