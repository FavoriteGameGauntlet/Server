package ctrlgames_test

import (
	"FGG-Service/src/common"
	"FGG-Service/src/games/controller"
	"FGG-Service/src/games/types"
	"FGG-Service/tests/auth/mock"
	"FGG-Service/tests/games/mock/srvgames"
	"FGG-Service/tests/parties/srvmock"
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
		// GetUserId returns an error. The error will return.
		Name: "GetUserId_Error",
		Body: `{"name":"Half-Life 1"}`,
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
		Body: `{"name":"Half-Life 1"}`,
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
		Body: `{"name":"Half-Life 1"}`,
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
		Body: `{"name":"Half-Life 1"}`,
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
		// The request body is malformed. 400 will return.
		Name: "InvalidBody",
		Body: `{invalid}`,
		SetupMock: func() (*srvgamesmock.ServiceMock, *srvauthmock.ServiceMock) {
			gameServiceMock := new(srvgamesmock.ServiceMock)
			authServiceMock := new(srvauthmock.ServiceMock)

			authServiceMock.
				On("GetUserId", mock.Anything).
				Return(1, nil)
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
				On("GetUserId", mock.Anything).
				Return(1, nil)
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
				On("GetUserId", mock.Anything).
				Return(1, nil)
			authServiceMock.
				On("GetUserIdByLogin", login).
				Return(1, nil)
			gameServiceMock.
				On("AddWishlistGame", 1, 1, typegames.WishlistGame{Name: "Half-Life 1"}).
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
				On("GetUserId", mock.Anything).
				Return(1, nil)
			authServiceMock.
				On("GetUserIdByLogin", login).
				Return(1, nil)
			gameServiceMock.
				On("AddWishlistGame", 1, 1, typegames.WishlistGame{Name: "Half-Life 1"}).
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
				On("GetUserId", mock.Anything).
				Return(1, nil)
			authServiceMock.
				On("GetUserIdByLogin", login).
				Return(1, nil)
			gameServiceMock.
				On("AddWishlistGame", 1, 1, typegames.WishlistGame{Name: "Half-Life 1"}).
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
			partyServiceMock := new(srvpartiesmock.ServiceMock)
			partyServiceMock.On("RequireMember", mock.Anything, 1).Return(nil).Maybe()
			sut := ctrlgames.Controller{
				Service:      gameServiceMock,
				AuthService:  authServiceMock,
				PartyService: partyServiceMock,
			}

			// Act
			sut.AddUserWishlistGame(ctx, 1, login)

			// Assert
			require.Equal(test, testCase.ExpectedStatus, rec.Code)

			gameServiceMock.AssertExpectations(test)
			authServiceMock.AssertExpectations(test)
		})
	}
}

// A user outside the party is turned away before anything is read or changed.
func TestCtrlGames_AddUserWishlistGame_NotPartyMember(test *testing.T) {
	// Arrange
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"Half-Life 1"}`))
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
	sut.AddUserWishlistGame(ctx, 1, login)

	// Assert
	require.Equal(test, http.StatusNotFound, rec.Code)

	gameServiceMock.AssertExpectations(test)
	authServiceMock.AssertExpectations(test)
	partyServiceMock.AssertExpectations(test)
}
