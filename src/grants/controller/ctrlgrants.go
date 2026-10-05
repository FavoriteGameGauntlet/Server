package ctrlgrants

import (
	"FGG-Service/api/generated/grants"
	"FGG-Service/src/auth/service"
	"FGG-Service/src/changes/types"
	"FGG-Service/src/common"
	"FGG-Service/src/grants/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

// defaultPartyId is a stopgap until real party context exists (see project plan) — every grant is
// scoped to this one hardcoded party.
const defaultPartyId = 1

type Controller struct {
	Service     srvgrants.IService
	AuthService srvauth.IService
}

func NewController() *Controller {
	s := srvgrants.NewService()
	as := srvauth.NewService()

	return &Controller{
		s,
		as,
	}
}

// GrantToUsers (POST /grants)
func (c *Controller) GrantToUsers(ctx echo.Context) error {
	err := common.RequireAdmin(ctx, c.AuthService)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var grantDto gengrants.ManualGrant
	err = ctx.Bind(&grantDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	targetUserIds := make([]int, len(grantDto.TargetLogins))
	for i, login := range grantDto.TargetLogins {
		targetUserIds[i], err = c.AuthService.GetUserIdByLogin(login)

		if err != nil {
			return common.SendJSONErrorResponse(ctx, err)
		}
	}

	err = c.Service.GrantToUsers(actorUserId, defaultPartyId, targetUserIds, convertDtoToChangeEntryInputs(grantDto.Entries))

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

func convertDtoToChangeEntryInputs(entriesDto gengrants.ChangeEntryInputs) []typechanges.ChangeEntryInput {
	inputs := make([]typechanges.ChangeEntryInput, len(entriesDto))

	for i, entryDto := range entriesDto {
		inputs[i] = typechanges.ChangeEntryInput{
			PointTypeName: entryDto.PointTypeName,
			ItemName:      entryDto.ItemName,
			PerkName:      entryDto.PerkName,
			EffectName:    entryDto.EffectName,
			Amount:        entryDto.Amount,
		}
	}

	return inputs
}