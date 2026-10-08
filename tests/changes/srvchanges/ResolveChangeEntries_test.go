package srvchanges_test

import (
	"FGG-Service/src/changes/srvchanges"
	"FGG-Service/src/changes/typechanges"
	"FGG-Service/src/common"
	"FGG-Service/src/points/srvpoints"
	"FGG-Service/tests/points/dbpointsmock"
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
		_, err := sut.ResolveChangeEntries(test.Context(), 1, []typechanges.NamedChangeEntry{{PointTypeName: &name, Amount: amount}})

		// Assert
		var unprocessable *common.UnprocessableError
		require.ErrorAs(test, err, &unprocessable)
		pointsDb.AssertNotCalled(test, "GetPointTypeByNameCommand")
	}
}
