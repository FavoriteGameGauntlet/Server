package srvpoints_test

import (
	"FGG-Service/src/common"
	"FGG-Service/src/points/service"
	"FGG-Service/src/points/type"
	"FGG-Service/tests/points/mock"
	"database/sql"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

// --- CreatePointType ---

type CreatePointTypeBoundsTestCase struct {
	Name       string
	StartValue int
	Minimum    *int
	Maximum    *int
}

var CreatePointTypeInvalidBoundsTestCases = []CreatePointTypeBoundsTestCase{
	// The start value is below the minimum.
	{Name: "StartValueBelowMinimum", StartValue: 0, Minimum: ptrInt(1), Maximum: ptrInt(1000)},
	// The start value is above the maximum.
	{Name: "StartValueAboveMaximum", StartValue: 11, Minimum: ptrInt(0), Maximum: ptrInt(10)},
	// The minimum is greater than the maximum.
	{Name: "MinimumAboveMaximum", StartValue: 5, Minimum: ptrInt(10), Maximum: ptrInt(0)},
	// Without a maximum the minimum still limits the start value.
	{Name: "NoMaximum_StartValueBelowMinimum", StartValue: 0, Minimum: ptrInt(1)},
	// Without a minimum the maximum still limits the start value.
	{Name: "NoMinimum_StartValueAboveMaximum", StartValue: 11, Maximum: ptrInt(10)},
	// A start value that does not fit an INTEGER column.
	{Name: "StartValueAboveIntegerRange", StartValue: math.MaxInt32 + 1},
	// A maximum that does not fit an INTEGER column.
	{Name: "MaximumAboveIntegerRange", StartValue: 0, Maximum: ptrInt(math.MaxInt32 + 1)},
	// A minimum that does not fit an INTEGER column.
	{Name: "MinimumBelowIntegerRange", StartValue: 0, Minimum: ptrInt(math.MinInt32 - 1)},
}

// Bounds the schema would reject are rejected as unprocessable before anything is written.
func TestSrvPoints_CreatePointType_InvalidBounds(test *testing.T) {
	for _, testCase := range CreatePointTypeInvalidBoundsTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			databaseMock := new(dbpointsmock.DatabaseMock)
			sut := srvpoints.Service{Database: databaseMock}

			// Act
			_, err := sut.CreatePointType(1, typepoints.PointType{
				Name:       "Rolls",
				StartValue: testCase.StartValue,
				Minimum:    testCase.Minimum,
				Maximum:    testCase.Maximum,
			})

			// Assert
			var unprocessable *common.UnprocessableError
			require.ErrorAs(test, err, &unprocessable)
			databaseMock.AssertNotCalled(test, "CreatePointTypeCommand")
		})
	}
}

var CreatePointTypeValidBoundsTestCases = []CreatePointTypeBoundsTestCase{
	// A start value on either bound is valid.
	{Name: "StartValueOnBound", StartValue: 1, Minimum: ptrInt(1), Maximum: ptrInt(1000)},
	// A missing maximum does not limit the start value from above.
	{Name: "NoMaximum", StartValue: 5000, Minimum: ptrInt(0)},
	// A missing minimum does not limit the start value from below.
	{Name: "NoMinimum", StartValue: -5, Maximum: ptrInt(10)},
	// Without either bound any start value is valid.
	{Name: "NoBounds", StartValue: -5},
}

func TestSrvPoints_CreatePointType_ValidBounds_Created(test *testing.T) {
	for _, testCase := range CreatePointTypeValidBoundsTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			databaseMock := new(dbpointsmock.DatabaseMock)
			databaseMock.On("GetPointTypeByNameCommand", 1, "Rolls").Return(typepoints.PointTypeInfo{}, sql.ErrNoRows)
			databaseMock.On("CreatePointTypeCommand", 1, "Rolls", "", testCase.StartValue, true, false, testCase.Minimum, testCase.Maximum).
				Return(typepoints.PointType{Name: "Rolls"}, nil)

			sut := srvpoints.Service{Database: databaseMock}

			// Act
			_, err := sut.CreatePointType(1, typepoints.PointType{
				Name:       "Rolls",
				StartValue: testCase.StartValue,
				IsPublic:   true,
				Minimum:    testCase.Minimum,
				Maximum:    testCase.Maximum,
			})

			// Assert
			require.NoError(test, err)
			databaseMock.AssertExpectations(test)
		})
	}
}

// --- ChangePointType ---

// New bounds have to keep the existing start value, which the change cannot touch.
func TestSrvPoints_ChangePointType_BoundsExcludeStartValue_Unprocessable(test *testing.T) {
	// Arrange
	databaseMock := new(dbpointsmock.DatabaseMock)
	databaseMock.On("GetPointTypeByNameCommand", 1, "Rolls").Return(typepoints.PointTypeInfo{
		Id: 5, PartyId: 1, Name: "Rolls", StartValue: 0, Minimum: ptrInt(0), Maximum: ptrInt(10),
	}, nil)

	sut := srvpoints.Service{Database: databaseMock}

	// Act
	err := sut.ChangePointType(1, "Rolls", typepoints.PointType{Minimum: ptrInt(1), Maximum: ptrInt(10)})

	// Assert
	var unprocessable *common.UnprocessableError
	require.ErrorAs(test, err, &unprocessable)
	databaseMock.AssertNotCalled(test, "ChangePointTypeCommand")
}

// Removing the maximum keeps the existing start value, so the change is saved.
func TestSrvPoints_ChangePointType_RemoveMaximum_Changed(test *testing.T) {
	// Arrange
	databaseMock := new(dbpointsmock.DatabaseMock)
	databaseMock.On("GetPointTypeByNameCommand", 1, "Rolls").Return(typepoints.PointTypeInfo{
		Id: 5, PartyId: 1, Name: "Rolls", StartValue: 3, Minimum: ptrInt(0), Maximum: ptrInt(10),
	}, nil)
	databaseMock.On("ChangePointTypeCommand", 1, 5, "Rolls", "", false, false, ptrInt(3), (*int)(nil)).Return(nil)

	sut := srvpoints.Service{Database: databaseMock}

	// Act
	err := sut.ChangePointType(1, "Rolls", typepoints.PointType{Minimum: ptrInt(3)})

	// Assert
	require.NoError(test, err)
	databaseMock.AssertExpectations(test)
}
