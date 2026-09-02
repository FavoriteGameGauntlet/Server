package dbgames

import (
	"FGG-Service/src/dbaccess"
	"FGG-Service/src/games/types"
	"database/sql"
	"errors"
	"time"
)

type IDatabase interface {
	DoesGameExistCommand(partyId int, name string) (doesExist bool, err error)
	CreateGameCommand(partyId int, name string) (game typegames.Game, err error)
	GetGameCommand(partyId int, gameId int) (game typegames.Game, err error)
	GetWishlistGameCommand(name string) (game typegames.WishlistGame, err error)
	DoesWishlistGameExistCommand(userId int, partyId int, gameId int) (doesExist bool, err error)
	CreateWishlistGameCommand(userId int, partyId int, gameId int) (game typegames.CreatedWishlistGame, err error)
	DeleteUnplayedGameCommand(userId int, partyId int, gameId int) error
	GetWishlistGamesCommand(userId int, partyId int) (games typegames.WishlistGames, err error)
	CreateCurrentGameCommand(userId int, partyId int, gameId int, sourceEventId *int) (game typegames.CreatedUserGame, err error)
	GetCurrentGameCommand(userId int, partyId int) (game typegames.UserGame, err error)
	DoesUserGameExistCommand(userId int, partyId int) (doesExist bool, err error)
	GetGameTimeSpentCommand(userId int, gameId int) (timeSpent time.Duration, err error)
	ChangeGameTimeSpentCommand(userId int, partyId int, gameId int, changeValue time.Duration, sourceEventId *int) error
	CancelCurrentGameCommand(userId int, partyId int, gameId int, sourceEventId *int) error
	FinishCurrentGameCommand(userId int, partyId int, gameId int, sourceEventId *int) error
	RateGameCommand(userId int, partyId int, gameId int, rating int, reviewComment *string, sourceEventId *int) error
	GetGameReviewCommand(userId int, partyId int, gameId int) (reviewComment *string, err error)
	GetGameHistoryCommand(userId int, partyId int) (games []typegames.GameHistoryEntry, err error)
	GetAllCurrentGamesCommand(partyId int) (games []typegames.UserGameWithLogin, err error)
}

type Database struct {
}

var doesGameExistQuery = dbaccess.Query{Name: "DoesGameExistQuery", SQL: `SELECT does_game_exist($1::integer, $2::text)`}

func (db *Database) DoesGameExistCommand(partyId int, name string) (doesExist bool, err error) {
	row := dbaccess.QueryRow(doesGameExistQuery, partyId, name)

	err = row.Scan(&doesExist)

	dbaccess.LogDbResult(doesGameExistQuery, doesExist, err)

	return
}

var createGameQuery = dbaccess.Query{Name: "CreateGameQuery", SQL: `SELECT * FROM create_game($1::integer, $2::text)`}

func (db *Database) CreateGameCommand(partyId int, name string) (game typegames.Game, err error) {
	row := dbaccess.QueryRow(createGameQuery, partyId, name)

	err = row.Scan(&game.Id, &game.PartyId, &game.Name)

	dbaccess.LogDbResult(createGameQuery, game, err)

	return
}

var getGameQuery = dbaccess.Query{Name: "GetGameQuery", SQL: `SELECT * FROM get_game($1::integer, $2::integer)`}

func (db *Database) GetGameCommand(partyId int, gameId int) (game typegames.Game, err error) {
	row := dbaccess.QueryRow(getGameQuery, partyId, gameId)

	err = row.Scan(&game.Id, &game.PartyId, &game.Name)

	dbaccess.LogDbResult(getGameQuery, game, err)

	return
}

var getWishlistGameQuery = dbaccess.Query{Name: "GetWishlistGameQuery", SQL: `SELECT * FROM get_wishlist_game($1::text)`}

func (db *Database) GetWishlistGameCommand(name string) (game typegames.WishlistGame, err error) {
	row := dbaccess.QueryRow(getWishlistGameQuery, name)

	err = row.Scan(&game.GameId, &game.Name)

	dbaccess.LogDbResult(getWishlistGameQuery, game, err)

	return
}

var doesUnplayedGameExistQuery = dbaccess.Query{Name: "DoesUnplayedGameExistQuery", SQL: `SELECT does_wishlist_game_exist($1::integer, $2::integer, $3::integer)`}

func (db *Database) DoesWishlistGameExistCommand(userId int, partyId int, gameId int) (doesExist bool, err error) {
	row := dbaccess.QueryRow(doesUnplayedGameExistQuery, userId, partyId, gameId)

	err = row.Scan(&doesExist)

	dbaccess.LogDbResult(doesUnplayedGameExistQuery, doesExist, err)

	return
}

var createUnplayedGameQuery = dbaccess.Query{Name: "CreateUnplayedGameQuery", SQL: `SELECT * FROM create_wishlist_game($1::integer, $2::integer, $3::integer)`}

func (db *Database) CreateWishlistGameCommand(userId int, partyId int, gameId int) (game typegames.CreatedWishlistGame, err error) {
	row := dbaccess.QueryRow(createUnplayedGameQuery, userId, partyId, gameId)

	err = row.Scan(&game.Id, &game.UserId, &game.PartyId, &game.GameId, &game.CreatedDate)

	dbaccess.LogDbResult(createUnplayedGameQuery, game, err)

	return
}

var deleteUnplayedGameQuery = dbaccess.Query{Name: "DeleteUnplayedGameQuery", SQL: `SELECT delete_wishlist_game($1::integer, $2::integer, $3::integer)`}

func (db *Database) DeleteUnplayedGameCommand(userId int, partyId int, gameId int) error {
	_, err := dbaccess.Exec(deleteUnplayedGameQuery, userId, partyId, gameId)

	dbaccess.LogDbResult(deleteUnplayedGameQuery, nil, err)

	return err
}

var getWishlistGamesQuery = dbaccess.Query{Name: "GetWishlistGamesQuery", SQL: `SELECT * FROM get_wishlist_games($1::integer, $2::integer)`}

func (db *Database) GetWishlistGamesCommand(userId int, partyId int) (games typegames.WishlistGames, err error) {
	rows, err := dbaccess.QueryRows(getWishlistGamesQuery, userId, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		game := typegames.WishlistGame{}
		err = rows.Scan(&game.Id, &game.GameId, &game.Name)

		if err != nil {
			_ = rows.Close()
			return
		}

		games = append(games, game)
	}

	dbaccess.LogDbResult(getWishlistGamesQuery, games, err)

	_ = rows.Close()
	return
}

var createCurrentGameQuery = dbaccess.Query{Name: "CreateCurrentGameQuery", SQL: `SELECT * FROM create_user_game($1::integer, $2::integer, $3::integer, $4::integer)`}

func (db *Database) CreateCurrentGameCommand(userId int, partyId int, gameId int, sourceEventId *int) (game typegames.CreatedUserGame, err error) {
	row := dbaccess.QueryRow(createCurrentGameQuery, userId, partyId, gameId, sourceEventId)

	var timeSpentRaw string
	err = row.Scan(&game.Id, &game.UserId, &game.PartyId, &game.GameId, &timeSpentRaw, &game.StartedDate)

	if err == nil {
		game.TimeSpent, err = dbaccess.ScanInterval(timeSpentRaw)
	}

	dbaccess.LogDbResult(createCurrentGameQuery, game, err)

	return
}

var getCurrentGameQuery = dbaccess.Query{Name: "GetCurrentGameQuery", SQL: `SELECT * FROM get_user_game($1::integer, $2::integer)`}

func (db *Database) GetCurrentGameCommand(userId int, partyId int) (game typegames.UserGame, err error) {
	row := dbaccess.QueryRow(getCurrentGameQuery, userId, partyId)

	var timeSpentRaw string
	err = row.Scan(&game.Id, &game.Name, &timeSpentRaw)

	if err == nil {
		game.TimeSpent, err = dbaccess.ScanInterval(timeSpentRaw)
	}

	dbaccess.LogDbResult(getCurrentGameQuery, game, err)

	return
}

var doesUserGameExistQuery = dbaccess.Query{Name: "DoesUserGameExistQuery", SQL: `SELECT does_user_game_exist($1::integer, $2::integer)`}

func (db *Database) DoesUserGameExistCommand(userId int, partyId int) (doesExist bool, err error) {
	row := dbaccess.QueryRow(doesUserGameExistQuery, userId, partyId)

	err = row.Scan(&doesExist)

	dbaccess.LogDbResult(doesUserGameExistQuery, doesExist, err)

	return
}

var getGameSecondsSpentQuery = dbaccess.Query{Name: "GetGameSecondsSpentQuery", SQL: `SELECT get_game_seconds_spent($1::integer, $2::integer)`}

func (db *Database) GetGameTimeSpentCommand(userId int, gameId int) (timeSpent time.Duration, err error) {
	row := dbaccess.QueryRow(getGameSecondsSpentQuery, userId, gameId)

	var secondsSpent int
	err = row.Scan(&secondsSpent)

	if errors.Is(err, sql.ErrNoRows) {
		dbaccess.LogDbResult(getGameSecondsSpentQuery, timeSpent, err)

		err = nil
		return
	}

	if err != nil {
		dbaccess.LogDbResult(getGameSecondsSpentQuery, timeSpent, err)

		return
	}

	timeSpent = time.Duration(secondsSpent) * time.Second

	dbaccess.LogDbResult(getGameSecondsSpentQuery, timeSpent, err)

	return
}

var changeGameTimeSpentQuery = dbaccess.Query{Name: "ChangeGameTimeSpentQuery", SQL: `SELECT change_game_time_spent($1::integer, $2::integer, $3::integer, $4::interval, $5::integer)`}

func (db *Database) ChangeGameTimeSpentCommand(userId int, partyId int, gameId int, changeValue time.Duration, sourceEventId *int) error {
	_, err := dbaccess.Exec(changeGameTimeSpentQuery, userId, partyId, gameId, changeValue, sourceEventId)

	dbaccess.LogDbResult(changeGameTimeSpentQuery, nil, err)

	return err
}

var cancelCurrentGameQuery = dbaccess.Query{Name: "CancelCurrentGameQuery", SQL: `SELECT cancel_user_game($1::integer, $2::integer, $3::integer, $4::integer)`}

func (db *Database) CancelCurrentGameCommand(userId int, partyId int, gameId int, sourceEventId *int) error {
	_, err := dbaccess.Exec(cancelCurrentGameQuery, userId, partyId, gameId, sourceEventId)

	dbaccess.LogDbResult(cancelCurrentGameQuery, nil, err)

	return err
}

var finishCurrentGameQuery = dbaccess.Query{Name: "FinishCurrentGameQuery", SQL: `SELECT finish_user_game($1::integer, $2::integer, $3::integer, $4::integer)`}

func (db *Database) FinishCurrentGameCommand(userId int, partyId int, gameId int, sourceEventId *int) error {
	_, err := dbaccess.Exec(finishCurrentGameQuery, userId, partyId, gameId, sourceEventId)

	dbaccess.LogDbResult(finishCurrentGameQuery, nil, err)

	return err
}

var rateGameQuery = dbaccess.Query{Name: "RateGameQuery", SQL: `SELECT rate_game($1::integer, $2::integer, $3::integer, $4::integer, $5::text, $6::integer)`}

func (db *Database) RateGameCommand(userId int, partyId int, gameId int, rating int, reviewComment *string, sourceEventId *int) error {
	_, err := dbaccess.Exec(rateGameQuery, userId, partyId, gameId, rating, reviewComment, sourceEventId)

	dbaccess.LogDbResult(rateGameQuery, nil, err)

	return err
}

var getGameReviewQuery = dbaccess.Query{Name: "GetGameReviewQuery", SQL: `SELECT * FROM get_game_review($1::integer, $2::integer, $3::integer)`}

func (db *Database) GetGameReviewCommand(userId int, partyId int, gameId int) (reviewComment *string, err error) {
	row := dbaccess.QueryRow(getGameReviewQuery, userId, partyId, gameId)

	err = row.Scan(&reviewComment)

	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}

	dbaccess.LogDbResult(getGameReviewQuery, reviewComment, err)

	return
}

var getGameHistoryQuery = dbaccess.Query{Name: "GetGameHistoryQuery", SQL: `SELECT * FROM get_game_history($1::integer, $2::integer)`}

func (db *Database) GetGameHistoryCommand(userId int, partyId int) (games []typegames.GameHistoryEntry, err error) {
	rows, err := dbaccess.QueryRows(getGameHistoryQuery, userId, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		entry := typegames.GameHistoryEntry{}
		var timeSpentRaw string
		err = rows.Scan(
			&entry.Id,
			&entry.GameId,
			&entry.Name,
			&entry.Action,
			&timeSpentRaw,
			&entry.Rating,
			&entry.ReviewComment,
			&entry.EndState,
			&entry.SourceEventId,
			&entry.CreatedDate)

		if err == nil {
			entry.TimeSpent, err = dbaccess.ScanInterval(timeSpentRaw)
		}

		if err != nil {
			_ = rows.Close()
			return
		}

		games = append(games, entry)
	}

	dbaccess.LogDbResult(getGameHistoryQuery, games, err)

	_ = rows.Close()
	return
}

var getAllCurrentGamesQuery = dbaccess.Query{Name: "GetAllCurrentGamesQuery", SQL: `SELECT * FROM get_all_user_games($1::integer)`}

func (db *Database) GetAllCurrentGamesCommand(partyId int) (games []typegames.UserGameWithLogin, err error) {
	rows, err := dbaccess.QueryRows(getAllCurrentGamesQuery, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		game := typegames.UserGameWithLogin{}
		var timeSpentRaw string
		err = rows.Scan(&game.Id, &game.Name, &timeSpentRaw, &game.Login)

		if err == nil {
			game.TimeSpent, err = dbaccess.ScanInterval(timeSpentRaw)
		}

		if err != nil {
			_ = rows.Close()
			return
		}

		games = append(games, game)
	}

	dbaccess.LogDbResult(getAllCurrentGamesQuery, games, err)

	_ = rows.Close()
	return
}
