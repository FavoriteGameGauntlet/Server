package ctrlgames

import (
	"FGG-Service/api/generated/games"
	"FGG-Service/src/auth/service"
	"FGG-Service/src/common"
	"FGG-Service/src/games/service"
	"FGG-Service/src/games/types"
	"FGG-Service/src/timers/service"
	"FGG-Service/src/validator"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Controller struct {
	Service     srvgames.IService
	AuthService srvauth.IService
}

func NewController(ts srvtimers.IService) *Controller {
	s := srvgames.NewService(ts)
	as := srvauth.NewService()

	return &Controller{
		s,
		as,
	}
}

// GetUserCurrentGame (GET /games/{login}/current)
func (c *Controller) GetUserCurrentGame(ctx echo.Context, login gengames.Login) error {
	doesExist, err := c.AuthService.DoesUserSessionExist(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	if !doesExist {
		err = common.NewActiveSessionNotFoundUnauthorizedError()
		return common.SendJSONErrorResponse(ctx, err)
	}

	userId, err := c.AuthService.GetUserIdByLogin(login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	game, err := c.Service.GetCurrentGame(userId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	gameDto := convertGameToDto(game)

	return ctx.JSON(http.StatusOK, gameDto)
}

func convertGameToDto(game typegames.CurrentGame) gengames.CurrentGame {
	return gengames.CurrentGame{
		Name:      game.Name,
		State:     gengames.CurrentGameState(game.State),
		TimeSpent: common.DurationToISO8601(game.TimeSpent),
		StartDate: game.StartDate,
	}
}

// CancelCurrentGame (POST /games/current/cancel)
func (c *Controller) CancelCurrentGame(ctx echo.Context) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.CancelCurrentGame(userId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// FinishCurrentGame (POST /games/current/finish)
func (c *Controller) FinishCurrentGame(ctx echo.Context) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.FinishCurrentGame(userId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// RollNewCurrentGame (POST /games/current/roll)
func (c *Controller) RollNewCurrentGame(ctx echo.Context) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	game, err := c.Service.MakeGameRoll(userId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	gameDto := convertGameToDto(game)

	return ctx.JSON(http.StatusOK, gameDto)
}

// GetUserGameHistory (GET /games/{login}/history)
func (c *Controller) GetUserGameHistory(ctx echo.Context, login gengames.Login) error {
	doesExist, err := c.AuthService.DoesUserSessionExist(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	if !doesExist {
		err = common.NewActiveSessionNotFoundUnauthorizedError()
		return common.SendJSONErrorResponse(ctx, err)
	}

	userId, err := c.AuthService.GetUserIdByLogin(login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	history, err := c.Service.GetGameHistory(userId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertGameHistoryToDto(history))
}

func convertGameHistoryToDto(history []typegames.GameHistoryEntry) gengames.GameHistoryEntries {
	historyDto := make(gengames.GameHistoryEntries, len(history))

	for i, entry := range history {
		historyDto[i] = gengames.GameHistoryEntry{
			Name:          entry.Name,
			Action:        entry.Action,
			TimeSpent:     common.DurationToISO8601(entry.TimeSpent),
			Rating:        entry.Rating,
			ReviewComment: entry.ReviewComment,
			CreatedDate:   entry.CreatedDate,
		}

		if entry.EndState != nil {
			endState := gengames.GameHistoryEntryEndState(*entry.EndState)
			historyDto[i].EndState = &endState
		}
	}

	return historyDto
}

// GetUserWishlistGames (GET /games/{login}/wishlist)
func (c *Controller) GetUserWishlistGames(ctx echo.Context, login gengames.Login) error {
	doesExist, err := c.AuthService.DoesUserSessionExist(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	if !doesExist {
		err = common.NewActiveSessionNotFoundUnauthorizedError()
		return common.SendJSONErrorResponse(ctx, err)
	}

	userId, err := c.AuthService.GetUserIdByLogin(login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	games, err := c.Service.GetUnplayedGames(userId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	gamesDto := convertWishlistGamesToDto(games)

	return ctx.JSON(http.StatusOK, gamesDto)
}

func convertWishlistGamesToDto(games typegames.WishlistGames) gengames.WishlistGames {
	gamesDto := make(gengames.WishlistGames, len(games))

	for i, game := range games {
		gamesDto[i] = gengames.WishlistGame{
			Name: game.Name,
		}
	}

	return gamesDto
}

// AddUserWishlistGame (POST /games/{login}/wishlist)
func (c *Controller) AddUserWishlistGame(ctx echo.Context, login gengames.Login) error {
	doesExist, err := c.AuthService.DoesUserSessionExist(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	if !doesExist {
		err = common.NewActiveSessionNotFoundUnauthorizedError()
		return common.SendJSONErrorResponse(ctx, err)
	}

	userId, err := c.AuthService.GetUserIdByLogin(login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var gameDto gengames.WishlistGame
	err = ctx.Bind(&gameDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	game := convertWishlistGameFromDto(gameDto)

	err = validator.ValidateName(game.Name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.AddWishlistGame(userId, game)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

func convertWishlistGameFromDto(gameDto gengames.WishlistGame) typegames.WishlistGame {
	return typegames.WishlistGame{
		Name: gameDto.Name,
	}
}

// GetAllCurrentGame (GET /games/all/current)
func (c *Controller) GetAllCurrentGame(ctx echo.Context) error {
	games, err := c.Service.GetAllCurrentGames()

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	gamesDto := convertAllCurrentGamesToDto(games)

	return ctx.JSON(http.StatusOK, gamesDto)
}

func convertAllCurrentGamesToDto(games []typegames.CurrentGameWithLogin) gengames.CurrentGameByLogins {
	dtos := make(gengames.CurrentGameByLogins, len(games))

	for i, g := range games {
		gameDto := convertGameToDto(g.Game)
		dtos[i].CurrentGame = &gameDto
		dtos[i].Login = g.Login
	}

	return dtos
}

// RateCurrentGame (POST /games/current/rate)
func (c *Controller) RateCurrentGame(ctx echo.Context) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var rateDto gengames.GameRate
	err = ctx.Bind(&rateDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.RateCurrentGame(userId, rateDto.Rating, rateDto.ReviewComment)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}