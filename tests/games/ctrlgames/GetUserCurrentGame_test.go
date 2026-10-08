package ctrlgames_test

import (
	"FGG-Service/src/common"
	"FGG-Service/src/games/ctrlgames"
	"FGG-Service/src/games/typegames"
	"FGG-Service/tests/auth/srvauthmock"
	"FGG-Service/tests/games/srvgamesmock"
	"FGG-Service/tests/parties/srvpartiesmock"
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
		// GetUserId returns an error. The error will return.
		Name: "GetUserId_Error",
		SetupMock: func() (*srvgamesmock.ServiceMock, *srvauthmock.ServiceMock) {
			gameServiceMock := new(srvgamesmock.ServiceMock)
			authServiceMock := new(srvauthmock.ServiceMock)

			authServiceMock.
				On("GetUserId", mock.Anything).
				Return(0, dbError)

			return gameServiceMock, authServiceMock
		},
		ExpectedStatus: http.StatusInternalServerError,
	},
	{
		// GetUserId returns UnauthorizedError. 401 will return.
		Name: "NoActiveSession",
		SetupMock: func() (*srvgamesmock.ServiceMock, *srvauthmock.ServiceMock) {
			gameServiceMock := new(srvgamesmock.ServiceMock)
			authServiceMock := new(srvauthmock.ServiceMock)

			authServiceMock.
				On("GetUserId", mock.Anything).
				Return(0, common.NewActiveSessionNotFoundUnauthorizedError())

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
				On("GetUserId", mock.Anything).
				Return(1, nil)
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
				On("GetUserId", mock.Anything).
				Return(1, nil)
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
				On("GetUserId", mock.Anything).
				Return(1, nil)
			authServiceMock.
				On("GetUserIdByLogin", login).
				Return(1, nil)
			gameServiceMock.
				On("GetCurrentGame", 1, 1).
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
				On("GetUserId", mock.Anything).
				Return(1, nil)
			authServiceMock.
				On("GetUserIdByLogin", login).
				Return(1, nil)
			gameServiceMock.
				On("GetCurrentGame", 1, 1).
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
				On("GetUserId", mock.Anything).
				Return(1, nil)
			authServiceMock.
				On("GetUserIdByLogin", login).
				Return(1, nil)
			gameServiceMock.
				On("GetCurrentGame", 1, 1).
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
			partyServiceMock := new(srvpartiesmock.ServiceMock)
			partyServiceMock.On("RequireMember", mock.Anything, 1).Return(nil).Maybe()
			sut := ctrlgames.Controller{
				Service:      gameServiceMock,
				AuthService:  authServiceMock,
				PartyService: partyServiceMock,
			}

			// Act
			sut.GetUserCurrentGame(ctx, 1, login)

			// Assert
			require.Equal(test, testCase.ExpectedStatus, rec.Code)

			gameServiceMock.AssertExpectations(test)
			authServiceMock.AssertExpectations(test)
		})
	}
}

// A user outside the party is turned away before anything is read or changed.
func TestCtrlGames_GetUserCurrentGame_NotPartyMember(test *testing.T) {
	// Arrange
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("login")
	ctx.SetParamValues(login)

	gameServiceMock := new(srvgamesmock.ServiceMock)
	authServiceMock := new(srvauthmock.ServiceMock)
	authServiceMock.On("GetUserId", mock.Anything).Return(1, nil)
	partyServiceMock := new(srvpartiesmock.ServiceMock)
	partyServiceMock.On("RequireMember", 1, 1).Return(common.NewPartyNotFoundError(1))
	sut := ctrlgames.Controller{
		Service:      gameServiceMock,
		AuthService:  authServiceMock,
		PartyService: partyServiceMock,
	}

	// Act
	sut.GetUserCurrentGame(ctx, 1, login)

	// Assert
	require.Equal(test, http.StatusNotFound, rec.Code)

	gameServiceMock.AssertExpectations(test)
	authServiceMock.AssertExpectations(test)
	partyServiceMock.AssertExpectations(test)
}
