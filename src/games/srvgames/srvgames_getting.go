package srvgames

import (
	"FGG-Service/src/common"
	"FGG-Service/src/games/dbgames"
	"FGG-Service/src/games/typegames"
	"context"
	"database/sql"
	"errors"
)

type IGettingService interface {
	GetWishlistGames(ctx context.Context, userId int, partyId int) (typegames.WishlistGames, error)
	GetCurrentGame(ctx context.Context, userId int, partyId int) (typegames.CurrentGame, error)
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

func (s *GettingService) GetCurrentGame(ctx context.Context, userId int, partyId int) (game typegames.CurrentGame, err error) {
	userGame, err := s.Database.GetCurrentGameCommand(ctx, userId, partyId)

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

func (s *GettingService) GetWishlistGames(ctx context.Context, userId int, partyId int) (typegames.WishlistGames, error) {
	return s.Database.GetWishlistGamesCommand(ctx, userId, partyId)
}
