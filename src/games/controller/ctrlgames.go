package ctrlgames

import (
	"FGG-Service/api/generated/games"
	"FGG-Service/src/auth/service"
	"FGG-Service/src/common"
	"FGG-Service/src/games/service"
	"FGG-Service/src/games/types"
	"FGG-Service/src/parties/service"
	"FGG-Service/src/timers/service"
	"FGG-Service/src/validator"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Controller struct {
	Service      srvgames.IService
	AuthService  srvauth.IService
	PartyService srvparties.IService
}

func NewController(ts srvtimers.IService) *Controller {
	s := srvgames.NewService(ts)
	as := srvauth.NewService()
	ps := srvparties.NewService()

	return &Controller{
		s,
		as,
		ps,
	}
}

// GetUserCurrentGame (GET /parties/{partyId}/games/users/{login}/current)
func (c *Controller) GetUserCurrentGame(ctx echo.Context, partyId gengames.PartyId, login gengames.Login) error {
	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), actorUserId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	userId, err := c.AuthService.GetUserIdByLogin(ctx.Request().Context(), login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	game, err := c.Service.GetCurrentGame(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	gameDto := convertGameToDto(game)

	return ctx.JSON(http.StatusOK, gameDto)
}

func convertGameToDto(game typegames.CurrentGame) gengames.CurrentGame {
	return gengames.CurrentGame{
		Name:      game.Name,
		TimeSpent: common.DurationToISO8601(game.TimeSpent),
		StartDate: game.StartDate,
	}
}

// CreateCurrentGameAction (POST /parties/{partyId}/games/current/actions)
func (c *Controller) CreateCurrentGameAction(ctx echo.Context, partyId gengames.PartyId) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var actionDto gengames.CurrentGameAction
	err = ctx.Bind(&actionDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	var game typegames.CurrentGame

	switch actionDto.Action {
	case gengames.Start:
		game, err = c.Service.StartCurrentGame(ctx.Request().Context(), userId, partyId)
	case gengames.Finish:
		game, err = c.Service.FinishCurrentGame(ctx.Request().Context(), userId, partyId)
	case gengames.Cancel:
		game, err = c.Service.CancelCurrentGame(ctx.Request().Context(), userId, partyId)
	default:
		err = common.NewCurrentGameActionUnprocessableError(string(actionDto.Action))
	}

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertGameToDto(game))
}

// GetUserGameHistory (GET /parties/{partyId}/games/users/{login}/history)
func (c *Controller) GetUserGameHistory(ctx echo.Context, partyId gengames.PartyId, login gengames.Login) error {
	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), actorUserId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	userId, err := c.AuthService.GetUserIdByLogin(ctx.Request().Context(), login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	history, err := c.Service.GetGameHistory(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertGameHistoryToDto(history))
}

func convertGameHistoryToDto(history []typegames.GameHistoryEntry) gengames.GameHistoryEntries {
	historyDto := make(gengames.GameHistoryEntries, len(history))

	for i, entry := range history {
		historyDto[i] = gengames.GameHistoryEntry{
			Name:        entry.Name,
			Action:      entry.Action,
			TimeSpent:   common.DurationToISO8601(entry.TimeSpent),
			CreatedDate: entry.CreatedDate,
		}

		if entry.EndState != nil {
			endState := gengames.GameHistoryEntryEndState(*entry.EndState)
			historyDto[i].EndState = &endState
		}
	}

	return historyDto
}

// GetUserWishlistGames (GET /parties/{partyId}/games/users/{login}/wishlist)
func (c *Controller) GetUserWishlistGames(ctx echo.Context, partyId gengames.PartyId, login gengames.Login) error {
	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), actorUserId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	userId, err := c.AuthService.GetUserIdByLogin(ctx.Request().Context(), login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	games, err := c.Service.GetWishlistGames(ctx.Request().Context(), userId, partyId)

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

// AddUserWishlistGame (POST /parties/{partyId}/games/users/{login}/wishlist)
func (c *Controller) AddUserWishlistGame(ctx echo.Context, partyId gengames.PartyId, login gengames.Login) error {
	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), actorUserId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	userId, err := c.AuthService.GetUserIdByLogin(ctx.Request().Context(), login)

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

	err = c.Service.AddWishlistGame(ctx.Request().Context(), userId, partyId, game)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusCreated)
}

func convertWishlistGameFromDto(gameDto gengames.WishlistGame) typegames.WishlistGame {
	return typegames.WishlistGame{
		Name: gameDto.Name,
	}
}

// GetAllCurrentGame (GET /parties/{partyId}/games/current)
func (c *Controller) GetAllCurrentGame(ctx echo.Context, partyId gengames.PartyId) error {
	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), actorUserId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	games, err := c.Service.GetAllCurrentGames(ctx.Request().Context(), partyId)

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

// PutGameReview (PUT /parties/{partyId}/games/reviews/{name})
func (c *Controller) PutGameReview(ctx echo.Context, partyId gengames.PartyId, name gengames.Name) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var reviewDto gengames.GameReview
	err = ctx.Bind(&reviewDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = validator.ValidateName(name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = validator.ValidateRating(reviewDto.Rating)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	created, err := c.Service.RateGame(ctx.Request().Context(), userId, partyId, name, reviewDto.Rating, reviewDto.ReviewComment)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	if created {
		return ctx.NoContent(http.StatusCreated)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetGameReview (GET /parties/{partyId}/games/reviews/{name})
func (c *Controller) GetGameReview(ctx echo.Context, partyId gengames.PartyId, name gengames.Name) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = validator.ValidateName(name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	review, err := c.Service.GetGameReview(ctx.Request().Context(), userId, partyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertGameReviewToDto(review))
}

func convertGameReviewToDto(review typegames.GameReview) gengames.GameReview {
	return gengames.GameReview{
		Rating:        review.Rating,
		ReviewComment: review.ReviewComment,
	}
}
