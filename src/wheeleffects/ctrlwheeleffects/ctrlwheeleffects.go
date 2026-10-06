package ctrlwheeleffects

import (
	"FGG-Service/api/generated/wheel_effects"
	"FGG-Service/src/auth/srvauth"
	"FGG-Service/src/common"
	"FGG-Service/src/parties/srvparties"
	"FGG-Service/src/wheeleffects/srvwheeleffects"
	"FGG-Service/src/wheeleffects/typewheeleffects"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Controller struct {
	Service      srvwheeleffects.Service
	AuthService  srvauth.Service
	PartyService srvparties.IService
}

func NewController() *Controller {
	s := srvwheeleffects.NewService()
	as := srvauth.NewService()
	ps := srvparties.NewService()

	return &Controller{
		*s,
		*as,
		ps,
	}
}

// RollAvailableWheelEffects (POST /parties/{partyId}/wheel-effects/available/roll)
func (c *Controller) RollAvailableWheelEffects(ctx echo.Context, partyId genwheeleffects.PartyId) error {
	var rollDto genwheeleffects.WheelEffectRoll
	err := ctx.Bind(&rollDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	isReroll := rollDto.IsReroll != nil && *rollDto.IsReroll

	rolled, err := c.Service.MakeEffectRoll(ctx.Request().Context(), userId, partyId, isReroll)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, convertLastWheelRowsToDto(rolled))
}

func convertWheelRowsToDto(rows []typewheeleffects.WheelRow) genwheeleffects.WheelRows {
	rowsDto := make(genwheeleffects.WheelRows, len(rows))

	for i, row := range rows {
		rowsDto[i] = genwheeleffects.WheelRow{
			Name:        row.Name,
			Description: &row.Description,
		}
	}

	return rowsDto
}

func convertLastWheelRowsToDto(rows []typewheeleffects.LastWheelRow) genwheeleffects.RolledWheelRows {
	rowsDto := make(genwheeleffects.RolledWheelRows, len(rows))

	for i, row := range rows {
		rowsDto[i] = genwheeleffects.RolledWheelRow{
			Name:        row.Name,
			Description: &row.Description,
			Position:    row.WheelPosition,
			RolledDate:  row.RolledDate,
		}
	}

	return rowsDto
}

// ApplyAvailableWheelEffectRoll (POST /parties/{partyId}/wheel-effects/available/roll/apply)
func (c *Controller) ApplyAvailableWheelEffectRoll(ctx echo.Context, partyId genwheeleffects.PartyId) error {
	var rollApplyDto genwheeleffects.WheelRowApply
	err := ctx.Bind(&rollApplyDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	targetUserIds := make([]int, len(rollApplyDto.TargetLogins))
	for i, login := range rollApplyDto.TargetLogins {
		targetUserIds[i], err = c.AuthService.GetUserIdByLogin(ctx.Request().Context(), login)

		if err != nil {
			return common.SendJSONErrorResponse(ctx, err)
		}
	}

	err = c.Service.ApplyWheelEffectRoll(ctx.Request().Context(), userId, partyId, rollApplyDto.WheelRowName, targetUserIds)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetLastRolledWheelEffects (GET /parties/{partyId}/wheel-effects/available/roll/last)
func (c *Controller) GetLastRolledWheelEffects(ctx echo.Context, partyId genwheeleffects.PartyId) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	rows, err := c.Service.GetLastRolledWheelEffects(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertLastWheelRowsToDto(rows))
}

// ClearLastRolledWheelEffects (POST /parties/{partyId}/wheel-effects/available/roll/last/clear)
func (c *Controller) ClearLastRolledWheelEffects(ctx echo.Context, partyId genwheeleffects.PartyId) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.ClearLastWheelEffects(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetAvailableWheelEffects (GET /parties/{partyId}/wheel-effects/available)
func (c *Controller) GetAvailableWheelEffects(ctx echo.Context, partyId genwheeleffects.PartyId) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	rows, err := c.Service.GetAvailableWheelRows(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertWheelRowsToDto(rows))
}

// GetUserWheelEffectHistory (GET /parties/{partyId}/wheel-effects/{login}/history)
func (c *Controller) GetUserWheelEffectHistory(ctx echo.Context, partyId genwheeleffects.PartyId, login genwheeleffects.Login) error {
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

	history, err := c.Service.GetEffectHistory(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	historyDto := make(genwheeleffects.WheelRowHistories, len(history))
	for i, entry := range history {
		historyDto[i] = genwheeleffects.WheelRowHistory{
			Name:        entry.Name,
			Description: &entry.Description,
			AppliedDate: entry.AppliedDate,
		}
	}

	return ctx.JSON(http.StatusOK, historyDto)
}
