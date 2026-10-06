package srvwheeleffects_test

import (
	"FGG-Service/src/common"
	"FGG-Service/src/sysparams/types"
	srvwheeleffects "FGG-Service/src/wheeleffects/service"
	typewheeleffects "FGG-Service/src/wheeleffects/types"
	srvsysparamsmock "FGG-Service/tests/sysparams/srvmock"
	dbwheeleffectsmock "FGG-Service/tests/timers/mock/dbwheeleffects"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var candidateRows = []typewheeleffects.WheelRow{
	{Id: 10, PartyId: 1, Name: "row-a", Description: "desc a", ChangeId: 100, CollectionId: 1},
	{Id: 11, PartyId: 1, Name: "row-b", Description: "desc b", ChangeId: 101, CollectionId: 1},
	{Id: 12, PartyId: 1, Name: "row-c", Description: "desc c", ChangeId: 102, CollectionId: 1},
}

var rolledResult = []typewheeleffects.LastWheelRow{
	{Id: 10, Name: "row-a", Description: "desc a", WheelPosition: 1},
	{Id: 11, Name: "row-b", Description: "desc b", WheelPosition: 2},
	{Id: 12, Name: "row-c", Description: "desc c", WheelPosition: 3},
}

// matchRolledRows matches an AddLastRolledWheelEffectsCommand call whose rows carry exactly
// expectedIds (order-agnostic, since MakeEffectRoll shuffles candidates) at positions 1..N.
func matchRolledRows(expectedIds ...int) any {
	return mock.MatchedBy(func(rows []typewheeleffects.RolledWheelRowInput) bool {
		if len(rows) != len(expectedIds) {
			return false
		}

		positionByRowId := map[int]int{}
		for _, row := range rows {
			positionByRowId[row.WheelRowId] = row.Position
		}

		seenPositions := map[int]bool{}
		for _, id := range expectedIds {
			position, ok := positionByRowId[id]
			if !ok {
				return false
			}
			seenPositions[position] = true
		}

		for i := 1; i <= len(expectedIds); i++ {
			if !seenPositions[i] {
				return false
			}
		}

		return true
	})
}

type MakeEffectRollTestCase struct {
	Name                 string
	UserId               int
	SetupWheelEffectMock func() *dbwheeleffectsmock.DatabaseMock
	SetupSysParams       func() *srvsysparamsmock.ServiceMock
	ExpectedEffects      []typewheeleffects.LastWheelRow
	ExpectedError        error
	ExpectedErrorCode    string
}

var MakeEffectRollTestCases = []MakeEffectRollTestCase{
	{
		// GetAvailableWheelRowsCommand fails. The error returns.
		Name:   "GetAvailableWheelRows_Error",
		UserId: 1,
		SetupWheelEffectMock: func() *dbwheeleffectsmock.DatabaseMock {
			databaseMock := new(dbwheeleffectsmock.DatabaseMock)
			databaseMock.On("GetAvailableWheelRowsCommand", 1, 1, 1).Return([]typewheeleffects.WheelRow(nil), dbError)
			return databaseMock
		},
		SetupSysParams: func() *srvsysparamsmock.ServiceMock { return new(srvsysparamsmock.ServiceMock) },
		ExpectedError:  dbError,
	},
	{
		// GetAvailableWheelRowsCommand succeeds but the minimum effects count sys param lookup fails.
		Name:   "GetMinimumEffectsCount_SysParam_Error",
		UserId: 1,
		SetupWheelEffectMock: func() *dbwheeleffectsmock.DatabaseMock {
			databaseMock := new(dbwheeleffectsmock.DatabaseMock)
			databaseMock.On("GetAvailableWheelRowsCommand", 1, 1, 1).Return(candidateRows, nil)
			return databaseMock
		},
		SetupSysParams: func() *srvsysparamsmock.ServiceMock {
			spSvc := new(srvsysparamsmock.ServiceMock)
			spSvc.On("GetInt", 1, typesysparams.ParamMinimumAvailableWheelEffectsForRoll).Return(0, dbError)
			return spSvc
		},
		ExpectedError: dbError,
	},
	{
		// Fewer candidates are available than the configured minimum. The NOT_ENOUGH_AVAILABLE_WHEEL_EFFECTS error returns.
		Name:   "NotEnoughEffects_Error",
		UserId: 1,
		SetupWheelEffectMock: func() *dbwheeleffectsmock.DatabaseMock {
			databaseMock := new(dbwheeleffectsmock.DatabaseMock)
			databaseMock.On("GetAvailableWheelRowsCommand", 1, 1, 1).Return(candidateRows[:1], nil)
			return databaseMock
		},
		SetupSysParams: func() *srvsysparamsmock.ServiceMock {
			spSvc := new(srvsysparamsmock.ServiceMock)
			spSvc.On("GetInt", 1, typesysparams.ParamMinimumAvailableWheelEffectsForRoll).Return(3, nil)
			return spSvc
		},
		ExpectedErrorCode: "NOT_ENOUGH_AVAILABLE_WHEEL_EFFECTS",
	},
	{
		// ClearLastWheelEffectsCommand fails. The error returns.
		Name:   "ClearLastWheelEffects_Error",
		UserId: 1,
		SetupWheelEffectMock: func() *dbwheeleffectsmock.DatabaseMock {
			databaseMock := new(dbwheeleffectsmock.DatabaseMock)
			databaseMock.On("GetAvailableWheelRowsCommand", 1, 1, 1).Return(candidateRows, nil)
			databaseMock.On("ClearLastWheelEffectsCommand", 1, 1).Return(dbError)
			return databaseMock
		},
		SetupSysParams: func() *srvsysparamsmock.ServiceMock {
			spSvc := new(srvsysparamsmock.ServiceMock)
			spSvc.On("GetInt", 1, typesysparams.ParamMinimumAvailableWheelEffectsForRoll).Return(3, nil)
			return spSvc
		},
		ExpectedError: dbError,
	},
	{
		// ClearLastWheelEffectsCommand succeeds but AddLastRolledWheelEffectsCommand fails.
		Name:   "AddLastRolledWheelEffects_Error",
		UserId: 1,
		SetupWheelEffectMock: func() *dbwheeleffectsmock.DatabaseMock {
			databaseMock := new(dbwheeleffectsmock.DatabaseMock)
			databaseMock.On("GetAvailableWheelRowsCommand", 1, 1, 1).Return(candidateRows, nil)
			databaseMock.On("ClearLastWheelEffectsCommand", 1, 1).Return(nil)
			databaseMock.On("AddLastRolledWheelEffectsCommand", 1, 1, matchRolledRows(10, 11, 12)).
				Return([]typewheeleffects.CreatedLastWheelRow(nil), dbError)
			return databaseMock
		},
		SetupSysParams: func() *srvsysparamsmock.ServiceMock {
			spSvc := new(srvsysparamsmock.ServiceMock)
			spSvc.On("GetInt", 1, typesysparams.ParamMinimumAvailableWheelEffectsForRoll).Return(3, nil)
			return spSvc
		},
		ExpectedError: dbError,
	},
	{
		// AddLastRolledWheelEffectsCommand succeeds but the re-fetch of the freshly rolled rows fails.
		Name:   "GetLastRolledWheelEffectsAfterAdd_Error",
		UserId: 1,
		SetupWheelEffectMock: func() *dbwheeleffectsmock.DatabaseMock {
			databaseMock := new(dbwheeleffectsmock.DatabaseMock)
			databaseMock.On("GetAvailableWheelRowsCommand", 1, 1, 1).Return(candidateRows, nil)
			databaseMock.On("ClearLastWheelEffectsCommand", 1, 1).Return(nil)
			databaseMock.On("AddLastRolledWheelEffectsCommand", 1, 1, matchRolledRows(10, 11, 12)).
				Return([]typewheeleffects.CreatedLastWheelRow{}, nil)
			databaseMock.On("GetLastRolledWheelEffectsCommand", 1, 1).Return([]typewheeleffects.LastWheelRow(nil), dbError)
			return databaseMock
		},
		SetupSysParams: func() *srvsysparamsmock.ServiceMock {
			spSvc := new(srvsysparamsmock.ServiceMock)
			spSvc.On("GetInt", 1, typesysparams.ParamMinimumAvailableWheelEffectsForRoll).Return(3, nil)
			return spSvc
		},
		ExpectedError: dbError,
	},
	{
		// All steps succeed. The freshly rolled rows return.
		Name:   "Success",
		UserId: 1,
		SetupWheelEffectMock: func() *dbwheeleffectsmock.DatabaseMock {
			databaseMock := new(dbwheeleffectsmock.DatabaseMock)
			databaseMock.On("GetAvailableWheelRowsCommand", 1, 1, 1).Return(candidateRows, nil)
			databaseMock.On("ClearLastWheelEffectsCommand", 1, 1).Return(nil)
			databaseMock.On("AddLastRolledWheelEffectsCommand", 1, 1, matchRolledRows(10, 11, 12)).
				Return([]typewheeleffects.CreatedLastWheelRow{}, nil)
			databaseMock.On("GetLastRolledWheelEffectsCommand", 1, 1).Return(rolledResult, nil)
			return databaseMock
		},
		SetupSysParams: func() *srvsysparamsmock.ServiceMock {
			spSvc := new(srvsysparamsmock.ServiceMock)
			spSvc.On("GetInt", 1, typesysparams.ParamMinimumAvailableWheelEffectsForRoll).Return(3, nil)
			return spSvc
		},
		ExpectedEffects: rolledResult,
	},
}

func TestSrvWheelEffects_MakeEffectRoll(test *testing.T) {
	for _, testCase := range MakeEffectRollTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			weDatabaseMock := testCase.SetupWheelEffectMock()
			spSvc := testCase.SetupSysParams()

			sut := srvwheeleffects.Service{
				Database:         weDatabaseMock,
				SysParamsService: spSvc,
			}

			// Act
			effects, err := sut.MakeEffectRoll(test.Context(), testCase.UserId, 1, false)

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
				require.Equal(test, testCase.ExpectedEffects, effects)
			}

			weDatabaseMock.AssertExpectations(test)
			spSvc.AssertExpectations(test)
		})
	}
}
