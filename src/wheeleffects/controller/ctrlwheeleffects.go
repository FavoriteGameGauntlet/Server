package ctrlwheeleffects

import (
	"FGG-Service/api/generated/games"
	"FGG-Service/api/generated/wheel_effects"
	"FGG-Service/src/auth/service"
	"FGG-Service/src/common"
	"FGG-Service/src/wheeleffects/service"
	"FGG-Service/src/wheeleffects/types"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Controller struct {
	Service     srvwheeleffects.Service
	AuthService srvauth.Service
}

func NewController() *Controller {
	s := srvwheeleffects.NewService()
	as := srvauth.NewService()

	return &Controller{
		*s,
		*as,
	}
}

// RollAvailableWheelEffects (POST /wheel-effects/available/roll)
func (c *Controller) RollAvailableWheelEffects(ctx echo.Context) error {
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

	isReroll := rollDto.IsReroll != nil && *rollDto.IsReroll

	rolled, err := c.Service.MakeEffectRoll(userId, isReroll)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertLastWheelRowsToDto(rolled))
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

// ApplyAvailableWheelEffectRoll (POST /wheel-effects/available/roll/apply)
func (c *Controller) ApplyAvailableWheelEffectRoll(ctx echo.Context) error {
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

	targetUserIds := make([]int, len(rollApplyDto.TargetLogins))
	for i, login := range rollApplyDto.TargetLogins {
		targetUserIds[i], err = c.AuthService.GetUserIdByLogin(login)

		if err != nil {
			return common.SendJSONErrorResponse(ctx, err)
		}
	}

	err = c.Service.ApplyWheelEffectRoll(userId, rollApplyDto.WheelRowName, targetUserIds)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetLastRolledWheelEffects (GET /wheel-effects/available/roll/last)
func (c *Controller) GetLastRolledWheelEffects(ctx echo.Context) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	rows, err := c.Service.GetLastRolledWheelEffects(userId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertLastWheelRowsToDto(rows))
}

// ClearLastRolledWheelEffects (POST /wheel-effects/available/roll/last/clear)
func (c *Controller) ClearLastRolledWheelEffects(ctx echo.Context) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.ClearLastWheelEffects(userId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetAvailableWheelEffectRollsCount (GET /wheel-effects/available/roll/count)
func (c *Controller) GetAvailableWheelEffectRollsCount(ctx echo.Context) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	count, err := c.Service.GetAvailableRollsCount(userId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, count)
}

// GetAvailableWheelEffects (GET /wheel-effects/available)
func (c *Controller) GetAvailableWheelEffects(ctx echo.Context) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	rows, err := c.Service.GetAvailableWheelRows(userId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertWheelRowsToDto(rows))
}

// GetUserWheelEffectHistory (GET /wheel-effects/{login}/history)
func (c *Controller) GetUserWheelEffectHistory(ctx echo.Context, login gengames.Login) error {
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

	history, err := c.Service.GetEffectHistory(userId)

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
