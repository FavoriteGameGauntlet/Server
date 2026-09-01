package dbauth

import (
	"FGG-Service/src/auth/types"
	"FGG-Service/src/dbaccess"
)

type IDatabase interface {
	GetUserByLoginCommand(userLogin string) (typeauth.User, error)
	GetUserByIdCommand(userId int) (typeauth.User, error)
	GetUserByEmailCommand(userEmail string) (typeauth.User, error)
	GetUserByLoginAndPasswordCommand(loginUser typeauth.LoginUser) (typeauth.User, error)
	CreateUserCommand(signupUser typeauth.SignupUser) (typeauth.User, error)
	GetUserSessionByIdCommand(sessionId string) (typeauth.UserSession, error)
	CreateUserSessionCommand(userId int) (typeauth.UserSession, error)
	DeleteUserSessionCommand(sessionId string) error
}

type Database struct {
}

var getUserByLoginQuery = dbaccess.Query{Name: "GetUserByLoginQuery", SQL: `SELECT * FROM get_user_by_login($1::text)`}

func (db *Database) GetUserByLoginCommand(userLogin string) (user typeauth.User, err error) {
	row := dbaccess.QueryRow(getUserByLoginQuery, userLogin)

	err = row.Scan(&user.Id, &user.Login, &user.Email)

	dbaccess.LogDbResult(getUserByLoginQuery, user, err)

	return
}

var getUserByIdQuery = dbaccess.Query{Name: "GetUserByIdQuery", SQL: `SELECT * FROM get_user_by_id($1::integer)`}

func (db *Database) GetUserByIdCommand(userId int) (user typeauth.User, err error) {
	row := dbaccess.QueryRow(getUserByIdQuery, userId)

	err = row.Scan(&user.Id, &user.Login, &user.Email)

	dbaccess.LogDbResult(getUserByIdQuery, user, err)

	return
}

var getUserByEmailQuery = dbaccess.Query{Name: "GetUserByEmailQuery", SQL: `SELECT * FROM get_user_by_email($1::text)`}

func (db *Database) GetUserByEmailCommand(userEmail string) (user typeauth.User, err error) {
	row := dbaccess.QueryRow(getUserByEmailQuery, userEmail)

	err = row.Scan(&user.Id, &user.Login, &user.Email)

	dbaccess.LogDbResult(getUserByEmailQuery, user, err)

	return
}

var getUserByLoginAndPasswordQuery = dbaccess.Query{Name: "GetUserByLoginAndPasswordQuery", SQL: `SELECT * FROM get_user_by_login_and_password($1::text, $2::text)`}

func (db *Database) GetUserByLoginAndPasswordCommand(loginUser typeauth.LoginUser) (user typeauth.User, err error) {
	row := dbaccess.QueryRow(getUserByLoginAndPasswordQuery, loginUser.Login, loginUser.Password)

	err = row.Scan(&user.Id, &user.Login, &user.Email)

	dbaccess.LogDbResult(getUserByLoginAndPasswordQuery, user, err)

	return
}

var createUserQuery = dbaccess.Query{Name: "CreateUserQuery", SQL: `SELECT * FROM create_user($1::text, $2::text, $3::text)`}

func (db *Database) CreateUserCommand(signupUser typeauth.SignupUser) (user typeauth.User, err error) {
	row := dbaccess.QueryRow(createUserQuery, signupUser.Login, signupUser.Email, signupUser.Password)

	err = row.Scan(&user.Id, &user.Login, &user.Email)

	dbaccess.LogDbResult(createUserQuery, user, err)

	return
}

var getUserSessionByIdQuery = dbaccess.Query{Name: "GetUserSessionByIdQuery", SQL: `SELECT * FROM get_user_session_by_id($1::text)`}

func (db *Database) GetUserSessionByIdCommand(sessionId string) (userSession typeauth.UserSession, err error) {
	row := dbaccess.QueryRow(getUserSessionByIdQuery, sessionId)

	err = row.Scan(&userSession.Id, &userSession.UserId)

	dbaccess.LogDbResult(getUserSessionByIdQuery, userSession, err)

	return
}

var createUserSessionQuery = dbaccess.Query{Name: "CreateUserSessionQuery", SQL: `SELECT * FROM create_user_session($1::integer)`}

func (db *Database) CreateUserSessionCommand(userId int) (userSession typeauth.UserSession, err error) {
	row := dbaccess.QueryRow(createUserSessionQuery, userId)

	err = row.Scan(&userSession.Id, &userSession.UserId, &userSession.CreatedDate, &userSession.ExpiryDate)

	dbaccess.LogDbResult(createUserSessionQuery, userSession, err)

	return
}

var deleteUserSessionQuery = dbaccess.Query{Name: "DeleteUserSessionQuery", SQL: `SELECT delete_user_session($1::text)`}

func (db *Database) DeleteUserSessionCommand(sessionId string) error {
	_, err := dbaccess.Exec(deleteUserSessionQuery, sessionId)

	dbaccess.LogDbResult(deleteUserSessionQuery, nil, err)

	return err
}
