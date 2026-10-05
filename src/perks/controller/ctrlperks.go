package ctrlperks

import (
	"FGG-Service/api/generated/perks"
	"FGG-Service/src/auth/service"
	"FGG-Service/src/common"
	"FGG-Service/src/perks/service"
	"FGG-Service/src/perks/types"
	"net/http"

	"github.com/labstack/echo/v4"
)

// defaultPartyId is a stopgap until real party context exists (see project plan) — every perk is
// scoped to this one hardcoded party.
const defaultPartyId = 1

type Controller struct {
	Service     srvperks.IService
	AuthService srvauth.IService
}

func NewController() *Controller {
	s := srvperks.NewService()
	as := srvauth.NewService()

	return &Controller{
		s,
		as,
	}
}

// GetPerks (GET /perks/catalog)
func (c *Controller) GetPerks(ctx echo.Context) error {
	_, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	perks, err := c.Service.GetPerks(defaultPartyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertPerksToDto(perks))
}

// GetRemovedPerks (GET /perks/catalog/removed)
func (c *Controller) GetRemovedPerks(ctx echo.Context) error {
	err := common.RequireAdmin(ctx, c.AuthService)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	perks, err := c.Service.GetRemovedPerks(defaultPartyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertPerksToDto(perks))
}

// CreatePerk (POST /perks/catalog)
func (c *Controller) CreatePerk(ctx echo.Context) error {
	err := common.RequireAdmin(ctx, c.AuthService)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var perkDto genperks.PerkCreate
	err = ctx.Bind(&perkDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	perk, err := c.Service.CreatePerk(defaultPartyId, perkDto.Name, perkDto.Description, perkDto.EffectName)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, convertPerkToDto(perk))
}

// RemovePerk (DELETE /perks/catalog/{name})
func (c *Controller) RemovePerk(ctx echo.Context, name genperks.Name) error {
	err := common.RequireAdmin(ctx, c.AuthService)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.RemovePerk(defaultPartyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetUserPerks (GET /perks/{login})
func (c *Controller) GetUserPerks(ctx echo.Context, login genperks.Login) error {
	userId, err := c.userIdFromLogin(ctx, login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	views, err := c.Service.GetUserPerks(userId, defaultPartyId)

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

// RevokeUserPerk (DELETE /perks/{login}/{name})
func (c *Controller) RevokeUserPerk(ctx echo.Context, login genperks.Login, name genperks.Name) error {
	err := common.RequireAdmin(ctx, c.AuthService)

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

	err = c.Service.RevokeUserPerk(actorUserId, userId, defaultPartyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetUserPerkHistory (GET /perks/{login}/history)
func (c *Controller) GetUserPerkHistory(ctx echo.Context, login genperks.Login) error {
	userId, err := c.userIdFromLogin(ctx, login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	history, err := c.Service.GetPerkHistory(userId, defaultPartyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	historyDto := make(genperks.PerkHistoryEntries, len(history))
	for i, entry := range history {
		historyDto[i] = genperks.PerkHistoryEntry{
			Name:        entry.Name,
			Action:      entry.Action,
			CreatedDate: entry.CreatedDate,
		}
	}

	return ctx.JSON(http.StatusOK, historyDto)
}

// userIdFromLogin resolves a login path parameter, rejecting the request when it carries no session.
func (c *Controller) userIdFromLogin(ctx echo.Context, login string) (userId int, err error) {
	_, err = c.AuthService.GetUserId(ctx)

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