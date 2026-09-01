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

// defaultPartyId and sourceEventIdPlaceholder are stopgaps until real party-context resolution and
// game-history-event wiring exist (see project plan) — game endpoints are not functionally correct
// against the new schema yet, this only gets them compiling.
const (
	defaultPartyId         = 1
	sourceEventIdPlaceholder = 0
)

type IService interface {
	GetCurrentGame(userId int) (typegames.CurrentGame, error)
	CancelCurrentGame(userId int) error
	FinishCurrentGame(userId int) error
	MakeGameRoll(userId int) (typegames.CurrentGame, error)
	GetGameHistory(userId int) (typegames.CurrentGames, error)
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

func (s *Service) AddWishlistGame(userId int, wishlistGame typegames.WishlistGame) error {
	doesGameExist, err := s.Database.DoesGameExistCommand(defaultPartyId, wishlistGame.Name)

	if err != nil {
		return err
	}

	game := typegames.WishlistGame{}

	if doesGameExist {
		game, err = s.Database.GetWishlistGameCommand(wishlistGame.Name)
	} else {
		game, err = s.createAndGetGame(wishlistGame)
	}

	if err != nil {
		return err
	}

	doesWishlistGameExist, err := s.Database.DoesWishlistGameExistCommand(userId, defaultPartyId, game.GameId)

	if err != nil {
		return err
	}

	if doesWishlistGameExist {
		return common.NewWishlistGameAlreadyExistsConflictError(wishlistGame.Name)
	}

	_, err = s.Database.CreateWishlistGameCommand(userId, defaultPartyId, game.GameId)

	return err
}

func (s *Service) GetUnplayedGames(userId int) (typegames.WishlistGames, error) {
	return s.GettingService.GetWishlistGames(userId)
}

func (s *Service) createAndGetGame(wishlistGame typegames.WishlistGame) (game typegames.WishlistGame, err error) {
	_, err = s.Database.CreateGameCommand(defaultPartyId, wishlistGame.Name)

	if err != nil {
		return
	}

	game, err = s.Database.GetWishlistGameCommand(wishlistGame.Name)

	return
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

	err = s.Database.CancelCurrentGameCommand(userId, defaultPartyId, game.Id, sourceEventIdPlaceholder)

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

	if game.TimeSpent == 0 {
		return common.NewCompletedTimersNotFoundError()
	}

	_, err = s.TimerService.ForceStopCurrentTimer(userId)

	if err != nil {
		return err
	}

	err = s.Database.FinishCurrentGameCommand(userId, defaultPartyId, game.Id, sourceEventIdPlaceholder)

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetGameHistory(userId int) (games typegames.CurrentGames, err error) {
	entries, err := s.Database.GetGameHistoryCommand(userId, defaultPartyId)

	if err != nil {
		return
	}

	games = make(typegames.CurrentGames, len(entries))

	for i, entry := range entries {
		game := typegames.CurrentGame{
			Id:        entry.GameId,
			Name:      entry.Name,
			TimeSpent: entry.TimeSpent,
			StartDate: entry.CreatedDate,
		}

		if entry.EndState != nil {
			game.State = typegames.CurrentGameState(*entry.EndState)
		}

		games[i] = game
	}

	return
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

	_, err = s.Database.CreateCurrentGameCommand(userId, defaultPartyId, randomUnplayedGame.GameId, sourceEventIdPlaceholder)

	if err != nil {
		return
	}

	err = s.Database.DeleteUnplayedGameCommand(userId, defaultPartyId, randomUnplayedGame.GameId)

	if err != nil {
		return
	}

	game.Id = randomUnplayedGame.GameId
	game.Name = randomUnplayedGame.Name
	game.State = typegames.GameStateStarted

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
			},
		}
	}

	return
}
