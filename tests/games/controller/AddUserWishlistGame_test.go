package ctrlgames_test

import (
	"FGG-Service/src/common"
	"FGG-Service/src/games/controller"
	"FGG-Service/src/games/types"
	"FGG-Service/tests/auth/mock"
	"FGG-Service/tests/games/mock/srvgames"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var dbError = errors.New("database connection lost")

const login = "testuser"

type AddUserWishlistGameTestCase struct {
	Name           string
	Body           string
	SetupMock      func() (*srvgamesmock.ServiceMock, *srvauthmock.ServiceMock)
	ExpectedStatus int
}

var AddUserWishlistGameTestCases = []AddUserWishlistGameTestCase{
	{
		// DoesUserSessionExist returns an error. The error will return.
		Name: "DoesUserSessionExist_Error",
		Body: `{"name":"Half-Life 1"}`,
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
		Body: `{"name":"Half-Life 1"}`,
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
		Body: `{"name":"Half-Life 1"}`,
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
		Body: `{"name":"Half-Life 1"}`,
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
		// The request body is malformed. 400 will return.
		Name: "InvalidBody",
		Body: `{invalid}`,
		SetupMock: func() (*srvgamesmock.ServiceMock, *srvauthmock.ServiceMock) {
			gameServiceMock := new(srvgamesmock.ServiceMock)
			authServiceMock := new(srvauthmock.ServiceMock)

			authServiceMock.
				On("DoesUserSessionExist", mock.Anything).
				Return(true, nil)
			authServiceMock.
				On("GetUserIdByLogin", login).
				Return(1, nil)

			return gameServiceMock, authServiceMock
		},
		ExpectedStatus: http.StatusBadRequest,
	},
	{
		// The game name is empty. 422 will return.
		Name: "InvalidGameName",
		Body: `{"name":""}`,
		SetupMock: func() (*srvgamesmock.ServiceMock, *srvauthmock.ServiceMock) {
			gameServiceMock := new(srvgamesmock.ServiceMock)
			authServiceMock := new(srvauthmock.ServiceMock)

			authServiceMock.
				On("DoesUserSessionExist", mock.Anything).
				Return(true, nil)
			authServiceMock.
				On("GetUserIdByLogin", login).
				Return(1, nil)

			return gameServiceMock, authServiceMock
		},
		ExpectedStatus: http.StatusUnprocessableEntity,
	},
	{
		// AddWishlistGame returns ConflictError. 409 will return.
		Name: "GameAlreadyExists",
		Body: `{"name":"Half-Life 1"}`,
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
				On("AddWishlistGame", 1, typegames.WishlistGame{Name: "Half-Life 1"}).
				Return(common.NewWishlistGameAlreadyExistsConflictError("Half-Life 1"))

			return gameServiceMock, authServiceMock
		},
		ExpectedStatus: http.StatusConflict,
	},
	{
		// AddWishlistGame returns a database error. 500 will return.
		Name: "AddWishlistGame_DatabaseError",
		Body: `{"name":"Half-Life 1"}`,
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
				On("AddWishlistGame", 1, typegames.WishlistGame{Name: "Half-Life 1"}).
				Return(dbError)

			return gameServiceMock, authServiceMock
		},
		ExpectedStatus: http.StatusInternalServerError,
	},
	{
		// Everything succeeds. 201 will return.
		Name: "SuccessReturn",
		Body: `{"name":"Half-Life 1"}`,
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
				On("AddWishlistGame", 1, typegames.WishlistGame{Name: "Half-Life 1"}).
				Return(nil)

			return gameServiceMock, authServiceMock
		},
		ExpectedStatus: http.StatusCreated,
	},
}

func TestCtrlGames_AddUserWishlistGame(test *testing.T) {
	for _, testCase := range AddUserWishlistGameTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(testCase.Body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
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
			sut.AddUserWishlistGame(ctx, login)

			// Assert
			require.Equal(test, testCase.ExpectedStatus, rec.Code)

			gameServiceMock.AssertExpectations(test)
			authServiceMock.AssertExpectations(test)
		})
	}
}
