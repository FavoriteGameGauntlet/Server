package srvauth

import (
	"FGG-Service/src/auth/database"
	"FGG-Service/src/auth/types"
	"FGG-Service/src/common"
	"FGG-Service/src/parties/database"
	"database/sql"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

type IService interface {
	DoesUserSessionExist(ctx echo.Context) (bool, error)
	GetUserId(ctx echo.Context) (int, error)
	GetUserIdByLogin(login string) (int, error)
	GetLoginByUserId(userId int) (string, error)
	IsAdmin(userId int, partyId int) (bool, error)
}

type Service struct {
	Database        dbauth.IDatabase
	PartiesDatabase dbparties.IDatabase
}

func NewService() *Service {
	return &Service{
		Database:        new(dbauth.Database),
		PartiesDatabase: new(dbparties.Database),
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
	return s.Database.GetUserSessionByIdCommand(sessionId)
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
	return s.Database.DeleteUserSessionCommand(userSessionId)
}

// IsAdmin reports whether the user is an admin of the party. Admin rights are party membership data
// (parties.Members.IsAdmin) — a party creator becomes its first admin. A user who is not a member of
// the party is simply not an admin.
func (s *Service) IsAdmin(userId int, partyId int) (isAdmin bool, err error) {
	member, err := s.PartiesDatabase.GetMemberCommand(userId, partyId)

	if errors.Is(err, sql.ErrNoRows) {
		err = nil
		return
	}

	if err != nil {
		return
	}

	isAdmin = member.IsAdmin && member.LeftDate == nil

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

// GetLoginByUserId resolves a user id to the login the API names that user by.
func (s *Service) GetLoginByUserId(userId int) (login string, err error) {
	user, err := s.Database.GetUserByIdCommand(userId)

	if err != nil {
		return
	}

	return user.Login, nil
}
