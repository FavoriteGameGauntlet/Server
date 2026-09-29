package srvpoints_test

import (
	"FGG-Service/src/common"
	typeparties "FGG-Service/src/parties/types"
	srvpoints "FGG-Service/src/points/service"
	typepoints "FGG-Service/src/points/type"
	dbpartiesmock "FGG-Service/tests/parties/mock"
	dbpointsmock "FGG-Service/tests/points/mock"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

// The user asking is 9 and the user the points belong to is 2 throughout.
var publicPointType = typepoints.PointTypeInfo{Id: 10, PartyId: 1, Name: "openPoints", StartValue: 1, IsPublic: true}
var hiddenPointType = typepoints.PointTypeInfo{Id: 11, PartyId: 1, Name: "secretPoints", StartValue: 2}
var hiddenSharedPointType = typepoints.PointTypeInfo{Id: 12, PartyId: 1, Name: "secretFunds", StartValue: 3, IsShared: true}

// partiesMockWithMember makes user 9 a member of party 1, with or without admin rights.
func partiesMockWithMember(isAdmin bool) *dbpartiesmock.DatabaseMock {
	partiesMock := new(dbpartiesmock.DatabaseMock)
	partiesMock.On("GetMemberCommand", 9, 1).Return(typeparties.MemberWithLogin{IsAdmin: isAdmin}, nil)

	return partiesMock
}

// partiesMockWithoutMember leaves user 9 outside party 1.
func partiesMockWithoutMember() *dbpartiesmock.DatabaseMock {
	partiesMock := new(dbpartiesmock.DatabaseMock)
	partiesMock.On("GetMemberCommand", 9, 1).Return(typeparties.MemberWithLogin{}, sql.ErrNoRows)

	return partiesMock
}

// A point type that is not public is listed for an admin only; a user outside the party is no admin.
func TestSrvPoints_GetVisiblePointTypes(test *testing.T) {
	allTypes := []typepoints.PointTypeInfo{publicPointType, hiddenPointType}

	test.Run("NotAdmin_LeavesHiddenTypeOut", func(test *testing.T) {
		databaseMock := new(dbpointsmock.DatabaseMock)
		databaseMock.On("GetPointTypesCommand", 1).Return(allTypes, nil)

		sut := srvpoints.Service{Database: databaseMock, PartiesDatabase: partiesMockWithMember(false)}

		pointTypes, err := sut.GetVisiblePointTypes(9, 1)

		require.NoError(test, err)
		require.Equal(test, []typepoints.PointTypeInfo{publicPointType}, pointTypes)
	})

	test.Run("NotMember_LeavesHiddenTypeOut", func(test *testing.T) {
		databaseMock := new(dbpointsmock.DatabaseMock)
		databaseMock.On("GetPointTypesCommand", 1).Return(allTypes, nil)

		sut := srvpoints.Service{Database: databaseMock, PartiesDatabase: partiesMockWithoutMember()}

		pointTypes, err := sut.GetVisiblePointTypes(9, 1)

		require.NoError(test, err)
		require.Equal(test, []typepoints.PointTypeInfo{publicPointType}, pointTypes)
	})

	test.Run("Admin_ListsEveryType", func(test *testing.T) {
		databaseMock := new(dbpointsmock.DatabaseMock)
		databaseMock.On("GetPointTypesCommand", 1).Return(allTypes, nil)

		sut := srvpoints.Service{Database: databaseMock, PartiesDatabase: partiesMockWithMember(true)}

		pointTypes, err := sut.GetVisiblePointTypes(9, 1)

		require.NoError(test, err)
		require.Equal(test, allTypes, pointTypes)
	})
}

// A hidden point type is left out of the points a non-admin reads, and its value is not even read.
func TestSrvPoints_PointLists_NotAdmin_LeaveHiddenTypesOut(test *testing.T) {
	test.Run("GetUserPoints", func(test *testing.T) {
		databaseMock := new(dbpointsmock.DatabaseMock)
		databaseMock.On("GetPointTypesCommand", 1).Return([]typepoints.PointTypeInfo{publicPointType, hiddenPointType}, nil)
		databaseMock.On("GetUserPointCommand", 2, 1, publicPointType.Id).Return(typepoints.UserPoint{}, sql.ErrNoRows)

		sut := srvpoints.Service{Database: databaseMock, PartiesDatabase: partiesMockWithMember(false)}

		values, err := sut.GetUserPoints(9, 2, 1)

		require.NoError(test, err)
		require.Equal(test, []typepoints.PointValue{{PointType: publicPointType, Value: publicPointType.StartValue}}, values)
		databaseMock.AssertNotCalled(test, "GetUserPointCommand", 2, 1, hiddenPointType.Id)
	})

	test.Run("GetAllUserPoints", func(test *testing.T) {
		databaseMock := new(dbpointsmock.DatabaseMock)
		partiesMock := partiesMockWithMember(false)
		partiesMock.On("GetMembersCommand", 1).Return([]typeparties.MemberWithLogin{{UserId: 2, Login: "alice"}}, nil)
		databaseMock.On("GetPointTypesCommand", 1).Return([]typepoints.PointTypeInfo{publicPointType, hiddenPointType}, nil)
		databaseMock.On("GetUserPointCommand", 2, 1, publicPointType.Id).Return(typepoints.UserPoint{Value: 5}, nil)

		sut := srvpoints.Service{Database: databaseMock, PartiesDatabase: partiesMock}

		byLogin, err := sut.GetAllUserPoints(9, 1)

		require.NoError(test, err)
		require.Equal(test, []typepoints.UserPointValuesByLogin{
			{Login: "alice", Points: []typepoints.PointValue{{PointType: publicPointType, Value: 5}}},
		}, byLogin)
		databaseMock.AssertNotCalled(test, "GetUserPointCommand", 2, 1, hiddenPointType.Id)
	})

	test.Run("GetAllPartyPoints", func(test *testing.T) {
		publicSharedType := typepoints.PointTypeInfo{Id: 13, PartyId: 1, Name: "openFunds", StartValue: 4, IsShared: true, IsPublic: true}
		databaseMock := new(dbpointsmock.DatabaseMock)
		databaseMock.On("GetPointTypesCommand", 1).Return([]typepoints.PointTypeInfo{publicSharedType, hiddenSharedPointType}, nil)
		databaseMock.On("GetPartyPointCommand", 1, publicSharedType.Id).Return(typepoints.PartyPoint{Value: 9}, nil)

		sut := srvpoints.Service{Database: databaseMock, PartiesDatabase: partiesMockWithMember(false)}

		values, err := sut.GetAllPartyPoints(9, 1)

		require.NoError(test, err)
		require.Equal(test, []typepoints.PointValue{{PointType: publicSharedType, Value: 9}}, values)
		databaseMock.AssertNotCalled(test, "GetPartyPointCommand", 1, hiddenSharedPointType.Id)
	})
}

// A hidden point type read by name is answered as not found to a non-admin, before anything about it
// is read, including the conflict a shared type would otherwise report.
func TestSrvPoints_ReadByTypeName_NotAdmin_HiddenType_NotFound(test *testing.T) {
	read := map[string]func(sut *srvpoints.Service, name string) error{
		"GetUserPointValueByTypeName": func(sut *srvpoints.Service, name string) error {
			_, err := sut.GetUserPointValueByTypeName(9, 2, 1, name)
			return err
		},
		"GetUserPointHistoryByTypeName": func(sut *srvpoints.Service, name string) error {
			_, err := sut.GetUserPointHistoryByTypeName(9, 2, 1, name)
			return err
		},
		"GetPartyPointValueByTypeName": func(sut *srvpoints.Service, name string) error {
			_, err := sut.GetPartyPointValueByTypeName(9, 1, name)
			return err
		},
		"GetPartyPointHistoryByTypeName": func(sut *srvpoints.Service, name string) error {
			_, err := sut.GetPartyPointHistoryByTypeName(9, 1, name)
			return err
		},
	}

	for readName, readHidden := range read {
		test.Run(readName, func(test *testing.T) {
			databaseMock := new(dbpointsmock.DatabaseMock)
			databaseMock.On("GetPointTypeByNameCommand", 1, hiddenSharedPointType.Name).Return(hiddenSharedPointType, nil)

			sut := srvpoints.Service{Database: databaseMock, PartiesDatabase: partiesMockWithMember(false)}

			err := readHidden(&sut, hiddenSharedPointType.Name)

			var notFound *common.NotFoundError
			require.ErrorAs(test, err, &notFound)
			databaseMock.AssertNotCalled(test, "GetUserPointCommand")
			databaseMock.AssertNotCalled(test, "GetPartyPointCommand")
			databaseMock.AssertNotCalled(test, "GetUserPointHistoryCommand")
			databaseMock.AssertNotCalled(test, "GetPartyPointHistoryCommand")
		})
	}
}

// An admin reads a hidden point type like any other.
func TestSrvPoints_GetUserPointValueByTypeName_Admin_HiddenType_ReadsValue(test *testing.T) {
	// Arrange
	databaseMock := new(dbpointsmock.DatabaseMock)
	databaseMock.On("GetPointTypeByNameCommand", 1, hiddenPointType.Name).Return(hiddenPointType, nil)
	databaseMock.On("GetUserPointCommand", 2, 1, hiddenPointType.Id).Return(typepoints.UserPoint{Value: 6}, nil)

	sut := srvpoints.Service{Database: databaseMock, PartiesDatabase: partiesMockWithMember(true)}

	// Act
	value, err := sut.GetUserPointValueByTypeName(9, 2, 1, hiddenPointType.Name)

	// Assert
	require.NoError(test, err)
	require.Equal(test, 6, value)
}

// Hiding a point type from readers does not stop its points from being changed.
func TestSrvPoints_ChangePointValueByTypeName_HiddenType_Changes(test *testing.T) {
	// Arrange
	databaseMock := new(dbpointsmock.DatabaseMock)
	databaseMock.On("GetPointTypeByNameCommand", 1, hiddenPointType.Name).Return(hiddenPointType, nil)
	databaseMock.On("GetUserPointCommand", 2, 1, hiddenPointType.Id).Return(typepoints.UserPoint{Value: 3}, nil)
	databaseMock.On("ChangeUserPointValueCommand", 2, 1, hiddenPointType.Id, 4).Return(nil)
	databaseMock.On("CreateUserPointHistoryCommand", 2, 1, hiddenPointType.Id, 2, 4, 4, 7, 50).
		Return(typepoints.UserPointHistoryEntry{}, nil)

	sut := srvpoints.Service{Database: databaseMock}

	// Act
	err := sut.ChangePointValueByTypeName(2, 1, hiddenPointType.Name, 4, 50)

	// Assert
	require.NoError(test, err)
	databaseMock.AssertExpectations(test)
}
