package srvgames

import (
	"FGG-Service/src/common"
	"FGG-Service/src/games/database"
	"FGG-Service/src/games/types"
	"FGG-Service/src/sysparams/service"
	"FGG-Service/src/sysparams/types"
	"FGG-Service/src/timers/service"
	"database/sql"
	"errors"
	"math/rand"
)

type IService interface {
	GetCurrentGame(userId int, partyId int) (typegames.CurrentGame, error)
	CancelCurrentGame(userId int, partyId int) (typegames.CurrentGame, error)
	FinishCurrentGame(userId int, partyId int) (typegames.CurrentGame, error)
	StartCurrentGame(userId int, partyId int) (typegames.CurrentGame, error)
	GetGameHistory(userId int, partyId int) ([]typegames.GameHistoryEntry, error)
	RateGame(userId int, partyId int, name string, rating int, reviewComment *string) (bool, error)
	GetGameReview(userId int, partyId int, name string) (typegames.GameReview, error)
	GetWishlistGames(userId int, partyId int) (typegames.WishlistGames, error)
	AddWishlistGame(userId int, partyId int, wishlistGame typegames.WishlistGame) error
	GetAllCurrentGames(partyId int) ([]typegames.CurrentGameWithLogin, error)
}

type Service struct {
	Database         dbgames.IDatabase
	TimerService     srvtimers.IService
	GettingService   IGettingService
	SysParamsService srvsysparams.IService
}

func NewService(ts srvtimers.IService) *Service {
	db := new(dbgames.Database)
	gs := NewGettingService()
	sps := srvsysparams.NewService()

	return &Service{
		Database:         db,
		TimerService:     ts,
		GettingService:   gs,
		SysParamsService: sps,
	}
}

// AddWishlistGame puts a game on the user's wishlist, registering the game in the party first if
// nobody has named it before.
func (s *Service) AddWishlistGame(userId int, partyId int, wishlistGame typegames.WishlistGame) error {
	game, err := s.Database.GetGameByNameCommand(partyId, wishlistGame.Name)

	if errors.Is(err, sql.ErrNoRows) {
		game, err = s.Database.CreateGameCommand(partyId, wishlistGame.Name)
	}

	if err != nil {
		return err
	}

	_, err = s.Database.GetWishlistGameCommand(userId, partyId, game.Id)

	// A row means the game is already on the wishlist; only its absence lets the insert through.
	if err == nil {
		return common.NewWishlistGameAlreadyExistsConflictError(wishlistGame.Name)
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	_, err = s.Database.CreateWishlistGameCommand(userId, partyId, game.Id)

	return err
}

func (s *Service) GetWishlistGames(userId int, partyId int) (typegames.WishlistGames, error) {
	return s.GettingService.GetWishlistGames(userId, partyId)
}

func (s *Service) GetCurrentGame(userId int, partyId int) (typegames.CurrentGame, error) {
	return s.GettingService.GetCurrentGame(userId, partyId)
}

// CancelCurrentGame cancels the current game and returns it with the time spent on it in total.
func (s *Service) CancelCurrentGame(userId int, partyId int) (typegames.CurrentGame, error) {
	game, err := s.GettingService.GetCurrentGame(userId, partyId)

	if err != nil {
		return typegames.CurrentGame{}, err
	}

	_, err = s.TimerService.ForceStopCurrentTimer(userId, partyId)

	if err != nil {
		return typegames.CurrentGame{}, err
	}

	cancelledGame, err := s.Database.CancelCurrentGameCommand(userId, partyId, game.Id, userId, nil)

	if err != nil {
		return typegames.CurrentGame{}, err
	}

	return typegames.CurrentGame{
		Id:        cancelledGame.Id,
		Name:      cancelledGame.Name,
		TimeSpent: cancelledGame.TimeSpent,
		StartDate: cancelledGame.StartDate,
	}, nil
}

// FinishCurrentGame finishes the current game and returns it with the time spent on it in total.
func (s *Service) FinishCurrentGame(userId int, partyId int) (typegames.CurrentGame, error) {
	game, err := s.GettingService.GetCurrentGame(userId, partyId)

	if err != nil {
		return typegames.CurrentGame{}, err
	}

	// The current timer's time is added to the game when it is stopped below, so it counts too.
	timerTimeSpent, err := s.TimerService.GetCurrentTimerTimeSpent(userId, partyId)

	if err != nil {
		return typegames.CurrentGame{}, err
	}

	if game.TimeSpent+timerTimeSpent == 0 {
		return typegames.CurrentGame{}, common.NewGameTimeSpentIsZeroError()
	}

	_, err = s.TimerService.ForceStopCurrentTimer(userId, partyId)

	if err != nil {
		return typegames.CurrentGame{}, err
	}

	finishedGame, err := s.Database.FinishCurrentGameCommand(userId, partyId, game.Id, userId, nil)

	if err != nil {
		return typegames.CurrentGame{}, err
	}

	return typegames.CurrentGame{
		Id:        finishedGame.Id,
		Name:      finishedGame.Name,
		TimeSpent: finishedGame.TimeSpent,
		StartDate: finishedGame.StartDate,
	}, nil
}

// GetGameHistory returns the recorded game events of a user. A history entry is not a current game:
// it carries what happened to the game rather than how it stands now.
func (s *Service) GetGameHistory(userId int, partyId int) (history []typegames.GameHistoryEntry, err error) {
	return s.Database.GetGameHistoryCommand(userId, partyId)
}

// RateGame records a rating and an optional review for a game the user has already played. Rating
// the same game again overwrites the previous rating; created tells whether this was the first.
func (s *Service) RateGame(userId int, partyId int, name string, rating int, reviewComment *string) (created bool, err error) {
	gameId, err := s.getPlayedGameId(userId, partyId, name)

	if err != nil {
		return
	}

	gameRating, err := s.Database.RateGameCommand(userId, partyId, gameId, rating, reviewComment)

	// The game is in the history but hasn't ended yet, so there is nothing to rate and the
	// database writes nothing.
	if errors.Is(err, sql.ErrNoRows) {
		err = common.NewPlayedGameNotFoundError(name)
	}

	if err != nil {
		return
	}

	// Both dates are set to the same NOW() on insert, and only UpdatedDate moves on an update.
	return gameRating.CreatedDate.Equal(gameRating.UpdatedDate), nil
}

// GetGameReview returns the rating and review the user left for a game they have played.
func (s *Service) GetGameReview(userId int, partyId int, name string) (review typegames.GameReview, err error) {
	gameId, err := s.getPlayedGameId(userId, partyId, name)

	if err != nil {
		return
	}

	review, err = s.Database.GetGameReviewCommand(userId, partyId, gameId)

	if errors.Is(err, sql.ErrNoRows) {
		err = common.NewGameReviewNotFoundError(name)
	}

	return
}

// getPlayedGameId resolves a game name to its id through the games the user has finished, which are
// the only ones a rating can be attached to. A history entry without an end state is a game still in
// progress.
func (s *Service) getPlayedGameId(userId int, partyId int, name string) (gameId int, err error) {
	history, err := s.Database.GetGameHistoryCommand(userId, partyId)

	if err != nil {
		return
	}

	for _, entry := range history {
		if entry.Name == name && entry.EndState != nil {
			return entry.GameId, nil
		}
	}

	return 0, common.NewPlayedGameNotFoundError(name)
}

func (s *Service) StartCurrentGame(userId int, partyId int) (game typegames.CurrentGame, err error) {
	game, err = s.GettingService.GetCurrentGame(userId, partyId)

	var notFoundError *common.NotFoundError
	if err != nil && !errors.As(err, &notFoundError) {
		return
	}

	if game.Name != "" {
		err = common.NewCurrentGameAlreadyExistsConflictError()
		return
	}

	wishlistGames, err := s.GettingService.GetWishlistGames(userId, partyId)

	if err != nil {
		return
	}

	minimumNumberOfWishlistGames, err := s.SysParamsService.GetInt(partyId, typesysparams.ParamMinimumNumberOfWishlistGames)

	if err != nil {
		return
	}

	if wishlistGames == nil || len(wishlistGames) < minimumNumberOfWishlistGames {
		err = common.NewWishlistGamesNotFoundError(minimumNumberOfWishlistGames)
		return
	}

	randomNumber := rand.Intn(len(wishlistGames))
	randomWishlistGame := wishlistGames[randomNumber]

	createdGame, err := s.Database.CreateCurrentGameCommand(userId, partyId, randomWishlistGame.GameId, userId, nil)

	if err != nil {
		return
	}

	err = s.Database.DeleteWishlistGameCommand(userId, partyId, randomWishlistGame.GameId)

	if err != nil {
		return
	}

	game.Id = randomWishlistGame.GameId
	game.Name = randomWishlistGame.Name
	game.StartDate = createdGame.StartedDate

	return
}

func (s *Service) GetAllCurrentGames(partyId int) (games []typegames.CurrentGameWithLogin, err error) {
	userGames, err := s.Database.GetAllCurrentGamesCommand(partyId)

	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}

	if err != nil {
		return
	}

	games = make([]typegames.CurrentGameWithLogin, len(userGames))

	for i, userGame := range userGames {
		games[i] = typegames.CurrentGameWithLogin{
			Login: userGame.Login,
			Game: typegames.CurrentGame{
				Id:        userGame.Id,
				Name:      userGame.Name,
				TimeSpent: userGame.TimeSpent,
				StartDate: userGame.StartDate,
			},
		}
	}

	return
}
