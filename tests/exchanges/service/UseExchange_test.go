package srvexchanges_test

import (
	typechanges "FGG-Service/src/changes/types"
	srvexchanges "FGG-Service/src/exchanges/service"
	typeexchanges "FGG-Service/src/exchanges/types"
	srvpoints "FGG-Service/src/points/service"
	typepoints "FGG-Service/src/points/type"
	dbchangesmock "FGG-Service/tests/changes/mock"
	srvchangesmock "FGG-Service/tests/changes/srvmock"
	dbexchangesmock "FGG-Service/tests/exchanges/mock"
	dbpointsmock "FGG-Service/tests/points/mock"
	"testing"

	"github.com/stretchr/testify/require"
)

var seize = typeexchanges.Exchange{Id: 2, PartyId: 1, Name: "seize", Description: "takes territory"}

var boundedPoints = typepoints.PointTypeInfo{
	Id: 6, PartyId: 1, Name: "boundedPoints", Minimum: ptr(0), Maximum: ptr(100),
}

func ptr[T any](value T) *T {
	return &value
}

// The source change of an exchange is its cost and carries a negative amount, so it is applied as
// stored rather than negated.
func TestSrvExchanges_UseExchange(test *testing.T) {
	test.Run("SelfExchange_ChargesAndGrantsTheSameUser", func(test *testing.T) {
		exchangesDb := new(dbexchangesmock.DatabaseMock)
		pointsDb := new(dbpointsmock.DatabaseMock)
		changesDb := new(dbchangesmock.DatabaseMock)
		changesSvc := new(srvchangesmock.ServiceMock)

		cost := typechanges.ChangeEntry{Amount: -3, PointTypeId: ptr(boundedPoints.Id)}
		reward := typechanges.ChangeEntry{Amount: 1, ItemId: ptr(9)}

		exchangesDb.On("GetActualExchangesCommand", 1).Return([]typeexchanges.Exchange{seize}, nil)
		exchangesDb.On("GetExchangeWithEntriesCommand", 1, seize.Id).Return(typeexchanges.ExchangeWithChanges{
			Id:           seize.Id,
			SourceChange: typechanges.Change{Entries: []typechanges.ChangeEntry{cost}},
			TargetChange: typechanges.Change{Entries: []typechanges.ChangeEntry{reward}},
		}, nil)
		pointsDb.On("GetPointTypeCommand", 1, boundedPoints.Id).Return(boundedPoints, nil)
		pointsDb.On("GetPointTypeByNameCommand", 1, boundedPoints.Name).Return(boundedPoints, nil)
		pointsDb.On("GetUserPointCommand", 5, 1, boundedPoints.Id).Return(typepoints.UserPoint{Value: 10}, nil)
		exchangesDb.On("CreateExchangeHistoryCommand", 5, 1, seize.Id, 5, (*int)(nil)).
			Return(typeexchanges.ExchangeHistoryEntry{Id: 31}, nil)

		chargedEntry := typechanges.ChangeEntry{Amount: -3, PointTypeId: ptr(boundedPoints.Id), UserId: ptr(5)}
		rewardedEntry := typechanges.ChangeEntry{Amount: 1, ItemId: ptr(9), UserId: ptr(5)}

		changesDb.On("CreateUserChangeFromJsonbCommand", 1, []typechanges.ChangeEntry{chargedEntry}).
			Return(typechanges.UserChange{Entries: []typechanges.ChangeEntry{chargedEntry}}, nil)
		changesDb.On("CreateUserChangeFromJsonbCommand", 1, []typechanges.ChangeEntry{rewardedEntry}).
			Return(typechanges.UserChange{Entries: []typechanges.ChangeEntry{rewardedEntry}}, nil)
		changesSvc.On("ApplyChangeEntries", 1, []typechanges.ChangeEntry{chargedEntry}, 5, 31).Return(nil)
		changesSvc.On("ApplyChangeEntries", 1, []typechanges.ChangeEntry{rewardedEntry}, 5, 31).Return(nil)

		sut := srvexchanges.Service{
			Database:        exchangesDb,
			ChangesDatabase: changesDb,
			ChangesService:  changesSvc,
			PointsService:   &srvpoints.Service{Database: pointsDb},
		}

		err := sut.UseExchange(5, 1, seize.Name, nil)

		require.NoError(test, err)
		exchangesDb.AssertExpectations(test)
		changesDb.AssertExpectations(test)
		changesSvc.AssertExpectations(test)
	})

	test.Run("CostBeyondBalance_Rejected", func(test *testing.T) {
		exchangesDb := new(dbexchangesmock.DatabaseMock)
		pointsDb := new(dbpointsmock.DatabaseMock)

		cost := typechanges.ChangeEntry{Amount: -30, PointTypeId: ptr(boundedPoints.Id)}

		exchangesDb.On("GetActualExchangesCommand", 1).Return([]typeexchanges.Exchange{seize}, nil)
		exchangesDb.On("GetExchangeWithEntriesCommand", 1, seize.Id).Return(typeexchanges.ExchangeWithChanges{
			Id:           seize.Id,
			SourceChange: typechanges.Change{Entries: []typechanges.ChangeEntry{cost}},
		}, nil)
		pointsDb.On("GetPointTypeCommand", 1, boundedPoints.Id).Return(boundedPoints, nil)
		pointsDb.On("GetPointTypeByNameCommand", 1, boundedPoints.Name).Return(boundedPoints, nil)
		pointsDb.On("GetUserPointCommand", 5, 1, boundedPoints.Id).Return(typepoints.UserPoint{Value: 10}, nil)

		sut := srvexchanges.Service{Database: exchangesDb, PointsService: &srvpoints.Service{Database: pointsDb}}

		err := sut.UseExchange(5, 1, seize.Name, nil)

		require.Error(test, err)
		exchangesDb.AssertNotCalled(test, "CreateExchangeHistoryCommand")
	})
}
