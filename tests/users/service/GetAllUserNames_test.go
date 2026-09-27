package srvusers_test

import (
	typeparties "FGG-Service/src/parties/types"
	srvusers "FGG-Service/src/users/service"
	typeusers "FGG-Service/src/users/types"
	dbpartiesmock "FGG-Service/tests/parties/mock"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type GetAllUserNamesTestCase struct {
	Name            string
	SetupMock       func() *dbpartiesmock.DatabaseMock
	ExpectedUsers   typeusers.Users
	ExpectedErrorIs error
}

func ptr(s string) *string { return &s }

var leftDate = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

var GetAllUserNamesTestCases = []GetAllUserNamesTestCase{
	{
		// Reading the members fails. The error returns.
		Name: "DatabaseError",
		SetupMock: func() *dbpartiesmock.DatabaseMock {
			databaseMock := new(dbpartiesmock.DatabaseMock)

			databaseMock.
				On("GetMembersCommand", 1).
				Return([]typeparties.MemberWithLogin{}, dbError)

			return databaseMock
		},
		ExpectedErrorIs: dbError,
	},
	{
		// The party has no members. An empty list returns.
		Name: "EmptyList",
		SetupMock: func() *dbpartiesmock.DatabaseMock {
			databaseMock := new(dbpartiesmock.DatabaseMock)

			databaseMock.
				On("GetMembersCommand", 1).
				Return([]typeparties.MemberWithLogin{}, nil)

			return databaseMock
		},
	},
	{
		// Every current member is listed by login and display name.
		Name: "Success",
		SetupMock: func() *dbpartiesmock.DatabaseMock {
			databaseMock := new(dbpartiesmock.DatabaseMock)

			databaseMock.
				On("GetMembersCommand", 1).
				Return([]typeparties.MemberWithLogin{
					{Login: "alice", DisplayName: ptr("Alice")},
					{Login: "bob", DisplayName: ptr("Bob")},
				}, nil)

			return databaseMock
		},
		ExpectedUsers: typeusers.Users{
			{Login: "alice", DisplayName: ptr("Alice")},
			{Login: "bob", DisplayName: ptr("Bob")},
		},
	},
	{
		// A member who has left the party is not listed.
		Name: "LeftMemberSkipped",
		SetupMock: func() *dbpartiesmock.DatabaseMock {
			databaseMock := new(dbpartiesmock.DatabaseMock)

			databaseMock.
				On("GetMembersCommand", 1).
				Return([]typeparties.MemberWithLogin{
					{Login: "alice", DisplayName: ptr("Alice")},
					{Login: "gone", DisplayName: ptr("Gone"), LeftDate: &leftDate},
				}, nil)

			return databaseMock
		},
		ExpectedUsers: typeusers.Users{
			{Login: "alice", DisplayName: ptr("Alice")},
		},
	},
}

func TestSrvUsers_GetAllUserNames(test *testing.T) {
	for _, testCase := range GetAllUserNamesTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			databaseMock := testCase.SetupMock()
			sut := srvusers.Service{PartiesDatabase: databaseMock}

			// Act
			users, err := sut.GetAllUserNames()

			// Assert
			if testCase.ExpectedErrorIs != nil {
				require.ErrorIs(test, err, testCase.ExpectedErrorIs)
			} else {
				require.NoError(test, err)
				require.Equal(test, testCase.ExpectedUsers, users)
			}

			databaseMock.AssertExpectations(test)
		})
	}
}