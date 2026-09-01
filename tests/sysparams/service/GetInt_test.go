package srvsysparams_test

import (
	"FGG-Service/src/common"
	"FGG-Service/src/sysparams/service"
	"FGG-Service/src/sysparams/types"
	"FGG-Service/tests/sysparams/mock"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

func ptrInt(i int) *int { return &i }

type GetIntTestCase struct {
	Name              string
	ParameterName     string
	SetupMock         func() *dbsysparamsmock.DatabaseMock
	ExpectedValue     *int
	ExpectedError     error
	ExpectAnyError    bool
	ExpectedErrorCode string
}

var GetIntTestCases = []GetIntTestCase{
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
		// The stored value isn't a valid integer. A parsing error returns.
		Name:          "NotAnInteger_Error",
		ParameterName: "DefaultTimerDurationInS",
		SetupMock: func() *dbsysparamsmock.DatabaseMock {
			databaseMock := new(dbsysparamsmock.DatabaseMock)

			databaseMock.
				On("GetSystemParameterCommand", 1, "DefaultTimerDurationInS").
				Return(typesysparams.SystemParameter{Id: 1, Name: "DefaultTimerDurationInS", Value: "not-a-number"}, nil)

			return databaseMock
		},
		ExpectAnyError: true,
	},
	{
		// GetSystemParameterCommand succeeds. The parsed integer value returns.
		Name:          "Success",
		ParameterName: "DefaultTimerDurationInS",
		SetupMock: func() *dbsysparamsmock.DatabaseMock {
			databaseMock := new(dbsysparamsmock.DatabaseMock)

			databaseMock.
				On("GetSystemParameterCommand", 1, "DefaultTimerDurationInS").
				Return(typesysparams.SystemParameter{Id: 1, Name: "DefaultTimerDurationInS", Value: "30"}, nil)

			return databaseMock
		},
		ExpectedValue: ptrInt(30),
	},
}

func TestSrvSysParams_GetInt(test *testing.T) {
	for _, testCase := range GetIntTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			databaseMock := testCase.SetupMock()
			sut := srvsysparams.Service{Database: databaseMock}

			// Act
			value, err := sut.GetInt(testCase.ParameterName)

			// Assert
			if testCase.ExpectedErrorCode != "" {
				require.Error(test, err)
				appErr, ok := err.(common.AppError)
				require.True(test, ok)
				require.Equal(test, testCase.ExpectedErrorCode, appErr.GetCode())
			} else if testCase.ExpectedError != nil {
				require.ErrorIs(test, err, testCase.ExpectedError)
			} else if testCase.ExpectAnyError {
				require.Error(test, err)
			} else {
				require.NoError(test, err)
				require.Equal(test, *testCase.ExpectedValue, value)
			}

			databaseMock.AssertExpectations(test)
		})
	}
}
