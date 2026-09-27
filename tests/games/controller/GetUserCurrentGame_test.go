package ctrlgames_test

import (
	"FGG-Service/src/common"
	"FGG-Service/src/games/controller"
	"FGG-Service/src/games/types"
	"FGG-Service/tests/auth/mock"
	"FGG-Service/tests/games/mock/srvgames"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type GetUserCurrentGameTestCase struct {
	Name           string
	SetupMock      func() (*srvgamesmock.ServiceMock, *srvauthmock.ServiceMock)
	ExpectedStatus int
}

var currentGame = typegames.CurrentGame{Name: "Half-Life 1"}

var GetUserCurrentGameTestCases = []GetUserCurrentGameTestCase{
	{
		// DoesUserSessionExist returns an error. The error will return.
		Name: "DoesUserSessionExist_Error",
		SetupMock: func() (*srvgamesmock.ServiceMock, *srvauthmock.ServiceMock) {
			gameServiceMock := new(srvgamesmock.ServiceMock)
			authServiceMock := new(srvauthmock.ServiceMock)

			authServiceMock.
				On("DoesUserSessionExist", mock.Anything).
				Return(false, dbError)

			return gameServiceMock, authServiceMock
		},
		ExpectedStatus: http.StatusInternalServerError,
	},
	{
		// DoesUserSessionExist returns false. 401 will return.
		Name: "NoActiveSession",
		SetupMock: func() (*srvgamesmock.ServiceMock, *srvauthmock.ServiceMock) {
			gameServiceMock := new(srvgamesmock.ServiceMock)
			authServiceMock := new(srvauthmock.ServiceMock)

			authServiceMock.
				On("DoesUserSessionExist", mock.Anything).
				Return(false, nil)

			return gameServiceMock, authServiceMock
		},
		ExpectedStatus: http.StatusUnauthorized,
	},
	{
		// GetUserIdByLogin returns NotFoundError. 404 will return.
		Name: "LoginNotFound",
		SetupMock: func() (*srvgamesmock.ServiceMock, *srvauthmock.ServiceMock) {
			gameServiceMock := new(srvgamesmock.ServiceMock)
			authServiceMock := new(srvauthmock.ServiceMock)

			authServiceMock.
				On("DoesUserSessionExist", mock.Anything).
				Return(true, nil)
			authServiceMock.
				On("GetUserIdByLogin", login).
				Return(0, common.NewUserLoginNotFoundError(login))

			return gameServiceMock, authServiceMock
		},
		ExpectedStatus: http.StatusNotFound,
	},
	{
		// GetUserIdByLogin returns a database error. 500 will return.
		Name: "GetUserIdByLogin_DatabaseError",
		SetupMock: func() (*srvgamesmock.ServiceMock, *srvauthmock.ServiceMock) {
			gameServiceMock := new(srvgamesmock.ServiceMock)
			authServiceMock := new(srvauthmock.ServiceMock)

			authServiceMock.
				On("DoesUserSessionExist", mock.Anything).
				Return(true, nil)
			authServiceMock.
				On("GetUserIdByLogin", login).
				Return(0, dbError)

			return gameServiceMock, authServiceMock
		},
		ExpectedStatus: http.StatusInternalServerError,
	},
	{
		// GetCurrentGame returns NotFoundError. 404 will return.
		Name: "CurrentGameNotFound",
		SetupMock: func() (*srvgamesmock.ServiceMock, *srvauthmock.ServiceMock) {
			gameServiceMock := new(srvgamesmock.ServiceMock)
			authServiceMock := new(srvauthmock.ServiceMock)

			authServiceMock.
				On("DoesUserSessionExist", mock.Anything).
				Return(true, nil)
			authServiceMock.
				On("GetUserIdByLogin", login).
				Return(1, nil)
			gameServiceMock.
				On("GetCurrentGame", 1).
				Return(typegames.CurrentGame{}, common.NewCurrentGameNotFoundError())

			return gameServiceMock, authServiceMock
		},
		ExpectedStatus: http.StatusNotFound,
	},
	{
		// GetCurrentGame returns a database error. 500 will return.
		Name: "GetCurrentGame_DatabaseError",
		SetupMock: func() (*srvgamesmock.ServiceMock, *srvauthmock.ServiceMock) {
			gameServiceMock := new(srvgamesmock.ServiceMock)
			authServiceMock := new(srvauthmock.ServiceMock)

			authServiceMock.
				On("DoesUserSessionExist", mock.Anything).
				Return(true, nil)
			authServiceMock.
				On("GetUserIdByLogin", login).
				Return(1, nil)
			gameServiceMock.
				On("GetCurrentGame", 1).
				Return(typegames.CurrentGame{}, dbError)

			return gameServiceMock, authServiceMock
		},
		ExpectedStatus: http.StatusInternalServerError,
	},
	{
		// Everything succeeds. 200 will return.
		Name: "SuccessReturn",
		SetupMock: func() (*srvgamesmock.ServiceMock, *srvauthmock.ServiceMock) {
			gameServiceMock := new(srvgamesmock.ServiceMock)
			authServiceMock := new(srvauthmock.ServiceMock)

			authServiceMock.
				On("DoesUserSessionExist", mock.Anything).
				Return(true, nil)
			authServiceMock.
				On("GetUserIdByLogin", login).
				Return(1, nil)
			gameServiceMock.
				On("GetCurrentGame", 1).
				Return(currentGame, nil)

			return gameServiceMock, authServiceMock
		},
		ExpectedStatus: http.StatusOK,
	},
}

func TestCtrlGames_GetUserCurrentGame(test *testing.T) {
	for _, testCase := range GetUserCurrentGameTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)
			ctx.SetParamNames("login")
			ctx.SetParamValues(login)

			gameServiceMock, authServiceMock := testCase.SetupMock()
			sut := ctrlgames.Controller{
				Service:     gameServiceMock,
				AuthService: authServiceMock,
			}

			// Act
			sut.GetUserCurrentGame(ctx, login)

			// Assert
			require.Equal(test, testCase.ExpectedStatus, rec.Code)

			gameServiceMock.AssertExpectations(test)
			authServiceMock.AssertExpectations(test)
		})
	}
}
