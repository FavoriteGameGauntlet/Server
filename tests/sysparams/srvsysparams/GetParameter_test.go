package srvsysparams_test

import (
	"FGG-Service/src/common"
	"FGG-Service/src/sysparams/srvsysparams"
	"FGG-Service/src/sysparams/typesysparams"
	"FGG-Service/tests/sysparams/dbsysparamsmock"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

type GetParameterTestCase struct {
	Name              string
	ParameterName     string
	SetupMock         func() *dbsysparamsmock.DatabaseMock
	ExpectedParameter *typesysparams.SystemParameter
	ExpectedError     error
	ExpectedErrorCode string
}

var GetParameterTestCases = []GetParameterTestCase{
	{
		// GetSystemParameterCommand returns sql.ErrNoRows. The SYSTEM_PARAMETER_NOT_FOUND error returns.
		Name:          "NoRows_NotFoundError",
		ParameterName: "DefaultTimerDurationInS",
		SetupMock: func() *dbsysparamsmock.DatabaseMock {
			databaseMock := new(dbsysparamsmock.DatabaseMock)

			databaseMock.
				On("GetSystemParameterCommand", 1, "DefaultTimerDurationInS").
				Return(typesysparams.SystemParameter{}, sql.ErrNoRows)

			return databaseMock
		},
		ExpectedErrorCode: "SYSTEM_PARAMETER_NOT_FOUND",
	},
	{
		// GetSystemParameterCommand returns a database error. The error returns as-is.
		Name:          "DatabaseError",
		ParameterName: "DefaultTimerDurationInS",
		SetupMock: func() *dbsysparamsmock.DatabaseMock {
			databaseMock := new(dbsysparamsmock.DatabaseMock)

			databaseMock.
				On("GetSystemParameterCommand", 1, "DefaultTimerDurationInS").
				Return(typesysparams.SystemParameter{}, dbError)

			return databaseMock
		},
		ExpectedError: dbError,
	},
	{
		// GetSystemParameterCommand succeeds. The parameter returns.
		Name:          "Success",
		ParameterName: "DefaultTimerDurationInS",
		SetupMock: func() *dbsysparamsmock.DatabaseMock {
			databaseMock := new(dbsysparamsmock.DatabaseMock)

			databaseMock.
				On("GetSystemParameterCommand", 1, "DefaultTimerDurationInS").
				Return(typesysparams.SystemParameter{Id: 1, Name: "DefaultTimerDurationInS", Value: "30"}, nil)

			return databaseMock
		},
		ExpectedParameter: &typesysparams.SystemParameter{Id: 1, Name: "DefaultTimerDurationInS", Value: "30"},
	},
}

func TestSrvSysParams_GetParameter(test *testing.T) {
	for _, testCase := range GetParameterTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			databaseMock := testCase.SetupMock()
			sut := srvsysparams.Service{Database: databaseMock}

			// Act
			parameter, err := sut.GetParameter(test.Context(), 1, testCase.ParameterName)

			// Assert
			if testCase.ExpectedErrorCode != "" {
				require.Error(test, err)
				appErr, ok := err.(common.AppError)
				require.True(test, ok)
				require.Equal(test, testCase.ExpectedErrorCode, appErr.GetCode())
			} else if testCase.ExpectedError != nil {
				require.ErrorIs(test, err, testCase.ExpectedError)
			} else {
				require.NoError(test, err)
				require.Equal(test, *testCase.ExpectedParameter, parameter)
			}

			databaseMock.AssertExpectations(test)
		})
	}
}
