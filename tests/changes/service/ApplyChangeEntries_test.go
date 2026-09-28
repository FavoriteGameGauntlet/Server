package srvchanges_test

import (
	srvchanges "FGG-Service/src/changes/service"
	"FGG-Service/src/changes/types"
	"FGG-Service/src/effects/types"
	"FGG-Service/src/items/types"
	"FGG-Service/src/perks/types"
	srvpoints "FGG-Service/src/points/service"
	typepoints "FGG-Service/src/points/type"
	dbeffectsmock "FGG-Service/tests/effects/mock"
	dbitemsmock "FGG-Service/tests/items/mock"
	dbperksmock "FGG-Service/tests/perks/mock"
	dbpointsmock "FGG-Service/tests/points/mock"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

var dbError = errors.New("database connection lost")

func ptr[T any](v T) *T { return &v }

type ApplyChangeEntriesTestCase struct {
	Name           string
	Entries        []typechanges.ChangeEntry
	SetupMocks     func() (*dbitemsmock.DatabaseMock, *dbperksmock.DatabaseMock, *dbeffectsmock.DatabaseMock, *dbpointsmock.DatabaseMock)
	ExpectedError  error
	ExpectAnyError bool
}

var ApplyChangeEntriesTestCases = []ApplyChangeEntriesTestCase{
	{
		// An entry with no UserId can't be applied. An error returns.
		Name: "NoUserId_Error",
		Entries: []typechanges.ChangeEntry{
			{Amount: 5, PointTypeId: ptr(7)},
		},
		SetupMocks: func() (*dbitemsmock.DatabaseMock, *dbperksmock.DatabaseMock, *dbeffectsmock.DatabaseMock, *dbpointsmock.DatabaseMock) {
			return new(dbitemsmock.DatabaseMock), new(dbperksmock.DatabaseMock), new(dbeffectsmock.DatabaseMock), new(dbpointsmock.DatabaseMock)
		},
		ExpectAnyError: true,
	},
	{
		// An entry with none of PointTypeId/ItemId/PerkId/EffectId set can't be applied. An error returns.
		Name: "NoTarget_Error",
		Entries: []typechanges.ChangeEntry{
			{Amount: 5, UserId: ptr(2)},
		},
		SetupMocks: func() (*dbitemsmock.DatabaseMock, *dbperksmock.DatabaseMock, *dbeffectsmock.DatabaseMock, *dbpointsmock.DatabaseMock) {
			return new(dbitemsmock.DatabaseMock), new(dbperksmock.DatabaseMock), new(dbeffectsmock.DatabaseMock), new(dbpointsmock.DatabaseMock)
		},
		ExpectAnyError: true,
	},
	{
		// A PointTypeId entry is applied via the points service (clamped write + history).
		Name: "PointTypeEntry_Success",
		Entries: []typechanges.ChangeEntry{
			{Amount: 5, PointTypeId: ptr(7), UserId: ptr(2)},
		},
		SetupMocks: func() (*dbitemsmock.DatabaseMock, *dbperksmock.DatabaseMock, *dbeffectsmock.DatabaseMock, *dbpointsmock.DatabaseMock) {
			pointsDb := new(dbpointsmock.DatabaseMock)
			pointsDb.On("GetPointTypeCommand", 1, 7).Return(typepoints.PointTypeInfo{Id: 7, PartyId: 1, Minimum: ptr(0), Maximum: ptr(100)}, nil)
			pointsDb.On("GetUserPointCommand", 2, 1, 7).Return(typepoints.UserPoint{Value: 10}, nil)
			pointsDb.On("ChangeUserPointValueCommand", 2, 1, 7, 5).Return(nil)
			pointsDb.On("CreateUserPointHistoryCommand", 2, 1, 7, 9, 5, 5, 15, 999).Return(typepoints.UserPointHistoryEntry{}, nil)
			return new(dbitemsmock.DatabaseMock), new(dbperksmock.DatabaseMock), new(dbeffectsmock.DatabaseMock), pointsDb
		},
	},
	{
		// The points service fails to apply the entry. The error returns.
		Name: "PointTypeEntry_Error",
		Entries: []typechanges.ChangeEntry{
			{Amount: 5, PointTypeId: ptr(7), UserId: ptr(2)},
		},
		SetupMocks: func() (*dbitemsmock.DatabaseMock, *dbperksmock.DatabaseMock, *dbeffectsmock.DatabaseMock, *dbpointsmock.DatabaseMock) {
			pointsDb := new(dbpointsmock.DatabaseMock)
			pointsDb.On("GetPointTypeCommand", 1, 7).Return(typepoints.PointTypeInfo{}, dbError)
			return new(dbitemsmock.DatabaseMock), new(dbperksmock.DatabaseMock), new(dbeffectsmock.DatabaseMock), pointsDb
		},
		ExpectedError: dbError,
	},
	{
		// An ItemId entry grants the item to the user.
		Name: "ItemEntry_Success",
		Entries: []typechanges.ChangeEntry{
			{Amount: 1, ItemId: ptr(30), UserId: ptr(2)},
		},
		SetupMocks: func() (*dbitemsmock.DatabaseMock, *dbperksmock.DatabaseMock, *dbeffectsmock.DatabaseMock, *dbpointsmock.DatabaseMock) {
			itemsDb := new(dbitemsmock.DatabaseMock)
			itemsDb.On("CreateUserItemCommand", 2, 1, 30, 9, 999).Return(typeitems.UserItem{Id: 1, UserId: 2, PartyId: 1, ItemId: 30}, nil)
			return itemsDb, new(dbperksmock.DatabaseMock), new(dbeffectsmock.DatabaseMock), new(dbpointsmock.DatabaseMock)
		},
	},
	{
		// Granting the item fails. The error returns.
		Name: "ItemEntry_Error",
		Entries: []typechanges.ChangeEntry{
			{Amount: 1, ItemId: ptr(30), UserId: ptr(2)},
		},
		SetupMocks: func() (*dbitemsmock.DatabaseMock, *dbperksmock.DatabaseMock, *dbeffectsmock.DatabaseMock, *dbpointsmock.DatabaseMock) {
			itemsDb := new(dbitemsmock.DatabaseMock)
			itemsDb.On("CreateUserItemCommand", 2, 1, 30, 9, 999).Return(typeitems.UserItem{}, dbError)
			return itemsDb, new(dbperksmock.DatabaseMock), new(dbeffectsmock.DatabaseMock), new(dbpointsmock.DatabaseMock)
		},
		ExpectedError: dbError,
	},
	{
		// An EffectId entry grants the effect to the user.
		Name: "EffectEntry_Success",
		Entries: []typechanges.ChangeEntry{
			{Amount: 1, EffectId: ptr(40), UserId: ptr(2)},
		},
		SetupMocks: func() (*dbitemsmock.DatabaseMock, *dbperksmock.DatabaseMock, *dbeffectsmock.DatabaseMock, *dbpointsmock.DatabaseMock) {
			effectsDb := new(dbeffectsmock.DatabaseMock)
			effectsDb.On("CreateUserEffectCommand", 2, 1, 40, 9, 999).Return(typeeffects.UserEffect{Id: 1, UserId: 2, PartyId: 1, EffectId: 40}, nil)
			return new(dbitemsmock.DatabaseMock), new(dbperksmock.DatabaseMock), effectsDb, new(dbpointsmock.DatabaseMock)
		},
	},
	{
		// Granting the effect fails. The error returns.
		Name: "EffectEntry_Error",
		Entries: []typechanges.ChangeEntry{
			{Amount: 1, EffectId: ptr(40), UserId: ptr(2)},
		},
		SetupMocks: func() (*dbitemsmock.DatabaseMock, *dbperksmock.DatabaseMock, *dbeffectsmock.DatabaseMock, *dbpointsmock.DatabaseMock) {
			effectsDb := new(dbeffectsmock.DatabaseMock)
			effectsDb.On("CreateUserEffectCommand", 2, 1, 40, 9, 999).Return(typeeffects.UserEffect{}, dbError)
			return new(dbitemsmock.DatabaseMock), new(dbperksmock.DatabaseMock), effectsDb, new(dbpointsmock.DatabaseMock)
		},
		ExpectedError: dbError,
	},
	{
		// A PerkId entry first grants the perk's underlying Effect, then grants the perk itself using
		// the same source event.
		Name: "PerkEntry_Success_GrantsUnderlyingEffectFirst",
		Entries: []typechanges.ChangeEntry{
			{Amount: 1, PerkId: ptr(50), UserId: ptr(2)},
		},
		SetupMocks: func() (*dbitemsmock.DatabaseMock, *dbperksmock.DatabaseMock, *dbeffectsmock.DatabaseMock, *dbpointsmock.DatabaseMock) {
			perksDb := new(dbperksmock.DatabaseMock)
			effectsDb := new(dbeffectsmock.DatabaseMock)

			perksDb.On("GetPerkCommand", 1, 50).Return(typeperks.Perk{Id: 50, PartyId: 1, EffectId: 40}, nil)
			effectsDb.On("CreateUserEffectCommand", 2, 1, 40, 9, 999).
				Return(typeeffects.UserEffect{Id: 1, UserId: 2, PartyId: 1, EffectId: 40, EffectHistoryId: 777}, nil)
			perksDb.On("CreateUserPerkCommand", 2, 1, 50, 9, 999).Return(typeperks.UserPerk{Id: 1, UserId: 2, PartyId: 1, PerkId: 50, UserEffectId: 777}, nil)

			return new(dbitemsmock.DatabaseMock), perksDb, effectsDb, new(dbpointsmock.DatabaseMock)
		},
	},
	{
		// Looking up the perk fails. The error returns and the effect is never granted.
		Name: "PerkEntry_GetPerk_Error",
		Entries: []typechanges.ChangeEntry{
			{Amount: 1, PerkId: ptr(50), UserId: ptr(2)},
		},
		SetupMocks: func() (*dbitemsmock.DatabaseMock, *dbperksmock.DatabaseMock, *dbeffectsmock.DatabaseMock, *dbpointsmock.DatabaseMock) {
			perksDb := new(dbperksmock.DatabaseMock)
			perksDb.On("GetPerkCommand", 1, 50).Return(typeperks.Perk{}, dbError)

			return new(dbitemsmock.DatabaseMock), perksDb, new(dbeffectsmock.DatabaseMock), new(dbpointsmock.DatabaseMock)
		},
		ExpectedError: dbError,
	},
	{
		// Granting the perk's underlying effect fails. The error returns and the perk is never granted.
		Name: "PerkEntry_GrantEffect_Error",
		Entries: []typechanges.ChangeEntry{
			{Amount: 1, PerkId: ptr(50), UserId: ptr(2)},
		},
		SetupMocks: func() (*dbitemsmock.DatabaseMock, *dbperksmock.DatabaseMock, *dbeffectsmock.DatabaseMock, *dbpointsmock.DatabaseMock) {
			perksDb := new(dbperksmock.DatabaseMock)
			effectsDb := new(dbeffectsmock.DatabaseMock)

			perksDb.On("GetPerkCommand", 1, 50).Return(typeperks.Perk{Id: 50, PartyId: 1, EffectId: 40}, nil)
			effectsDb.On("CreateUserEffectCommand", 2, 1, 40, 9, 999).Return(typeeffects.UserEffect{}, dbError)

			return new(dbitemsmock.DatabaseMock), perksDb, effectsDb, new(dbpointsmock.DatabaseMock)
		},
		ExpectedError: dbError,
	},
	{
		// The effect is granted but granting the perk fails. The error returns.
		Name: "PerkEntry_CreateUserPerk_Error",
		Entries: []typechanges.ChangeEntry{
			{Amount: 1, PerkId: ptr(50), UserId: ptr(2)},
		},
		SetupMocks: func() (*dbitemsmock.DatabaseMock, *dbperksmock.DatabaseMock, *dbeffectsmock.DatabaseMock, *dbpointsmock.DatabaseMock) {
			perksDb := new(dbperksmock.DatabaseMock)
			effectsDb := new(dbeffectsmock.DatabaseMock)

			perksDb.On("GetPerkCommand", 1, 50).Return(typeperks.Perk{Id: 50, PartyId: 1, EffectId: 40}, nil)
			effectsDb.On("CreateUserEffectCommand", 2, 1, 40, 9, 999).
				Return(typeeffects.UserEffect{EffectHistoryId: 777}, nil)
			perksDb.On("CreateUserPerkCommand", 2, 1, 50, 9, 999).Return(typeperks.UserPerk{}, dbError)

			return new(dbitemsmock.DatabaseMock), perksDb, effectsDb, new(dbpointsmock.DatabaseMock)
		},
		ExpectedError: dbError,
	},
	{
		// Multiple entries apply in order; processing stops at the first failure without applying
		// entries after it.
		Name: "MultipleEntries_StopsAtFirstError",
		Entries: []typechanges.ChangeEntry{
			{Amount: 1, ItemId: ptr(30), UserId: ptr(2)},
			{Amount: 1, EffectId: ptr(40), UserId: ptr(2)},
		},
		SetupMocks: func() (*dbitemsmock.DatabaseMock, *dbperksmock.DatabaseMock, *dbeffectsmock.DatabaseMock, *dbpointsmock.DatabaseMock) {
			itemsDb := new(dbitemsmock.DatabaseMock)
			effectsDb := new(dbeffectsmock.DatabaseMock)

			itemsDb.On("CreateUserItemCommand", 2, 1, 30, 9, 999).Return(typeitems.UserItem{}, dbError)
			// EffectId entry is never reached - no mock expectation set for it, AssertExpectations
			// only checks the effectsDb mock has no unfulfilled expectations, which holds trivially
			// since none were set; the important assertion is effectsDb.AssertNotCalled below.

			return itemsDb, new(dbperksmock.DatabaseMock), effectsDb, new(dbpointsmock.DatabaseMock)
		},
		ExpectedError: dbError,
	},
}

func TestSrvChanges_ApplyChangeEntries(test *testing.T) {
	for _, testCase := range ApplyChangeEntriesTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			itemsDb, perksDb, effectsDb, pointsDb := testCase.SetupMocks()

			sut := srvchanges.Service{
				ItemsDatabase:   itemsDb,
				PerksDatabase:   perksDb,
				EffectsDatabase: effectsDb,
				PointsService:   &srvpoints.Service{Database: pointsDb},
			}

			// Act
			err := sut.ApplyChangeEntries(1, testCase.Entries, 9, 999)

			// Assert
			if testCase.ExpectedError != nil {
				require.ErrorIs(test, err, testCase.ExpectedError)
			} else if testCase.ExpectAnyError {
				require.Error(test, err)
			} else {
				require.NoError(test, err)
			}

			itemsDb.AssertExpectations(test)
			perksDb.AssertExpectations(test)
			effectsDb.AssertExpectations(test)
			pointsDb.AssertExpectations(test)

			if testCase.Name == "MultipleEntries_StopsAtFirstError" {
				effectsDb.AssertNotCalled(test, "CreateUserEffectCommand")
			}
		})
	}
}
