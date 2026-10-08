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

type ChangeValueTestCase struct {
	Name              string
	ParameterName     string
	Value             string
	SetupMock         func() *dbsysparamsmock.DatabaseMock
	ExpectedError     error
	ExpectedErrorCode string
	ExpectedCreated   bool
}

var ChangeValueTestCases = []ChangeValueTestCase{
	{
		// GetSystemParameterCommand returns sql.ErrNoRows. The SYSTEM_PARAMETER_NOT_FOUND error returns and the value isn't changed.
		Name:          "NoRows_NotFoundError",
		ParameterName: "DefaultTimerDurationInS",
		Value:         "60",
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
		// GetSystemParameterCommand returns a database error. The error returns as-is and the value isn't changed.
		Name:          "GetDatabaseError",
		ParameterName: "DefaultTimerDurationInS",
		Value:         "60",
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
		// The parameter exists but ChangeSystemParameterValueCommand returns a database error. The error returns as-is.
		Name:          "ChangeDatabaseError",
		ParameterName: "DefaultTimerDurationInS",
		Value:         "60",
		SetupMock: func() *dbsysparamsmock.DatabaseMock {
			databaseMock := new(dbsysparamsmock.DatabaseMock)

			databaseMock.
				On("GetSystemParameterCommand", 1, "DefaultTimerDurationInS").
				Return(typesysparams.SystemParameter{Id: 1, Name: "DefaultTimerDurationInS", Value: "30"}, nil)

			databaseMock.
				On("ChangeSystemParameterValueCommand", 1, 1, "60").
				Return(false, dbError)

			return databaseMock
		},
		ExpectedError: dbError,
	},
	{
		// The party has no override yet, so one is created. created is true.
		Name:          "OverrideCreated",
		ParameterName: "DefaultTimerDurationInS",
		Value:         "60",
		SetupMock: func() *dbsysparamsmock.DatabaseMock {
			databaseMock := new(dbsysparamsmock.DatabaseMock)

			databaseMock.
				On("GetSystemParameterCommand", 1, "DefaultTimerDurationInS").
				Return(typesysparams.SystemParameter{Id: 1, Name: "DefaultTimerDurationInS", Value: "30"}, nil)

			databaseMock.
				On("ChangeSystemParameterValueCommand", 1, 1, "60").
				Return(true, nil)

			return databaseMock
		},
		ExpectedCreated: true,
	},
	{
		// The party already overrides the parameter, so the override changes. created is false.
		Name:          "OverrideChanged",
		ParameterName: "DefaultTimerDurationInS",
		Value:         "60",
		SetupMock: func() *dbsysparamsmock.DatabaseMock {
			databaseMock := new(dbsysparamsmock.DatabaseMock)

			databaseMock.
				On("GetSystemParameterCommand", 1, "DefaultTimerDurationInS").
				Return(typesysparams.SystemParameter{Id: 1, Name: "DefaultTimerDurationInS", Value: "30"}, nil)

			databaseMock.
				On("ChangeSystemParameterValueCommand", 1, 1, "60").
				Return(false, nil)

			return databaseMock
		},
	},
}

func TestSrvSysParams_ChangeValue(test *testing.T) {
	for _, testCase := range ChangeValueTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			databaseMock := testCase.SetupMock()
			sut := srvsysparams.Service{Database: databaseMock}

			// Act
			created, err := sut.ChangeValue(test.Context(), 1, testCase.ParameterName, testCase.Value)

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
				require.Equal(test, testCase.ExpectedCreated, created)
			}

			databaseMock.AssertExpectations(test)
		})
	}
}
