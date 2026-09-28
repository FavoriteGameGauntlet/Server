package ctrltimers

import (
	"FGG-Service/api/generated/timers"
	"FGG-Service/src/auth/service"
	"FGG-Service/src/changes/types"
	"FGG-Service/src/common"
	"FGG-Service/src/timers/service"
	"FGG-Service/src/timers/types"
	"net/http"

	"github.com/labstack/echo/v4"
)

// defaultPartyId is a stopgap until real party context exists (see project plan) — the timer reward
// is scoped to this one hardcoded party.
const defaultPartyId = 1

type Controller struct {
	Service     srvtimers.Service
	AuthService srvauth.Service
}

func NewController(s *srvtimers.Service) *Controller {
	as := srvauth.NewService()

	return &Controller{
		*s,
		*as,
	}
}

// GetCurrentTimer (GET /timers/current)
func (c *Controller) GetCurrentTimer(ctx echo.Context) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	timer, err := c.Service.GetCurrentTimer(userId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	timerDto := convertTimerToDto(timer)

	return ctx.JSON(http.StatusOK, timerDto)
}

// CreateCurrentTimer (POST /timers/current)
func (c *Controller) CreateCurrentTimer(ctx echo.Context) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	timer, err := c.Service.CreateCurrentTimer(userId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, convertTimerToDto(timer))
}

func convertTimerToDto(timer typetimers.Timer) gentimers.Timer {
	return gentimers.Timer{
		Duration:       common.DurationToISO8601(timer.Duration),
		RemainingTime:  common.DurationToISO8601(timer.RemainingTime),
		State:          gentimers.TimerState(timer.State),
		LastActionDate: timer.LastActionDate,
	}
}

// PauseCurrentTimer (POST /timers/current/pause)
func (c *Controller) PauseCurrentTimer(ctx echo.Context) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	timer, err := c.Service.PauseCurrentTimer(userId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	timerActionDto := convertTimerToDto(timer)

	return ctx.JSON(http.StatusOK, timerActionDto)
}

// StartCurrentTimer (POST /timers/current/start)
func (c *Controller) StartCurrentTimer(ctx echo.Context) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	timer, err := c.Service.StartCurrentTimer(userId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	timerActionDto := convertTimerToDto(timer)

	return ctx.JSON(http.StatusOK, timerActionDto)
}

// GetTimerReward (GET /timers/reward)
func (c *Controller) GetTimerReward(ctx echo.Context) error {
	_, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	reward, err := c.Service.GetTimerReward(defaultPartyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertTimerRewardToDto(reward))
}

// SetTimerReward (PUT /timers/reward)
func (c *Controller) SetTimerReward(ctx echo.Context) error {
	err := common.RequireAdmin(ctx, &c.AuthService)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var rewardDto gentimers.TimerReward
	err = ctx.Bind(&rewardDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	reward, err := c.Service.SetTimerReward(defaultPartyId, convertDtoToChangeEntryInputs(rewardDto.Entries))

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, convertTimerRewardToDto(reward))
}

// RemoveTimerReward (DELETE /timers/reward)
func (c *Controller) RemoveTimerReward(ctx echo.Context) error {
	err := common.RequireAdmin(ctx, &c.AuthService)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.RemoveTimerReward(defaultPartyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

func convertTimerRewardToDto(entries []typechanges.ChangeEntryInput) gentimers.TimerReward {
	entriesDto := make(gentimers.ChangeEntryInputs, len(entries))

	for i, entry := range entries {
		entriesDto[i] = gentimers.ChangeEntryInput{
			PointTypeName: entry.PointTypeName,
			ItemName:      entry.ItemName,
			PerkName:      entry.PerkName,
			EffectName:    entry.EffectName,
			Amount:        entry.Amount,
		}
	}

	return gentimers.TimerReward{Entries: entriesDto}
}

func convertDtoToChangeEntryInputs(entriesDto gentimers.ChangeEntryInputs) []typechanges.ChangeEntryInput {
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
