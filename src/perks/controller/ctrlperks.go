package ctrlperks

import (
	"FGG-Service/api/generated/perks"
	"FGG-Service/src/auth/service"
	"FGG-Service/src/common"
	"FGG-Service/src/parties/service"
	"FGG-Service/src/perks/service"
	"FGG-Service/src/perks/types"
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

	err = c.PartyService.RequireMember(userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	perks, err := c.Service.GetPerks(partyId)

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

	perks, err := c.Service.GetRemovedPerks(partyId)

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

	perk, err := c.Service.CreatePerk(partyId, perkDto.Name, perkDto.Description, perkDto.EffectName)

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

	err = c.Service.RemovePerk(partyId, name)

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

	views, err := c.Service.GetUserPerks(userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	viewsDto := make(genperks.UserPerks, len(views))
	for i, view := range views {
		viewsDto[i] = genperks.UserPerk{
			Name:         view.Name,
			Description:  view.Description,
			ReceivedDate: view.ReceivedDate,
		}
	}

	return ctx.JSON(http.StatusOK, viewsDto)
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

	userId, err := c.AuthService.GetUserIdByLogin(login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.RevokeUserPerk(actorUserId, userId, partyId, name)

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

	history, err := c.Service.GetPerkHistory(userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	actorUserIds := make([]int, len(history))
	for i, entry := range history {
		actorUserIds[i] = entry.ActorUserId
	}

	loginsByUserId, err := common.GetLoginsByUserIds(c.AuthService, actorUserIds)

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

	err = c.PartyService.RequireMember(actorUserId, partyId)

	if err != nil {
		return
	}

	return c.AuthService.GetUserIdByLogin(login)
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
