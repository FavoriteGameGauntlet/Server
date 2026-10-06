package ctrlauth

import (
	"FGG-Service/api/generated/auth"
	"FGG-Service/src/auth/service"
	typeauth "FGG-Service/src/auth/types"
	"FGG-Service/src/common"
	"FGG-Service/src/validator"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type Controller struct {
	Service srvauth.Service
}

func NewController() *Controller {
	s := srvauth.NewService()

	return &Controller{
		Service: *s,
	}
}

// Login (POST /auth/login)
func (c *Controller) Login(ctx echo.Context) error {
	var loginUserDto genauth.LoginUser
	err := ctx.Bind(&loginUserDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	doesExist, _ := c.Service.DoesUserSessionExist(ctx)

	if doesExist {
		err = common.NewSessionAlreadyExistsConflictError()
		return common.SendJSONErrorResponse(ctx, err)
	}

	loginUser := convertDtoToLoginUser(loginUserDto)

	userSession, err := c.Service.CreateSession(loginUser)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	cookie := createSessionCookie(userSession.Id)
	ctx.SetCookie(cookie)

	return ctx.NoContent(http.StatusCreated)
}

func convertDtoToLoginUser(userDto genauth.LoginUser) typeauth.LoginUser {
	return typeauth.LoginUser{
		Login:    userDto.Login,
		Password: typeauth.Password{Value: userDto.Password},
	}
}

func createSessionCookie(sessionId string) *http.Cookie {
	cookie := new(http.Cookie)
	cookie.Name = common.SessionCookieName
	cookie.Value = sessionId
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.HttpOnly = true
	cookie.Path = "/"
	cookie.Secure = false

	return cookie
}

// Logout (POST /auth/logout)
func (c *Controller) Logout(ctx echo.Context) error {
	cookie, err := c.Service.GetSessionCookie(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	sessionId := cookie.Value
	err = c.Service.DeleteUserSession(sessionId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	cookie.MaxAge = -1
	ctx.SetCookie(cookie)

	return ctx.NoContent(http.StatusNoContent)
}

// SignUp (POST /auth/signup)
func (c *Controller) SignUp(ctx echo.Context) error {
	var signupUserDto genauth.SignupUser
	err := ctx.Bind(&signupUserDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	signupUser := convertDtoToSignupUser(signupUserDto)

	err = validator.ValidateUserLogin(signupUser.Login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = validator.ValidateEmail(signupUser.Email)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = validator.ValidatePassword(signupUser.Password.Value)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.CreateUser(signupUser)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusCreated)
}

func convertDtoToSignupUser(userDto genauth.SignupUser) typeauth.SignupUser {
	return typeauth.SignupUser{
		Email:    userDto.Email,
		Login:    userDto.Login,
		Password: typeauth.Password{Value: userDto.Password},
	}
}
