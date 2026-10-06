package dbgames

import (
	"FGG-Service/src/dbaccess"
	"FGG-Service/src/games/types"
	"time"
)

type IDatabase interface {
	GetGameByNameCommand(partyId int, name string) (game typegames.Game, err error)
	CreateGameCommand(partyId int, name string) (game typegames.Game, err error)
	GetGameCommand(partyId int, gameId int) (game typegames.Game, err error)
	GetWishlistGameCommand(userId int, partyId int, gameId int) (game typegames.WishlistGame, err error)
	CreateWishlistGameCommand(userId int, partyId int, gameId int) (game typegames.CreatedWishlistGame, err error)
	DeleteWishlistGameCommand(userId int, partyId int, gameId int) error
	GetWishlistGamesCommand(userId int, partyId int) (games typegames.WishlistGames, err error)
	CreateCurrentGameCommand(userId int, partyId int, gameId int, actorUserId int, sourceEventId *int) (game typegames.CreatedUserGame, err error)
	GetCurrentGameCommand(userId int, partyId int) (game typegames.UserGame, err error)
	ChangeGameTimeSpentCommand(userId int, partyId int, gameId int, changeValue time.Duration, actorUserId int, sourceEventId *int) error
	CancelCurrentGameCommand(userId int, partyId int, gameId int, actorUserId int, sourceEventId *int) (game typegames.UserGame, err error)
	FinishCurrentGameCommand(userId int, partyId int, gameId int, actorUserId int, sourceEventId *int) (game typegames.UserGame, err error)
	RateGameCommand(userId int, partyId int, gameId int, rating int, reviewComment *string) (gameRating typegames.GameRating, err error)
	GetGameReviewCommand(userId int, partyId int, gameId int) (review typegames.GameReview, err error)
	GetGameHistoryCommand(userId int, partyId int) (games []typegames.GameHistoryEntry, err error)
	GetAllCurrentGamesCommand(partyId int) (games []typegames.UserGameWithLogin, err error)
}

type Database struct {
}

var getGameByNameQuery = dbaccess.Query{Name: "GetGameByNameQuery", SQL: `SELECT * FROM get_game_by_name($1::integer, $2::text)`}

func (db *Database) GetGameByNameCommand(partyId int, name string) (game typegames.Game, err error) {
	row := dbaccess.QueryRow(getGameByNameQuery, partyId, name)

	err = row.Scan(&game.Id, &game.PartyId, &game.Name)

	dbaccess.LogDbResult(getGameByNameQuery, game, err)

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

var getWishlistGameQuery = dbaccess.Query{Name: "GetWishlistGameQuery", SQL: `SELECT * FROM get_wishlist_game($1::integer, $2::integer, $3::integer)`}

func (db *Database) GetWishlistGameCommand(userId int, partyId int, gameId int) (game typegames.WishlistGame, err error) {
	row := dbaccess.QueryRow(getWishlistGameQuery, userId, partyId, gameId)

	err = row.Scan(&game.Id, &game.GameId, &game.Name)

	dbaccess.LogDbResult(getWishlistGameQuery, game, err)

	return
}

var createWishlistGameQuery = dbaccess.Query{Name: "CreateWishlistGameQuery", SQL: `SELECT * FROM create_wishlist_game($1::integer, $2::integer, $3::integer)`}

func (db *Database) CreateWishlistGameCommand(userId int, partyId int, gameId int) (game typegames.CreatedWishlistGame, err error) {
	row := dbaccess.QueryRow(createWishlistGameQuery, userId, partyId, gameId)

	err = row.Scan(&game.Id, &game.UserId, &game.PartyId, &game.GameId, &game.CreatedDate)

	dbaccess.LogDbResult(createWishlistGameQuery, game, err)

	return
}

var deleteWishlistGameQuery = dbaccess.Query{Name: "DeleteWishlistGameQuery", SQL: `SELECT delete_wishlist_game($1::integer, $2::integer, $3::integer)`}

func (db *Database) DeleteWishlistGameCommand(userId int, partyId int, gameId int) error {
	_, err := dbaccess.Exec(deleteWishlistGameQuery, userId, partyId, gameId)

	dbaccess.LogDbResult(deleteWishlistGameQuery, nil, err)

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

	err = rows.Err()

	dbaccess.LogDbResult(getWishlistGamesQuery, games, err)

	_ = rows.Close()
	return
}

var createCurrentGameQuery = dbaccess.Query{Name: "CreateCurrentGameQuery", SQL: `SELECT * FROM create_user_game($1::integer, $2::integer, $3::integer, $4::integer, $5::integer)`}

func (db *Database) CreateCurrentGameCommand(userId int, partyId int, gameId int, actorUserId int, sourceEventId *int) (game typegames.CreatedUserGame, err error) {
	row := dbaccess.QueryRow(createCurrentGameQuery, userId, partyId, gameId, actorUserId, sourceEventId)

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
	err = row.Scan(&game.Id, &game.Name, &timeSpentRaw, &game.StartDate)

	if err == nil {
		game.TimeSpent, err = dbaccess.ScanInterval(timeSpentRaw)
	}

	dbaccess.LogDbResult(getCurrentGameQuery, game, err)

	return
}

var changeGameTimeSpentQuery = dbaccess.Query{Name: "ChangeGameTimeSpentQuery", SQL: `SELECT change_game_time_spent($1::integer, $2::integer, $3::integer, $4::interval, $5::integer, $6::integer)`}

func (db *Database) ChangeGameTimeSpentCommand(userId int, partyId int, gameId int, changeValue time.Duration, actorUserId int, sourceEventId *int) error {
	_, err := dbaccess.Exec(changeGameTimeSpentQuery, userId, partyId, gameId, changeValue, actorUserId, sourceEventId)

	dbaccess.LogDbResult(changeGameTimeSpentQuery, nil, err)

	return err
}

var cancelCurrentGameQuery = dbaccess.Query{Name: "CancelCurrentGameQuery", SQL: `SELECT * FROM cancel_user_game($1::integer, $2::integer, $3::integer, $4::integer, $5::integer)`}

// CancelCurrentGameCommand returns the cancelled game with the time spent on it as it was recorded in
// the history. No row means the user has no such current game.
func (db *Database) CancelCurrentGameCommand(userId int, partyId int, gameId int, actorUserId int, sourceEventId *int) (game typegames.UserGame, err error) {
	row := dbaccess.QueryRow(cancelCurrentGameQuery, userId, partyId, gameId, actorUserId, sourceEventId)

	var timeSpentRaw string
	err = row.Scan(&game.Id, &game.Name, &timeSpentRaw, &game.StartDate)

	if err == nil {
		game.TimeSpent, err = dbaccess.ScanInterval(timeSpentRaw)
	}

	dbaccess.LogDbResult(cancelCurrentGameQuery, game, err)

	return
}

var finishCurrentGameQuery = dbaccess.Query{Name: "FinishCurrentGameQuery", SQL: `SELECT * FROM finish_user_game($1::integer, $2::integer, $3::integer, $4::integer, $5::integer)`}

// FinishCurrentGameCommand returns the finished game with the time spent on it as it was recorded in
// the history. No row means the user has no such current game.
func (db *Database) FinishCurrentGameCommand(userId int, partyId int, gameId int, actorUserId int, sourceEventId *int) (game typegames.UserGame, err error) {
	row := dbaccess.QueryRow(finishCurrentGameQuery, userId, partyId, gameId, actorUserId, sourceEventId)

	var timeSpentRaw string
	err = row.Scan(&game.Id, &game.Name, &timeSpentRaw, &game.StartDate)

	if err == nil {
		game.TimeSpent, err = dbaccess.ScanInterval(timeSpentRaw)
	}

	dbaccess.LogDbResult(finishCurrentGameQuery, game, err)

	return
}

var rateGameQuery = dbaccess.Query{Name: "RateGameQuery", SQL: `SELECT * FROM rate_game($1::integer, $2::integer, $3::integer, $4::integer, $5::text)`}

// RateGameCommand upserts the rating a user left for a game. Only a game the user has already
// finished or cancelled can be rated, and rating any other one writes nothing and returns no row.
func (db *Database) RateGameCommand(userId int, partyId int, gameId int, rating int, reviewComment *string) (gameRating typegames.GameRating, err error) {
	row := dbaccess.QueryRow(rateGameQuery, userId, partyId, gameId, rating, reviewComment)

	err = row.Scan(
		&gameRating.Id,
		&gameRating.UserId,
		&gameRating.PartyId,
		&gameRating.GameId,
		&gameRating.Rating,
		&gameRating.ReviewComment,
		&gameRating.CreatedDate,
		&gameRating.UpdatedDate)

	dbaccess.LogDbResult(rateGameQuery, gameRating, err)

	return
}

var getGameReviewQuery = dbaccess.Query{Name: "GetGameReviewQuery", SQL: `SELECT * FROM get_game_review($1::integer, $2::integer, $3::integer)`}

func (db *Database) GetGameReviewCommand(userId int, partyId int, gameId int) (review typegames.GameReview, err error) {
	row := dbaccess.QueryRow(getGameReviewQuery, userId, partyId, gameId)

	err = row.Scan(&review.Rating, &review.ReviewComment)

	dbaccess.LogDbResult(getGameReviewQuery, review, err)

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

	err = rows.Err()

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
		err = rows.Scan(&game.Id, &game.Name, &timeSpentRaw, &game.Login, &game.StartDate)

		if err == nil {
			game.TimeSpent, err = dbaccess.ScanInterval(timeSpentRaw)
		}

		if err != nil {
			_ = rows.Close()
			return
		}

		games = append(games, game)
	}

	err = rows.Err()

	dbaccess.LogDbResult(getAllCurrentGamesQuery, games, err)

	_ = rows.Close()
	return
}
