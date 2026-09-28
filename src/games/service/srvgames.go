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

// defaultPartyId is a stopgap until real party context exists (see project plan) — every game is
// scoped to this one hardcoded party.
const defaultPartyId = 1

type IService interface {
	GetCurrentGame(userId int) (typegames.CurrentGame, error)
	CancelCurrentGame(userId int) error
	FinishCurrentGame(userId int) error
	MakeGameRoll(userId int) (typegames.CurrentGame, error)
	GetGameHistory(userId int) ([]typegames.GameHistoryEntry, error)
	RateGame(userId int, name string, rating int, reviewComment *string) error
	GetGameReview(userId int, name string) (typegames.GameReview, error)
	GetUnplayedGames(userId int) (typegames.WishlistGames, error)
	AddWishlistGame(userId int, wishlistGame typegames.WishlistGame) error
	GetAllCurrentGames() ([]typegames.CurrentGameWithLogin, error)
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
func (s *Service) AddWishlistGame(userId int, wishlistGame typegames.WishlistGame) error {
	game, err := s.Database.GetGameByNameCommand(defaultPartyId, wishlistGame.Name)

	if errors.Is(err, sql.ErrNoRows) {
		game, err = s.Database.CreateGameCommand(defaultPartyId, wishlistGame.Name)
	}

	if err != nil {
		return err
	}

	_, err = s.Database.GetWishlistGameCommand(userId, defaultPartyId, game.Id)

	// A row means the game is already on the wishlist; only its absence lets the insert through.
	if err == nil {
		return common.NewWishlistGameAlreadyExistsConflictError(wishlistGame.Name)
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	_, err = s.Database.CreateWishlistGameCommand(userId, defaultPartyId, game.Id)

	return err
}

func (s *Service) GetUnplayedGames(userId int) (typegames.WishlistGames, error) {
	return s.GettingService.GetWishlistGames(userId)
}

func (s *Service) GetCurrentGame(userId int) (typegames.CurrentGame, error) {
	return s.GettingService.GetCurrentGame(userId)
}

func (s *Service) CancelCurrentGame(userId int) error {
	game, err := s.GettingService.GetCurrentGame(userId)

	if err != nil {
		return err
	}

	_, err = s.TimerService.ForceStopCurrentTimer(userId)

	if err != nil {
		return err
	}

	err = s.Database.CancelCurrentGameCommand(userId, defaultPartyId, game.Id, userId, nil)

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) FinishCurrentGame(userId int) error {
	game, err := s.GettingService.GetCurrentGame(userId)

	if err != nil {
		return err
	}

	// The current timer's time is added to the game when it is stopped below, so it counts too.
	timerTimeSpent, err := s.TimerService.GetCurrentTimerTimeSpent(userId)

	if err != nil {
		return err
	}

	if game.TimeSpent+timerTimeSpent == 0 {
		return common.NewGameTimeSpentIsZeroError()
	}

	_, err = s.TimerService.ForceStopCurrentTimer(userId)

	if err != nil {
		return err
	}

	err = s.Database.FinishCurrentGameCommand(userId, defaultPartyId, game.Id, userId, nil)

	if err != nil {
		return err
	}

	return nil
}

// GetGameHistory returns the recorded game events of a user. A history entry is not a current game:
// it carries what happened to the game rather than how it stands now.
func (s *Service) GetGameHistory(userId int) (history []typegames.GameHistoryEntry, err error) {
	return s.Database.GetGameHistoryCommand(userId, defaultPartyId)
}

// RateGame records a rating and an optional review for a game the user has already played. Rating
// the same game again overwrites the previous rating.
func (s *Service) RateGame(userId int, name string, rating int, reviewComment *string) (err error) {
	gameId, err := s.getPlayedGameId(userId, name)

	if err != nil {
		return
	}

	_, err = s.Database.RateGameCommand(userId, defaultPartyId, gameId, rating, reviewComment)

	// The game is in the history but hasn't ended yet, so there is nothing to rate and the
	// database writes nothing.
	if errors.Is(err, sql.ErrNoRows) {
		err = common.NewPlayedGameNotFoundError(name)
	}

	return
}

// GetGameReview returns the rating and review the user left for a game they have played.
func (s *Service) GetGameReview(userId int, name string) (review typegames.GameReview, err error) {
	gameId, err := s.getPlayedGameId(userId, name)

	if err != nil {
		return
	}

	review, err = s.Database.GetGameReviewCommand(userId, defaultPartyId, gameId)

	if errors.Is(err, sql.ErrNoRows) {
		err = common.NewGameReviewNotFoundError(name)
	}

	return
}

// getPlayedGameId resolves a game name to its id through the games the user has finished, which are
// the only ones a rating can be attached to. A history entry without an end state is a game still in
// progress.
func (s *Service) getPlayedGameId(userId int, name string) (gameId int, err error) {
	history, err := s.Database.GetGameHistoryCommand(userId, defaultPartyId)

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

func (s *Service) MakeGameRoll(userId int) (game typegames.CurrentGame, err error) {
	game, err = s.GettingService.GetCurrentGame(userId)

	var notFoundError *common.NotFoundError
	if err != nil && !errors.As(err, &notFoundError) {
		return
	}

	if game.Name != "" {
		err = common.NewCurrentGameAlreadyExistsConflictError()
		return
	}

	unplayedGames, err := s.GettingService.GetWishlistGames(userId)

	if err != nil {
		return
	}

	minimumNumberOfWishlistGames, err := s.SysParamsService.GetInt(typesysparams.ParamMinimumNumberOfWishlistGames)

	if err != nil {
		return
	}

	if unplayedGames == nil || len(unplayedGames) < minimumNumberOfWishlistGames {
		err = common.NewUnplayedGamesNotFoundError(minimumNumberOfWishlistGames)
		return
	}

	randomNumber := rand.Intn(len(unplayedGames))
	randomUnplayedGame := unplayedGames[randomNumber]

	createdGame, err := s.Database.CreateCurrentGameCommand(userId, defaultPartyId, randomUnplayedGame.GameId, userId, nil)

	if err != nil {
		return
	}

	err = s.Database.DeleteUnplayedGameCommand(userId, defaultPartyId, randomUnplayedGame.GameId)

	if err != nil {
		return
	}

	game.Id = randomUnplayedGame.GameId
	game.Name = randomUnplayedGame.Name
	game.StartDate = createdGame.StartedDate

	return
}

func (s *Service) GetAllCurrentGames() (games []typegames.CurrentGameWithLogin, err error) {
	userGames, err := s.Database.GetAllCurrentGamesCommand(defaultPartyId)

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
