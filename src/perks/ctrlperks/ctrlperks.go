package ctrlperks

import (
	"FGG-Service/api/generated/perks"
	"FGG-Service/src/auth/srvauth"
	"FGG-Service/src/common"
	"FGG-Service/src/parties/srvparties"
	"FGG-Service/src/perks/srvperks"
	"FGG-Service/src/perks/typeperks"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Controller struct {
	Service      srvperks.IService
	AuthService  srvauth.IService
	PartyService srvparties.IService
}

func NewController() *Controller {
	s := srvperks.NewService()
	as := srvauth.NewService()
	ps := srvparties.NewService()

	return &Controller{
		s,
		as,
		ps,
	}
}

// GetPerks (GET /parties/{partyId}/perks/catalog)
func (c *Controller) GetPerks(ctx echo.Context, partyId genperks.PartyId) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	perks, err := c.Service.GetPerks(ctx.Request().Context(), partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertPerksToDto(perks))
}

// GetRemovedPerks (GET /parties/{partyId}/perks/catalog/removed)
func (c *Controller) GetRemovedPerks(ctx echo.Context, partyId genperks.PartyId) error {
	err := common.RequireAdmin(ctx, c.AuthService, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	perks, err := c.Service.GetRemovedPerks(ctx.Request().Context(), partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertPerksToDto(perks))
}

// CreatePerk (POST /parties/{partyId}/perks/catalog)
func (c *Controller) CreatePerk(ctx echo.Context, partyId genperks.PartyId) error {
	err := common.RequireAdmin(ctx, c.AuthService, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var perkDto genperks.PerkCreate
	err = ctx.Bind(&perkDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	perk, err := c.Service.CreatePerk(ctx.Request().Context(), partyId, perkDto.Name, perkDto.Description, perkDto.EffectId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, convertPerkToDto(perk))
}

// RemovePerk (DELETE /parties/{partyId}/perks/catalog/{name})
func (c *Controller) RemovePerk(ctx echo.Context, partyId genperks.PartyId, name genperks.Name) error {
	err := common.RequireAdmin(ctx, c.AuthService, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.RemovePerk(ctx.Request().Context(), partyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetUserPerks (GET /parties/{partyId}/perks/{login})
func (c *Controller) GetUserPerks(ctx echo.Context, partyId genperks.PartyId, login genperks.Login) error {
	userId, err := c.userIdFromLogin(ctx, partyId, login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	userPerks, err := c.Service.GetUserPerks(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	userPerksDto := make(genperks.UserPerks, len(userPerks))
	for i, userPerk := range userPerks {
		userPerksDto[i] = genperks.UserPerk{
			Name:         userPerk.Name,
			Description:  userPerk.Description,
			ReceivedDate: userPerk.ReceivedDate,
		}
	}

	return ctx.JSON(http.StatusOK, userPerksDto)
}

// RevokeUserPerk (DELETE /parties/{partyId}/perks/{login}/{name})
func (c *Controller) RevokeUserPerk(ctx echo.Context, partyId genperks.PartyId, login genperks.Login, name genperks.Name) error {
	err := common.RequireAdmin(ctx, c.AuthService, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	userId, err := c.AuthService.GetUserIdByLogin(ctx.Request().Context(), login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.RevokeUserPerk(ctx.Request().Context(), actorUserId, userId, partyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetUserPerkHistory (GET /parties/{partyId}/perks/{login}/history)
func (c *Controller) GetUserPerkHistory(ctx echo.Context, partyId genperks.PartyId, login genperks.Login) error {
	userId, err := c.userIdFromLogin(ctx, partyId, login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	history, err := c.Service.GetPerkHistory(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	actorUserIds := make([]int, len(history))
	for i, entry := range history {
		actorUserIds[i] = entry.ActorUserId
	}

	loginsByUserId, err := common.GetLoginsByUserIds(ctx.Request().Context(), c.AuthService, actorUserIds)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	historyDto := make(genperks.PerkHistoryEntries, len(history))
	for i, entry := range history {
		historyDto[i] = genperks.PerkHistoryEntry{
			Name:        entry.Name,
			Action:      entry.Action,
			ActorLogin:  loginsByUserId[entry.ActorUserId],
			CreatedDate: entry.CreatedDate,
		}
	}

	return ctx.JSON(http.StatusOK, historyDto)
}

// userIdFromLogin resolves a login path parameter, rejecting the request when it carries no session
// or comes from a user outside the party.
func (c *Controller) userIdFromLogin(ctx echo.Context, partyId genperks.PartyId, login string) (userId int, err error) {
	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), actorUserId, partyId)

	if err != nil {
		return
	}

	return c.AuthService.GetUserIdByLogin(ctx.Request().Context(), login)
}

func convertPerkToDto(perk typeperks.Perk) genperks.Perk {
	return genperks.Perk{
		Name:        perk.Name,
		Description: perk.Description,
	}
}

func convertPerksToDto(perks []typeperks.Perk) genperks.Perks {
	perksDto := make(genperks.Perks, len(perks))

	for i, perk := range perks {
		perksDto[i] = convertPerkToDto(perk)
	}

	return perksDto
}
