package srvsysparams_test

import (
	"FGG-Service/src/sysparams/service"
	"FGG-Service/src/sysparams/types"
	"FGG-Service/tests/sysparams/mock"
	"testing"

	"github.com/stretchr/testify/require"
)

type GetAllTestCase struct {
	Name               string
	SetupMock          func() *dbsysparamsmock.DatabaseMock
	ExpectedParameters []typesysparams.SystemParameter
	ExpectedErrorIs    error
}

var GetAllTestCases = []GetAllTestCase{
	{
		// GetAllSystemParametersCommand returns a database error. The error returns.
		Name: "DatabaseError",
		SetupMock: func() *dbsysparamsmock.DatabaseMock {
			databaseMock := new(dbsysparamsmock.DatabaseMock)

			databaseMock.
				On("GetAllSystemParametersCommand", 1).
				Return([]typesysparams.SystemParameter{}, dbError)

			return databaseMock
		},
		ExpectedErrorIs: dbError,
	},
	{
		// GetAllSystemParametersCommand returns an empty list. An empty list returns.
		Name: "EmptyList",
		SetupMock: func() *dbsysparamsmock.DatabaseMock {
			databaseMock := new(dbsysparamsmock.DatabaseMock)

			databaseMock.
				On("GetAllSystemParametersCommand", 1).
				Return([]typesysparams.SystemParameter{}, nil)

			return databaseMock
		},
		ExpectedParameters: []typesysparams.SystemParameter{},
	},
	{
		// GetAllSystemParametersCommand returns a populated list. The list returns.
		Name: "Success",
		SetupMock: func() *dbsysparamsmock.DatabaseMock {
			databaseMock := new(dbsysparamsmock.DatabaseMock)

			databaseMock.
				On("GetAllSystemParametersCommand", 1).
				Return([]typesysparams.SystemParameter{
					{Id: 1, Name: "DefaultTimerDurationInS", Value: "30"},
					{Id: 2, Name: "DefaultTerritoryHoursIncreasing", Value: "2"},
				}, nil)

			return databaseMock
		},
		ExpectedParameters: []typesysparams.SystemParameter{
			{Id: 1, Name: "DefaultTimerDurationInS", Value: "30"},
			{Id: 2, Name: "DefaultTerritoryHoursIncreasing", Value: "2"},
		},
	},
}

func TestSrvSysParams_GetAll(test *testing.T) {
	for _, testCase := range GetAllTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			databaseMock := testCase.SetupMock()
			sut := srvsysparams.Service{Database: databaseMock}

			// Act
			parameters, err := sut.GetAll()

			// Assert
			if testCase.ExpectedErrorIs != nil {
				require.ErrorIs(test, err, testCase.ExpectedErrorIs)
			}

			if testCase.ExpectedParameters != nil {
				require.NoError(test, err)
				require.Equal(test, testCase.ExpectedParameters, parameters)
			}

			databaseMock.AssertExpectations(test)
		})
	}
}
