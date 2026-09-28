package ctrleffects

import (
	"FGG-Service/api/generated/effects"
	"FGG-Service/src/auth/service"
	"FGG-Service/src/changes/types"
	"FGG-Service/src/common"
	"FGG-Service/src/effects/service"
	"FGG-Service/src/effects/types"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// defaultPartyId is a stopgap until real party context exists (see project plan) — every effect is
// scoped to this one hardcoded party.
const defaultPartyId = 1

type Controller struct {
	Service     srveffects.IService
	AuthService srvauth.IService
}

func NewController() *Controller {
	s := srveffects.NewService()
	as := srvauth.NewService()

	return &Controller{
		s,
		as,
	}
}

// GetEffects (GET /effects/catalog)
func (c *Controller) GetEffects(ctx echo.Context) error {
	_, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	effects, err := c.Service.GetEffects(defaultPartyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertEffectsToDto(effects))
}

// GetRemovedEffects (GET /effects/catalog/removed)
func (c *Controller) GetRemovedEffects(ctx echo.Context) error {
	err := common.RequireAdmin(ctx, c.AuthService)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	effects, err := c.Service.GetRemovedEffects(defaultPartyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertEffectsToDto(effects))
}

// CreateEffect (POST /effects/catalog)
func (c *Controller) CreateEffect(ctx echo.Context) error {
	err := common.RequireAdmin(ctx, c.AuthService)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var effectDto geneffects.EffectCreate
	err = ctx.Bind(&effectDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	var duration *time.Duration
	if effectDto.DurationInSeconds != nil {
		effectDuration := time.Duration(*effectDto.DurationInSeconds) * time.Second
		duration = &effectDuration
	}

	effect, err := c.Service.CreateEffect(
		defaultPartyId,
		effectDto.Name,
		effectDto.Description,
		effectDto.UseCount,
		duration,
		convertDtoToChangeEntryInputs(effectDto.Entries))

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, convertEffectToDto(effect))
}

// RemoveEffect (DELETE /effects/catalog/{name})
func (c *Controller) RemoveEffect(ctx echo.Context, name geneffects.Name) error {
	err := common.RequireAdmin(ctx, c.AuthService)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.RemoveEffect(defaultPartyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}
// GetUserEffects (GET /effects/{login})
func (c *Controller) GetUserEffects(ctx echo.Context, login geneffects.Login) error {
	userId, err := c.userIdFromLogin(ctx, login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	views, err := c.Service.GetUserEffectViews(userId, defaultPartyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	viewsDto := make(geneffects.UserEffects, len(views))
	for i, view := range views {
		modifiersDto := make([]geneffects.PointModifier, len(view.Modifiers))

		for j, modifier := range view.Modifiers {
			modifiersDto[j] = geneffects.PointModifier{
				PointTypeName: modifier.PointTypeName,
				Amount:        modifier.Amount,
			}
		}

		viewsDto[i] = geneffects.UserEffect{
			Name:        view.Name,
			Description: view.Description,
			UseCount:    view.UseCount,
			UsesLeft:    view.UsesLeft,
			Duration:    convertDurationToDto(view.Duration),
			StartedDate: view.StartedDate,
			Modifiers:   modifiersDto,
		}
	}

	return ctx.JSON(http.StatusOK, viewsDto)
}

// UseUserEffect (POST /effects/{login}/{name}/use)
func (c *Controller) UseUserEffect(ctx echo.Context, login geneffects.Login, name geneffects.Name) error {
	userId, err := c.userIdFromLogin(ctx, login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.UseEffect(userId, defaultPartyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// EndUserEffect (DELETE /effects/{login}/{name})
func (c *Controller) EndUserEffect(ctx echo.Context, login geneffects.Login, name geneffects.Name) error {
	userId, err := c.userIdFromLogin(ctx, login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.EndUserEffect(userId, defaultPartyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetUserEffectHistory (GET /effects/{login}/history)
func (c *Controller) GetUserEffectHistory(ctx echo.Context, login geneffects.Login) error {
	userId, err := c.userIdFromLogin(ctx, login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	history, err := c.Service.GetEffectHistory(userId, defaultPartyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	historyDto := make(geneffects.EffectHistoryEntries, len(history))
	for i, entry := range history {
		historyDto[i] = geneffects.EffectHistoryEntry{
			Name:        entry.Name,
			Description: entry.Description,
			Action:      entry.Action,
			UsesLeft:    entry.UsesLeft,
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

func convertDurationToDto(duration *time.Duration) *geneffects.Duration {
	if duration == nil {
		return nil
	}

	converted := geneffects.Duration(common.DurationToISO8601(*duration))

	return &converted
}

func convertEffectToDto(effect typeeffects.Effect) geneffects.Effect {
	return geneffects.Effect{
		Name:        effect.Name,
		Description: effect.Description,
		UseCount:    effect.UseCount,
		Duration:    convertDurationToDto(effect.Duration),
	}
}

func convertEffectsToDto(effects []typeeffects.Effect) geneffects.Effects {
	effectsDto := make(geneffects.Effects, len(effects))

	for i, effect := range effects {
		effectsDto[i] = convertEffectToDto(effect)
	}

	return effectsDto
}

func convertDtoToChangeEntryInputs(entriesDto geneffects.ChangeEntryInputs) []typechanges.ChangeEntryInput {
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