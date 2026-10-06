package dbauth

import (
	"FGG-Service/src/auth/typeauth"
	"FGG-Service/src/dbaccess"
	"context"
)

type IDatabase interface {
	GetUserByLoginCommand(ctx context.Context, userLogin string) (typeauth.User, error)
	GetUserByIdCommand(ctx context.Context, userId int) (typeauth.User, error)
	GetUserByEmailCommand(ctx context.Context, userEmail string) (typeauth.User, error)
	GetUserByLoginAndPasswordCommand(ctx context.Context, loginUser typeauth.LoginUser) (typeauth.User, error)
	CreateUserCommand(ctx context.Context, signupUser typeauth.SignupUser) (typeauth.User, error)
	GetUserSessionByIdCommand(ctx context.Context, sessionId string) (typeauth.UserSession, error)
	CreateUserSessionCommand(ctx context.Context, userId int) (typeauth.UserSession, error)
	DeleteUserSessionCommand(ctx context.Context, sessionId string) error
}

type Database struct {
}

var getUserByLoginQuery = dbaccess.Query{Name: "GetUserByLoginQuery", SQL: `SELECT * FROM get_user_by_login($1::text)`}

func (db *Database) GetUserByLoginCommand(ctx context.Context, userLogin string) (user typeauth.User, err error) {
	row := dbaccess.QueryRow(ctx, getUserByLoginQuery, userLogin)

	err = row.Scan(&user.Id, &user.Login, &user.Email)

	dbaccess.LogDbResult(getUserByLoginQuery, user, err)

	return
}

var getUserByIdQuery = dbaccess.Query{Name: "GetUserByIdQuery", SQL: `SELECT * FROM get_user_by_id($1::integer)`}

func (db *Database) GetUserByIdCommand(ctx context.Context, userId int) (user typeauth.User, err error) {
	row := dbaccess.QueryRow(ctx, getUserByIdQuery, userId)

	err = row.Scan(&user.Id, &user.Login, &user.Email)

	dbaccess.LogDbResult(getUserByIdQuery, user, err)

	return
}

var getUserByEmailQuery = dbaccess.Query{Name: "GetUserByEmailQuery", SQL: `SELECT * FROM get_user_by_email($1::text)`}

func (db *Database) GetUserByEmailCommand(ctx context.Context, userEmail string) (user typeauth.User, err error) {
	row := dbaccess.QueryRow(ctx, getUserByEmailQuery, userEmail)

	err = row.Scan(&user.Id, &user.Login, &user.Email)

	dbaccess.LogDbResult(getUserByEmailQuery, user, err)

	return
}

var getUserByLoginAndPasswordQuery = dbaccess.Query{Name: "GetUserByLoginAndPasswordQuery", SQL: `SELECT * FROM get_user_by_login_and_password($1::text, $2::text)`}

func (db *Database) GetUserByLoginAndPasswordCommand(ctx context.Context, loginUser typeauth.LoginUser) (user typeauth.User, err error) {
	row := dbaccess.QueryRow(ctx, getUserByLoginAndPasswordQuery, loginUser.Login, loginUser.Password)

	err = row.Scan(&user.Id, &user.Login, &user.Email)

	dbaccess.LogDbResult(getUserByLoginAndPasswordQuery, user, err)

	return
}

var createUserQuery = dbaccess.Query{Name: "CreateUserQuery", SQL: `SELECT * FROM create_user($1::text, $2::text, $3::text)`}

func (db *Database) CreateUserCommand(ctx context.Context, signupUser typeauth.SignupUser) (user typeauth.User, err error) {
	row := dbaccess.QueryRow(ctx, createUserQuery, signupUser.Login, signupUser.Email, signupUser.Password)

	err = row.Scan(&user.Id, &user.Login, &user.Email)

	dbaccess.LogDbResult(createUserQuery, user, err)

	return
}

var getUserSessionByIdQuery = dbaccess.Query{Name: "GetUserSessionByIdQuery", SQL: `SELECT * FROM get_user_session_by_id($1::text)`}

func (db *Database) GetUserSessionByIdCommand(ctx context.Context, sessionId string) (userSession typeauth.UserSession, err error) {
	row := dbaccess.QueryRow(ctx, getUserSessionByIdQuery, sessionId)

	err = row.Scan(&userSession.Id, &userSession.UserId)

	dbaccess.LogDbResult(getUserSessionByIdQuery, userSession, err)

	return
}

var createUserSessionQuery = dbaccess.Query{Name: "CreateUserSessionQuery", SQL: `SELECT * FROM create_user_session($1::integer)`}

func (db *Database) CreateUserSessionCommand(ctx context.Context, userId int) (userSession typeauth.UserSession, err error) {
	row := dbaccess.QueryRow(ctx, createUserSessionQuery, userId)

	err = row.Scan(&userSession.Id, &userSession.UserId, &userSession.CreatedDate, &userSession.ExpiryDate)

	dbaccess.LogDbResult(createUserSessionQuery, userSession, err)

	return
}

var deleteUserSessionQuery = dbaccess.Query{Name: "DeleteUserSessionQuery", SQL: `SELECT delete_user_session($1::text)`}

func (db *Database) DeleteUserSessionCommand(ctx context.Context, sessionId string) error {
	_, err := dbaccess.Exec(ctx, deleteUserSessionQuery, sessionId)

	dbaccess.LogDbResult(deleteUserSessionQuery, nil, err)

	return err
}
