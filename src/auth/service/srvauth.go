package srvauth

import (
	"FGG-Service/src/auth/database"
	"FGG-Service/src/auth/types"
	"FGG-Service/src/common"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
)

type IService interface {
	DoesUserSessionExist(ctx echo.Context) (bool, error)
	GetUserId(ctx echo.Context) (int, error)
	GetUserIdByLogin(login string) (int, error)
	IsAdmin(userId int) (bool, error)
}

type Service struct {
	Database dbauth.IDatabase
}

func NewService() *Service {
	return &Service{
		Database: new(dbauth.Database),
	}
}

func (s *Service) DoesUserSessionExist(ctx echo.Context) (doesExist bool, err error) {
	cookie, err := s.GetSessionCookie(ctx)

	if err != nil {
		err = common.NewCookieNotFoundUnauthorizedError()
		return
	}

	sessionId := cookie.Value

	_, err = s.GetUserSessionById(sessionId)

	if errors.Is(err, sql.ErrNoRows) {
		err = nil
		return
	}

	if err != nil {
		return
	}

	doesExist = true
	return
}

func (s *Service) GetSessionCookie(ctx echo.Context) (*http.Cookie, error) {
	cookie, err := ctx.Cookie(common.SessionCookieName)

	if err != nil {
		return nil, common.NewCookieNotFoundUnauthorizedError()
	}

	return cookie, nil
}

func (s *Service) GetUserId(ctx echo.Context) (userId int, err error) {
	cookie, err := ctx.Cookie(common.SessionCookieName)

	if err != nil {
		err = common.NewCookieNotFoundUnauthorizedError()
		return
	}

	sessionId := cookie.Value

	userSession, err := s.GetUserSessionById(sessionId)

	if errors.Is(err, sql.ErrNoRows) {
		err = common.NewActiveSessionNotFoundUnauthorizedError()
		return
	}

	if err != nil {
		return
	}

	userId = userSession.UserId

	return
}

func (s *Service) CreateUser(signupUser typeauth.SignupUser) error {
	user, err := s.Database.GetUserByLoginCommand(signupUser.Login)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if user.Login != "" {
		return common.NewUserNameAlreadyExistsConflictError()
	}

	user, err = s.Database.GetUserByEmailCommand(signupUser.Email)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if user.Email != "" {
		return common.NewUserEmailAlreadyExistsConflictError()
	}

	_, err = s.Database.CreateUserCommand(signupUser)

	return err
}

func (s *Service) GetUserSessionById(sessionId string) (userSession typeauth.UserSession, err error) {
	userSession, err = s.Database.GetUserSessionByIdCommand(sessionId)

	return
}

func (s *Service) CreateSession(loginUser typeauth.LoginUser) (userSession typeauth.UserSession, err error) {
	user, err := s.Database.GetUserByLoginAndPasswordCommand(loginUser)

	if errors.Is(err, sql.ErrNoRows) {
		err = common.NewWrongDataUnprocessableError()
		return
	}

	if err != nil {
		return
	}

	userSession, err = s.Database.CreateUserSessionCommand(user.Id)

	return
}

func (s *Service) DeleteUserSession(userSessionId string) error {
	err := s.Database.DeleteUserSessionCommand(userSessionId)

	return err
}

func (s *Service) IsAdmin(userId int) (isAdmin bool, err error) {
	user, err := s.Database.GetUserByIdCommand(userId)

	if err != nil {
		return
	}

	for _, adminLogin := range strings.Split(os.Getenv(common.AdminLoginsEnvVar), common.AdminLoginsSeparator) {
		if strings.TrimSpace(adminLogin) == user.Login {
			isAdmin = true
			return
		}
	}

	return
}

func (s *Service) GetUserIdByLogin(userLogin string) (userId int, err error) {
	user, err := s.Database.GetUserByLoginCommand(userLogin)

	if errors.Is(err, sql.ErrNoRows) {
		err = common.NewUserLoginNotFoundError(userLogin)
		return
	}

	if err != nil {
		return
	}

	userId = user.Id
	return
}
