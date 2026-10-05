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

// GrantToUsers (POST /parties/{partyId}/grants)
func (c *Controller) GrantToUsers(ctx echo.Context, partyId gengrants.PartyId) error {
	err := common.RequireAdmin(ctx, c.AuthService, partyId)

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

	err = c.Service.GrantToUsers(actorUserId, partyId, targetUserIds, convertDtoToChangeEntryInputs(grantDto.Entries))

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