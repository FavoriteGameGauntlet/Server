package ctrltimers

import (
	"FGG-Service/api/generated/timers"
	"FGG-Service/src/auth/srvauth"
	"FGG-Service/src/changes/typechanges"
	"FGG-Service/src/common"
	"FGG-Service/src/parties/srvparties"
	"FGG-Service/src/timers/srvtimers"
	"FGG-Service/src/timers/typetimers"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Controller struct {
	Service      srvtimers.Service
	AuthService  srvauth.Service
	PartyService srvparties.IService
}

func NewController(s *srvtimers.Service) *Controller {
	as := srvauth.NewService()
	ps := srvparties.NewService()

	return &Controller{
		*s,
		*as,
		ps,
	}
}

// GetCurrentTimer (GET /parties/{partyId}/timers/current)
func (c *Controller) GetCurrentTimer(ctx echo.Context, partyId gentimers.PartyId) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	timer, err := c.Service.GetCurrentTimer(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	timerDto := convertTimerToDto(timer)

	return ctx.JSON(http.StatusOK, timerDto)
}

// CreateCurrentTimer (POST /parties/{partyId}/timers/current)
func (c *Controller) CreateCurrentTimer(ctx echo.Context, partyId gentimers.PartyId) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	timer, err := c.Service.CreateCurrentTimer(ctx.Request().Context(), userId, partyId)

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

// PauseCurrentTimer (POST /parties/{partyId}/timers/current/pause)
func (c *Controller) PauseCurrentTimer(ctx echo.Context, partyId gentimers.PartyId) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	timer, err := c.Service.PauseCurrentTimer(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	timerActionDto := convertTimerToDto(timer)

	return ctx.JSON(http.StatusOK, timerActionDto)
}

// StartCurrentTimer (POST /parties/{partyId}/timers/current/start)
func (c *Controller) StartCurrentTimer(ctx echo.Context, partyId gentimers.PartyId) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	timer, err := c.Service.StartCurrentTimer(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	timerActionDto := convertTimerToDto(timer)

	return ctx.JSON(http.StatusOK, timerActionDto)
}

// GetTimerReward (GET /parties/{partyId}/timers/reward)
func (c *Controller) GetTimerReward(ctx echo.Context, partyId gentimers.PartyId) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	reward, err := c.Service.GetTimerReward(ctx.Request().Context(), partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertTimerRewardToDto(reward))
}

// SetTimerReward (PUT /parties/{partyId}/timers/reward)
func (c *Controller) SetTimerReward(ctx echo.Context, partyId gentimers.PartyId) error {
	err := common.RequireAdmin(ctx, &c.AuthService, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var rewardDto gentimers.TimerReward
	err = ctx.Bind(&rewardDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	reward, err := c.Service.SetTimerReward(ctx.Request().Context(), partyId, convertDtoToChangeEntryInputs(rewardDto.Entries))

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, convertTimerRewardToDto(reward))
}

// RemoveTimerReward (DELETE /parties/{partyId}/timers/reward)
func (c *Controller) RemoveTimerReward(ctx echo.Context, partyId gentimers.PartyId) error {
	err := common.RequireAdmin(ctx, &c.AuthService, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.RemoveTimerReward(ctx.Request().Context(), partyId)

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
