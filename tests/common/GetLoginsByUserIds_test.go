package common_test

import (
	"FGG-Service/src/common"
	srvauthmock "FGG-Service/tests/auth/mock"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// A history names the same few users many times, so each user is looked up once however often the
// id repeats.
func TestGetLoginsByUserIds(test *testing.T) {
	test.Run("Success_LooksUpEachUserOnce", func(test *testing.T) {
		authMock := new(srvauthmock.ServiceMock)
		authMock.On("GetLoginByUserId", 3).Return("Jegern", nil).Once()
		authMock.On("GetLoginByUserId", 9).Return("North", nil).Once()

		logins, err := common.GetLoginsByUserIds(test.Context(), authMock, []int{3, 9, 3, 3, 9})

		require.NoError(test, err)
		require.Equal(test, map[int]string{3: "Jegern", 9: "North"}, logins)
		authMock.AssertExpectations(test)
	})

	test.Run("NoIds_NothingLookedUp", func(test *testing.T) {
		authMock := new(srvauthmock.ServiceMock)

		logins, err := common.GetLoginsByUserIds(test.Context(), authMock, nil)

		require.NoError(test, err)
		require.Empty(test, logins)
		authMock.AssertNotCalled(test, "GetLoginByUserId")
	})

	test.Run("LookupFails_ErrorReturned", func(test *testing.T) {
		lookupError := errors.New("database connection lost")
		authMock := new(srvauthmock.ServiceMock)
		authMock.On("GetLoginByUserId", 3).Return("", lookupError)

		_, err := common.GetLoginsByUserIds(test.Context(), authMock, []int{3})

		require.ErrorIs(test, err, lookupError)
	})
}
