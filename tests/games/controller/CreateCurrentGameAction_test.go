package ctrlgames_test

import (
	"FGG-Service/src/common"
	"FGG-Service/src/games/controller"
	"FGG-Service/src/games/types"
	"FGG-Service/tests/auth/mock"
	"FGG-Service/tests/games/mock/srvgames"
	"FGG-Service/tests/parties/srvmock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type CreateCurrentGameActionTestCase struct {
	Name           string
	Body           string
	SetupMock      func() (*srvgamesmock.ServiceMock, *srvauthmock.ServiceMock)
	ExpectedStatus int
}

var CreateCurrentGameActionTestCases = []CreateCurrentGameActionTestCase{
	{
		// GetUserId returns an error. The error will return.
		Name: "GetUserId_Error",
		Body: `{"action":"start"}`,
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
		// The body is not a JSON. 400 will return.
		Name: "InvalidBody",
		Body: `not a json`,
		SetupMock: func() (*srvgamesmock.ServiceMock, *srvauthmock.ServiceMock) {
			gameServiceMock := new(srvgamesmock.ServiceMock)
			authServiceMock := new(srvauthmock.ServiceMock)

			authServiceMock.
				On("GetUserId", mock.Anything).
				Return(1, nil)

			return gameServiceMock, authServiceMock
		},
		ExpectedStatus: http.StatusBadRequest,
	},
	{
		// The action is unknown. No service is called and 422 will return.
		Name: "UnknownAction",
		Body: `{"action":"roll"}`,
		SetupMock: func() (*srvgamesmock.ServiceMock, *srvauthmock.ServiceMock) {
			gameServiceMock := new(srvgamesmock.ServiceMock)
			authServiceMock := new(srvauthmock.ServiceMock)

			authServiceMock.
				On("GetUserId", mock.Anything).
				Return(1, nil)

			return gameServiceMock, authServiceMock
		},
		ExpectedStatus: http.StatusUnprocessableEntity,
	},
	{
		// The action is missing. No service is called and 422 will return.
		Name: "MissingAction",
		Body: `{}`,
		SetupMock: func() (*srvgamesmock.ServiceMock, *srvauthmock.ServiceMock) {
			gameServiceMock := new(srvgamesmock.ServiceMock)
			authServiceMock := new(srvauthmock.ServiceMock)

			authServiceMock.
				On("GetUserId", mock.Anything).
				Return(1, nil)

			return gameServiceMock, authServiceMock
		},
		ExpectedStatus: http.StatusUnprocessableEntity,
	},
	{
		// The start action returns ConflictError. 409 will return.
		Name: "Start_CurrentGameAlreadyExists",
		Body: `{"action":"start"}`,
		SetupMock: func() (*srvgamesmock.ServiceMock, *srvauthmock.ServiceMock) {
			gameServiceMock := new(srvgamesmock.ServiceMock)
			authServiceMock := new(srvauthmock.ServiceMock)

			authServiceMock.
				On("GetUserId", mock.Anything).
				Return(1, nil)
			gameServiceMock.
				On("StartCurrentGame", 1, 1).
				Return(typegames.CurrentGame{}, common.NewCurrentGameAlreadyExistsConflictError())

			return gameServiceMock, authServiceMock
		},
		ExpectedStatus: http.StatusConflict,
	},
	{
		// The start action succeeds. 200 will return.
		Name: "Start_SuccessReturn",
		Body: `{"action":"start"}`,
		SetupMock: func() (*srvgamesmock.ServiceMock, *srvauthmock.ServiceMock) {
			gameServiceMock := new(srvgamesmock.ServiceMock)
			authServiceMock := new(srvauthmock.ServiceMock)

			authServiceMock.
				On("GetUserId", mock.Anything).
				Return(1, nil)
			gameServiceMock.
				On("StartCurrentGame", 1, 1).
				Return(currentGame, nil)

			return gameServiceMock, authServiceMock
		},
		ExpectedStatus: http.StatusOK,
	},
	{
		// The finish action returns NotFoundError. 404 will return.
		Name: "Finish_CurrentGameNotFound",
		Body: `{"action":"finish"}`,
		SetupMock: func() (*srvgamesmock.ServiceMock, *srvauthmock.ServiceMock) {
			gameServiceMock := new(srvgamesmock.ServiceMock)
			authServiceMock := new(srvauthmock.ServiceMock)

			authServiceMock.
				On("GetUserId", mock.Anything).
				Return(1, nil)
			gameServiceMock.
				On("FinishCurrentGame", 1, 1).
				Return(typegames.CurrentGame{}, common.NewCurrentGameNotFoundError())

			return gameServiceMock, authServiceMock
		},
		ExpectedStatus: http.StatusNotFound,
	},
	{
		// The finish action succeeds. 200 will return.
		Name: "Finish_SuccessReturn",
		Body: `{"action":"finish"}`,
		SetupMock: func() (*srvgamesmock.ServiceMock, *srvauthmock.ServiceMock) {
			gameServiceMock := new(srvgamesmock.ServiceMock)
			authServiceMock := new(srvauthmock.ServiceMock)

			authServiceMock.
				On("GetUserId", mock.Anything).
				Return(1, nil)
			gameServiceMock.
				On("FinishCurrentGame", 1, 1).
				Return(currentGame, nil)

			return gameServiceMock, authServiceMock
		},
		ExpectedStatus: http.StatusOK,
	},
	{
		// The cancel action returns a database error. 500 will return.
		Name: "Cancel_DatabaseError",
		Body: `{"action":"cancel"}`,
		SetupMock: func() (*srvgamesmock.ServiceMock, *srvauthmock.ServiceMock) {
			gameServiceMock := new(srvgamesmock.ServiceMock)
			authServiceMock := new(srvauthmock.ServiceMock)

			authServiceMock.
				On("GetUserId", mock.Anything).
				Return(1, nil)
			gameServiceMock.
				On("CancelCurrentGame", 1, 1).
				Return(typegames.CurrentGame{}, dbError)

			return gameServiceMock, authServiceMock
		},
		ExpectedStatus: http.StatusInternalServerError,
	},
	{
		// The cancel action succeeds. 200 will return.
		Name: "Cancel_SuccessReturn",
		Body: `{"action":"cancel"}`,
		SetupMock: func() (*srvgamesmock.ServiceMock, *srvauthmock.ServiceMock) {
			gameServiceMock := new(srvgamesmock.ServiceMock)
			authServiceMock := new(srvauthmock.ServiceMock)

			authServiceMock.
				On("GetUserId", mock.Anything).
				Return(1, nil)
			gameServiceMock.
				On("CancelCurrentGame", 1, 1).
				Return(currentGame, nil)

			return gameServiceMock, authServiceMock
		},
		ExpectedStatus: http.StatusOK,
	},
}

func TestCtrlGames_CreateCurrentGameAction(test *testing.T) {
	for _, testCase := range CreateCurrentGameActionTestCases {
		test.Run(testCase.Name, func(test *testing.T) {
			// Arrange
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(testCase.Body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			gameServiceMock, authServiceMock := testCase.SetupMock()
			partyServiceMock := new(srvpartiesmock.ServiceMock)
			partyServiceMock.On("RequireMember", mock.Anything, 1).Return(nil).Maybe()
			sut := ctrlgames.Controller{
				Service:      gameServiceMock,
				AuthService:  authServiceMock,
				PartyService: partyServiceMock,
			}

			// Act
			sut.CreateCurrentGameAction(ctx, 1)

			// Assert
			require.Equal(test, testCase.ExpectedStatus, rec.Code)

			gameServiceMock.AssertExpectations(test)
			authServiceMock.AssertExpectations(test)
		})
	}
}

// A user outside the party is turned away before anything is read or changed.
func TestCtrlGames_CreateCurrentGameAction_NotPartyMember(test *testing.T) {
	// Arrange
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"action":"start"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

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
	sut.CreateCurrentGameAction(ctx, 1)

	// Assert
	require.Equal(test, http.StatusNotFound, rec.Code)

	gameServiceMock.AssertExpectations(test)
	authServiceMock.AssertExpectations(test)
	partyServiceMock.AssertExpectations(test)
}
