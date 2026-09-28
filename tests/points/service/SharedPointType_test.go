package srvpoints_test

import (
	srvpoints "FGG-Service/src/points/service"
	typepoints "FGG-Service/src/points/type"
	typehistory "FGG-Service/src/history/types"
	dbhistorymock "FGG-Service/tests/history/mock"
	dbpointsmock "FGG-Service/tests/points/mock"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/require"
)

// A shared point type is held once for the whole party rather than per user.
var sharedPointType = typepoints.PointTypeInfo{
	Id: 7, PartyId: 1, Name: "partyFunds", StartValue: 4, IsShared: true, Minimum: ptrInt(0), Maximum: ptrInt(100),
}

func TestSrvPoints_SharedPointTypeUsesPartyPool(test *testing.T) {
	test.Run("Read_ReadsPartyPoolNotUserValue", func(test *testing.T) {
		databaseMock := new(dbpointsmock.DatabaseMock)
		databaseMock.On("GetPointTypeByNameCommand", 1, sharedPointType.Name).Return(sharedPointType, nil)
		databaseMock.On("GetPartyPointCommand", 1, sharedPointType.Id).
			Return(typepoints.PartyPoint{PartyId: 1, PointTypeId: sharedPointType.Id, Value: 42}, nil)

		sut := srvpoints.Service{Database: databaseMock}

		value, err := sut.GetPointValueByTypeName(2, 1, sharedPointType.Name)

		require.NoError(test, err)
		require.Equal(test, 42, value)
		databaseMock.AssertExpectations(test)
		databaseMock.AssertNotCalled(test, "GetUserPointCommand")
	})

	test.Run("Change_WritesPartyPoolAndPartyHistory", func(test *testing.T) {
		databaseMock := new(dbpointsmock.DatabaseMock)
		databaseMock.On("GetPointTypeCommand", 1, sharedPointType.Id).Return(sharedPointType, nil)
		databaseMock.On("GetPartyPointCommand", 1, sharedPointType.Id).
			Return(typepoints.PartyPoint{Value: 10}, nil)
		databaseMock.On("ChangePartyPointValueCommand", 1, sharedPointType.Id, 5).Return(nil)
		databaseMock.On("CreatePartyPointHistoryCommand", 1, sharedPointType.Id, 9, 5, 5, 15, 99).
			Return(typepoints.PartyPointHistoryEntry{}, nil)

		sut := srvpoints.Service{Database: databaseMock}

		err := sut.ChangeUserPointValueClamped(2, 1, sharedPointType.Id, 5, 9, 99)

		require.NoError(test, err)
		databaseMock.AssertExpectations(test)
		databaseMock.AssertNotCalled(test, "ChangeUserPointValueCommand")
		databaseMock.AssertNotCalled(test, "CreateUserPointHistoryCommand")
	})

	test.Run("MissingPartyPoolRow_ReadsStartValue", func(test *testing.T) {
		databaseMock := new(dbpointsmock.DatabaseMock)
		databaseMock.On("GetPointTypeByNameCommand", 1, sharedPointType.Name).Return(sharedPointType, nil)
		databaseMock.On("GetPartyPointCommand", 1, sharedPointType.Id).
			Return(typepoints.PartyPoint{}, sql.ErrNoRows)

		sut := srvpoints.Service{Database: databaseMock}

		value, err := sut.GetPointValueByTypeName(2, 1, sharedPointType.Name)

		require.NoError(test, err)
		require.Equal(test, sharedPointType.StartValue, value)
		databaseMock.AssertExpectations(test)
	})
}
// An administrator's direct change is recorded as a manual history entry, whose id becomes the
// source event of the resulting point history.
func TestSrvPoints_ChangeUserPointByTypeName(test *testing.T) {
	test.Run("Success_RecordsManualHistoryAsSourceEvent", func(test *testing.T) {
		databaseMock := new(dbpointsmock.DatabaseMock)
		historyMock := new(dbhistorymock.DatabaseMock)

		databaseMock.On("GetPointTypeByNameCommand", 1, availableRollsType.Name).Return(availableRollsType, nil)
		historyMock.On("CreateManualHistoryCommand", 1, 9, mock.Anything, (*int)(nil)).
			Return([]typehistory.ManualHistoryEntry{{Id: 77, UserId: 2, PartyId: 1}}, nil)
		databaseMock.On("GetUserPointCommand", 2, 1, availableRollsType.Id).
			Return(typepoints.UserPoint{Value: 3}, nil)
		databaseMock.On("ChangeUserPointValueCommand", 2, 1, availableRollsType.Id, 2).Return(nil)
		databaseMock.On("CreateUserPointHistoryCommand", 2, 1, availableRollsType.Id, 9, 2, 2, 5, 77).
			Return(typepoints.UserPointHistoryEntry{}, nil)

		sut := srvpoints.Service{Database: databaseMock, HistoryDatabase: historyMock}

		result, err := sut.ChangeUserPointByTypeName(9, 2, 1, availableRollsType.Name, 2)

		require.NoError(test, err)
		require.Equal(test, 2, result.ActualChangeValue)
		require.Equal(test, 5, result.FinalValue)
		databaseMock.AssertExpectations(test)
		historyMock.AssertExpectations(test)
	})

	test.Run("SharedPointType_Rejected", func(test *testing.T) {
		databaseMock := new(dbpointsmock.DatabaseMock)
		databaseMock.On("GetPointTypeByNameCommand", 1, sharedPointType.Name).Return(sharedPointType, nil)

		sut := srvpoints.Service{Database: databaseMock}

		_, err := sut.ChangeUserPointByTypeName(9, 2, 1, sharedPointType.Name, 2)

		require.Error(test, err)
		databaseMock.AssertNotCalled(test, "ChangeUserPointValueCommand")
	})
}

// SeedUserPoints gives a joining member a starting value for each per-user point type, and leaves
// shared types alone because those are held by the party.
func TestSrvPoints_SeedUserPoints(test *testing.T) {
	test.Run("Success_SeedsPerUserTypesOnly", func(test *testing.T) {
		databaseMock := new(dbpointsmock.DatabaseMock)
		startingType := typepoints.PointTypeInfo{Id: 3, PartyId: 1, Name: "freePoints", StartValue: 8}

		databaseMock.On("GetPointTypesCommand", 1).
			Return([]typepoints.PointTypeInfo{startingType, sharedPointType}, nil)
		databaseMock.On("CreateUserPointCommand", 2, 1, startingType.Id, 8).
			Return(typepoints.UserPoint{}, nil)

		sut := srvpoints.Service{Database: databaseMock}

		err := sut.SeedUserPoints(2, 1)

		require.NoError(test, err)
		databaseMock.AssertExpectations(test)
		databaseMock.AssertNotCalled(test, "CreateUserPointCommand", 2, 1, sharedPointType.Id, sharedPointType.StartValue)
	})
}