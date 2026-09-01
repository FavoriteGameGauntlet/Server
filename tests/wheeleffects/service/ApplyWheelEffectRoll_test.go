package srvwheeleffects_test

import (
	"FGG-Service/src/changes/types"
	"FGG-Service/src/common"
	srvwheeleffects "FGG-Service/src/wheeleffects/service"
	typewheeleffects "FGG-Service/src/wheeleffects/types"
	dbchangesmock "FGG-Service/tests/changes/mock"
	srvchangesmock "FGG-Service/tests/changes/srvmock"
	dbwheeleffectsmock "FGG-Service/tests/timers/mock/dbwheeleffects"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

var lastRolledRow = typewheeleffects.LastWheelRow{Id: 42, Name: "test-effect", Description: "desc"}

var wheelRowDetail = typewheeleffects.WheelRow{
	Id: 42, PartyId: 1, Name: "test-effect", Description: "desc", ChangeId: 100, CollectionId: 1,
}

var applyTemplateEntries = []typechanges.ChangeEntry{
	{Amount: 5, PointTypeId: ptr(7)},
}

var createdWheelRowHistory = typewheeleffects.CreatedWheelRowHistory{Id: 999, UserId: 1, PartyId: 1, WheelRowId: 42}

func setupApplyUpToChangeEntries(databaseMock *dbwheeleffectsmock.DatabaseMock, changesDbMock *dbchangesmock.DatabaseMock) {
	databaseMock.On("GetLastRolledWheelEffectsCommand", 1, 1).Return([]typewheeleffects.LastWheelRow{lastRolledRow}, nil)
	databaseMock.On("GetEffectHistoryByEffectNameCommand", 1, 1, "test-effect").Return(typewheeleffects.WheelRowHistory{}, sql.ErrNoRows)
	databaseMock.On("GetWheelRowCommand", 1, 42).Return(wheelRowDetail, nil)
	databaseMock.On("AddWheelEffectHistoryCommand", 1, 1, 42, (*int)(nil)).Return(createdWheelRowHistory, nil)
	changesDbMock.On("GetChangeEntriesJsonbCommand", 1, 100).Return(applyTemplateEntries, nil)
}

type ApplyWheelEffectRollTestCase struct {
	Name              string
	UserId            int
	WheelRowName      string
	TargetUserIds     []int
	SetupMocks        func() (*dbwheeleffectsmock.DatabaseMock, *dbchangesmock.DatabaseMock, *srvchangesmock.ServiceMock)
	ExpectedError     error
	ExpectedErrorCode string
}

var ApplyWheelEffectRollTestCases = []ApplyWheelEffectRollTestCase{
	{
		// The name doesn't match any last-rolled row. The error returns and nothing else happens.
		Name:          "NameNotFound_Error",
		UserId:        1,
		WheelRowName:  "unknown-effect",
		TargetUserIds: []int{2},
		SetupMocks: func() (*dbwheeleffectsmock.DatabaseMock, *dbchangesmock.DatabaseMock, *srvchangesmock.ServiceMock) {
			databaseMock := new(dbwheeleffectsmock.DatabaseMock)
			changesDbMock := new(dbchangesmock.DatabaseMock)
			changesSvcMock := new(srvchangesmock.ServiceMock)

			databaseMock.On("GetLastRolledWheelEffectsCommand", 1, 1).Return([]typewheeleffects.LastWheelRow{lastRolledRow}, nil)

			return databaseMock, changesDbMock, changesSvcMock
		},
		ExpectedErrorCode: "WHEEL_EFFECT_NAME_NOT_FOUND",
	},
	{
		// The row was already applied. The error returns and nothing else happens.
		Name:          "AlreadyApplied_Error",
		UserId:        1,
		WheelRowName:  "test-effect",
		TargetUserIds: []int{2},
		SetupMocks: func() (*dbwheeleffectsmock.DatabaseMock, *dbchangesmock.DatabaseMock, *srvchangesmock.ServiceMock) {
			databaseMock := new(dbwheeleffectsmock.DatabaseMock)
			changesDbMock := new(dbchangesmock.DatabaseMock)
			changesSvcMock := new(srvchangesmock.ServiceMock)

			databaseMock.On("GetLastRolledWheelEffectsCommand", 1, 1).Return([]typewheeleffects.LastWheelRow{lastRolledRow}, nil)
			databaseMock.On("GetEffectHistoryByEffectNameCommand", 1, 1, "test-effect").Return(typewheeleffects.WheelRowHistory{Name: "test-effect"}, nil)

			return databaseMock, changesDbMock, changesSvcMock
		},
		ExpectedErrorCode: "WHEEL_EFFECT_ROLL_ALREADY_APPLIED",
	},
	{
		// GetWheelRowCommand fails. The error returns.
		Name:          "GetWheelRow_Error",
		UserId:        1,
		WheelRowName:  "test-effect",
		TargetUserIds: []int{2},
		SetupMocks: func() (*dbwheeleffectsmock.DatabaseMock, *dbchangesmock.DatabaseMock, *srvchangesmock.ServiceMock) {
			databaseMock := new(dbwheeleffectsmock.DatabaseMock)
			changesDbMock := new(dbchangesmock.DatabaseMock)
			changesSvcMock := new(srvchangesmock.ServiceMock)

			databaseMock.On("GetLastRolledWheelEffectsCommand", 1, 1).Return([]typewheeleffects.LastWheelRow{lastRolledRow}, nil)
			databaseMock.On("GetEffectHistoryByEffectNameCommand", 1, 1, "test-effect").Return(typewheeleffects.WheelRowHistory{}, sql.ErrNoRows)
			databaseMock.On("GetWheelRowCommand", 1, 42).Return(typewheeleffects.WheelRow{}, dbError)

			return databaseMock, changesDbMock, changesSvcMock
		},
		ExpectedError: dbError,
	},
	{
		// AddWheelEffectHistoryCommand fails. The error returns.
		Name:          "AddWheelEffectHistory_Error",
		UserId:        1,
		WheelRowName:  "test-effect",
		TargetUserIds: []int{2},
		SetupMocks: func() (*dbwheeleffectsmock.DatabaseMock, *dbchangesmock.DatabaseMock, *srvchangesmock.ServiceMock) {
			databaseMock := new(dbwheeleffectsmock.DatabaseMock)
			changesDbMock := new(dbchangesmock.DatabaseMock)
			changesSvcMock := new(srvchangesmock.ServiceMock)

			databaseMock.On("GetLastRolledWheelEffectsCommand", 1, 1).Return([]typewheeleffects.LastWheelRow{lastRolledRow}, nil)
			databaseMock.On("GetEffectHistoryByEffectNameCommand", 1, 1, "test-effect").Return(typewheeleffects.WheelRowHistory{}, sql.ErrNoRows)
			databaseMock.On("GetWheelRowCommand", 1, 42).Return(wheelRowDetail, nil)
			databaseMock.On("AddWheelEffectHistoryCommand", 1, 1, 42, (*int)(nil)).Return(typewheeleffects.CreatedWheelRowHistory{}, dbError)

			return databaseMock, changesDbMock, changesSvcMock
		},
		ExpectedError: dbError,
	},
	{
		// GetChangeEntriesJsonbCommand fails. The error returns.
		Name:          "GetChangeEntriesJsonb_Error",
		UserId:        1,
		WheelRowName:  "test-effect",
		TargetUserIds: []int{2},
		SetupMocks: func() (*dbwheeleffectsmock.DatabaseMock, *dbchangesmock.DatabaseMock, *srvchangesmock.ServiceMock) {
			databaseMock := new(dbwheeleffectsmock.DatabaseMock)
			changesDbMock := new(dbchangesmock.DatabaseMock)
			changesSvcMock := new(srvchangesmock.ServiceMock)

			databaseMock.On("GetLastRolledWheelEffectsCommand", 1, 1).Return([]typewheeleffects.LastWheelRow{lastRolledRow}, nil)
			databaseMock.On("GetEffectHistoryByEffectNameCommand", 1, 1, "test-effect").Return(typewheeleffects.WheelRowHistory{}, sql.ErrNoRows)
			databaseMock.On("GetWheelRowCommand", 1, 42).Return(wheelRowDetail, nil)
			databaseMock.On("AddWheelEffectHistoryCommand", 1, 1, 42, (*int)(nil)).Return(createdWheelRowHistory, nil)
			changesDbMock.On("GetChangeEntriesJsonbCommand", 1, 100).Return([]typechanges.ChangeEntry(nil), dbError)

			return databaseMock, changesDbMock, changesSvcMock
		},
		ExpectedError: dbError,
	},
	{
		// CreateUserChangeFromJsonbCommand fails. The error returns.
		Name:          "CreateUserChangeFromJsonb_Error",
		UserId:        1,
		WheelRowName:  "test-effect",
		TargetUserIds: []int{2},
		SetupMocks: func() (*dbwheeleffectsmock.DatabaseMock, *dbchangesmock.DatabaseMock, *srvchangesmock.ServiceMock) {
			databaseMock := new(dbwheeleffectsmock.DatabaseMock)
			changesDbMock := new(dbchangesmock.DatabaseMock)
			changesSvcMock := new(srvchangesmock.ServiceMock)

			setupApplyUpToChangeEntries(databaseMock, changesDbMock)

			expectedEntries := []typechanges.ChangeEntry{{Amount: 5, PointTypeId: ptr(7), UserId: ptr(2)}}
			changesDbMock.On("CreateUserChangeFromJsonbCommand", 1, expectedEntries).Return(typechanges.UserChange{}, dbError)

			return databaseMock, changesDbMock, changesSvcMock
		},
		ExpectedError: dbError,
	},
	{
		// ApplyChangeEntries fails. The error returns.
		Name:          "ApplyChangeEntries_Error",
		UserId:        1,
		WheelRowName:  "test-effect",
		TargetUserIds: []int{2},
		SetupMocks: func() (*dbwheeleffectsmock.DatabaseMock, *dbchangesmock.DatabaseMock, *srvchangesmock.ServiceMock) {
			databaseMock := new(dbwheeleffectsmock.DatabaseMock)
			changesDbMock := new(dbchangesmock.DatabaseMock)
			changesSvcMock := new(srvchangesmock.ServiceMock)

			setupApplyUpToChangeEntries(databaseMock, changesDbMock)

			expectedEntries := []typechanges.ChangeEntry{{Amount: 5, PointTypeId: ptr(7), UserId: ptr(2)}}
			createdUserChange := typechanges.UserChange{ChangeId: ptr(200), Entries: expectedEntries}
			changesDbMock.On("CreateUserChangeFromJsonbCommand", 1, expectedEntries).Return(createdUserChange, nil)
			changesSvcMock.On("ApplyChangeEntries", 1, expectedEntries, 999).Return(dbError)

			return databaseMock, changesDbMock, changesSvcMock
		},
		ExpectedError: dbError,
	},
	{
		// Every step succeeds targeting a single user. No error returns.
		Name:          "Success_SingleTarget",
		UserId:        1,
		WheelRowName:  "test-effect",
		TargetUserIds: []int{2},
		SetupMocks: func() (*dbwheeleffectsmock.DatabaseMock, *dbchangesmock.DatabaseMock, *srvchangesmock.ServiceMock) {
			databaseMock := new(dbwheeleffectsmock.DatabaseMock)
			changesDbMock := new(dbchangesmock.DatabaseMock)
			changesSvcMock := new(srvchangesmock.ServiceMock)

			setupApplyUpToChangeEntries(databaseMock, changesDbMock)

			expectedEntries := []typechanges.ChangeEntry{{Amount: 5, PointTypeId: ptr(7), UserId: ptr(2)}}
			createdUserChange := typechanges.UserChange{ChangeId: ptr(200), Entries: expectedEntries}
			changesDbMock.On("CreateUserChangeFromJsonbCommand", 1, expectedEntries).Return(createdUserChange, nil)
			changesSvcMock.On("ApplyChangeEntries", 1, expectedEntries, 999).Return(nil)

			return databaseMock, changesDbMock, changesSvcMock
		},
	},
	{
		// Every step succeeds targeting multiple users. The template entries are copied per target,
		// each one stamped with its own UserId and a cleared EntryId, in target order.
		Name:          "Success_MultipleTargets",
		UserId:        1,
		WheelRowName:  "test-effect",
		TargetUserIds: []int{2, 3},
		SetupMocks: func() (*dbwheeleffectsmock.DatabaseMock, *dbchangesmock.DatabaseMock, *srvchangesmock.ServiceMock) {
			databaseMock := new(dbwheeleffectsmock.DatabaseMock)
			changesDbMock := new(dbchangesmock.DatabaseMock)
			changesSvcMock := new(srvchangesmock.ServiceMock)

			setupApplyUpToChangeEntries(databaseMock, changesDbMock)

			expectedEntries := []typechanges.ChangeEntry{
				{Amount: 5, PointTypeId: ptr(7), UserId: ptr(2)},
				{Amount: 5, PointTypeId: ptr(7), UserId: ptr(3)},
			}
			createdUserChange := typechanges.UserChange{ChangeId: ptr(200), Entries: expectedEntries}
			changesDbMock.On("CreateUserChangeFromJsonbCommand", 1, expectedEntries).Return(createdUserChange, nil)
			changesSvcMock.On("ApplyChangeEntries", 1, expectedEntries, 999).Return(nil)

			return databaseMock, changesDbMock, changesSvcMock
		},
	},
}

func TestSrvWheelEffects_ApplyWheelEffectRoll(test *testing.T) {
	for _, testCase := range ApplyWheelEffectRollTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			weDatabaseMock, changesDbMock, changesSvcMock := testCase.SetupMocks()

			sut := srvwheeleffects.Service{
				Database:        weDatabaseMock,
				ChangesDatabase: changesDbMock,
				ChangesService:  changesSvcMock,
			}

			// Act
			err := sut.ApplyWheelEffectRoll(testCase.UserId, testCase.WheelRowName, testCase.TargetUserIds)

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
			}

			weDatabaseMock.AssertExpectations(test)
			changesDbMock.AssertExpectations(test)
			changesSvcMock.AssertExpectations(test)
		})
	}
}
