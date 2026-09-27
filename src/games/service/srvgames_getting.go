package srvgames

import (
	"FGG-Service/src/common"
	"FGG-Service/src/games/database"
	"FGG-Service/src/games/types"
	"database/sql"
	"errors"
)

type IGettingService interface {
	GetWishlistGames(userId int) (typegames.WishlistGames, error)
	GetCurrentGame(userId int) (typegames.CurrentGame, error)
}

type GettingService struct {
	Database dbgames.IDatabase
}

func NewGettingService() IGettingService {
	db := new(dbgames.Database)

	return &GettingService{
		Database: db,
	}
}

func (s *GettingService) GetCurrentGame(userId int) (game typegames.CurrentGame, err error) {
	userGame, err := s.Database.GetCurrentGameCommand(userId, defaultPartyId)

	if errors.Is(err, sql.ErrNoRows) {
		err = common.NewCurrentGameNotFoundError()
		return
	}

	if err != nil {
		return
	}

	game = typegames.CurrentGame{
		Id:        userGame.Id,
		Name:      userGame.Name,
		TimeSpent: userGame.TimeSpent,
		StartDate: userGame.StartDate,
	}

	return
}

func (s *GettingService) GetWishlistGames(userId int) (typegames.WishlistGames, error) {
	return s.Database.GetWishlistGamesCommand(userId, defaultPartyId)
}
