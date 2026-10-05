package srvpoints_test

import (
	typeparties "FGG-Service/src/parties/types"
	srvpoints "FGG-Service/src/points/service"
	typepoints "FGG-Service/src/points/type"
	dbpartiesmock "FGG-Service/tests/parties/mock"
	dbpointsmock "FGG-Service/tests/points/mock"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

var personalPointType = typepoints.PointTypeInfo{Id: 3, PartyId: 1, Name: "personalPoints", StartValue: 8}

// A user's points leave shared point types out, since those are held by the party.
func TestSrvPoints_GetUserPoints_LeavesSharedTypesOut(test *testing.T) {
	// Arrange
	databaseMock := new(dbpointsmock.DatabaseMock)
	partiesMock := new(dbpartiesmock.DatabaseMock)

	partiesMock.On("GetMemberCommand", 9, 1).Return(typeparties.MemberWithLogin{IsAdmin: true}, nil)
	databaseMock.On("GetPointTypesCommand", 1).Return([]typepoints.PointTypeInfo{personalPointType, sharedPointType}, nil)
	databaseMock.On("GetUserPointCommand", 2, 1, personalPointType.Id).Return(typepoints.UserPoint{}, sql.ErrNoRows)

	sut := srvpoints.Service{Database: databaseMock, PartiesDatabase: partiesMock}

	// Act
	values, err := sut.GetUserPoints(9, 2, 1)

	// Assert
	require.NoError(test, err)
	require.Equal(test, []typepoints.PointValue{
		{PointType: personalPointType, Value: personalPointType.StartValue},
	}, values)
	databaseMock.AssertExpectations(test)
	databaseMock.AssertNotCalled(test, "GetPartyPointCommand")
}

// Member points leave shared point types out, since those are held by the party.
func TestSrvPoints_GetAllUserPoints_LeavesSharedTypesOut(test *testing.T) {
	// Arrange
	databaseMock := new(dbpointsmock.DatabaseMock)
	partiesMock := new(dbpartiesmock.DatabaseMock)

	partiesMock.On("GetMembersCommand", 1).Return([]typeparties.MemberWithLogin{{UserId: 2, Login: "alice"}}, nil)
	partiesMock.On("GetMemberCommand", 9, 1).Return(typeparties.MemberWithLogin{IsAdmin: true}, nil)
	databaseMock.On("GetPointTypesCommand", 1).Return([]typepoints.PointTypeInfo{personalPointType, sharedPointType}, nil)
	databaseMock.On("GetUserPointCommand", 2, 1, personalPointType.Id).Return(typepoints.UserPoint{Value: 5}, nil)

	sut := srvpoints.Service{Database: databaseMock, PartiesDatabase: partiesMock}

	// Act
	byLogin, err := sut.GetAllUserPoints(9, 1)

	// Assert
	require.NoError(test, err)
	require.Equal(test, []typepoints.UserPointValuesByLogin{
		{Login: "alice", Points: []typepoints.PointValue{{PointType: personalPointType, Value: 5}}},
	}, byLogin)
	databaseMock.AssertExpectations(test)
	databaseMock.AssertNotCalled(test, "GetPartyPointCommand")
}

// Party points list only shared point types, reading the start value where the party holds none yet.
func TestSrvPoints_GetAllPartyPoints_ListsSharedTypesOnly(test *testing.T) {
	// Arrange
	databaseMock := new(dbpointsmock.DatabaseMock)
	partiesMock := new(dbpartiesmock.DatabaseMock)
	otherSharedType := typepoints.PointTypeInfo{Id: 8, PartyId: 1, Name: "partyBonus", StartValue: 2, IsShared: true}

	partiesMock.On("GetMemberCommand", 9, 1).Return(typeparties.MemberWithLogin{IsAdmin: true}, nil)
	databaseMock.On("GetPointTypesCommand", 1).
		Return([]typepoints.PointTypeInfo{personalPointType, sharedPointType, otherSharedType}, nil)
	databaseMock.On("GetPartyPointCommand", 1, sharedPointType.Id).Return(typepoints.PartyPoint{Value: 42}, nil)
	databaseMock.On("GetPartyPointCommand", 1, otherSharedType.Id).Return(typepoints.PartyPoint{}, sql.ErrNoRows)

	sut := srvpoints.Service{Database: databaseMock, PartiesDatabase: partiesMock}

	// Act
	values, err := sut.GetAllPartyPoints(9, 1)

	// Assert
	require.NoError(test, err)
	require.Equal(test, []typepoints.PointValue{
		{PointType: sharedPointType, Value: 42},
		{PointType: otherSharedType, Value: otherSharedType.StartValue},
	}, values)
	databaseMock.AssertExpectations(test)
	databaseMock.AssertNotCalled(test, "GetUserPointCommand")
}
