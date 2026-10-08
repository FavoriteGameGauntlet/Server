package ctrleffects

import (
	"FGG-Service/api/generated/effects"
	"FGG-Service/src/auth/srvauth"
	"FGG-Service/src/changes/typechanges"
	"FGG-Service/src/common"
	"FGG-Service/src/effects/srveffects"
	"FGG-Service/src/effects/typeeffects"
	"FGG-Service/src/parties/srvparties"
	"FGG-Service/src/validator"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type Controller struct {
	Service      srveffects.IService
	AuthService  srvauth.IService
	PartyService srvparties.IService
}

func NewController() *Controller {
	s := srveffects.NewService()
	as := srvauth.NewService()
	ps := srvparties.NewService()

	return &Controller{
		s,
		as,
		ps,
	}
}

// GetEffects (GET /parties/{partyId}/effects/catalog)
func (c *Controller) GetEffects(ctx echo.Context, partyId geneffects.PartyId) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	effects, err := c.Service.GetEffects(ctx.Request().Context(), partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertEffectsToDto(effects))
}

// GetRemovedEffects (GET /parties/{partyId}/effects/catalog/removed)
func (c *Controller) GetRemovedEffects(ctx echo.Context, partyId geneffects.PartyId) error {
	err := common.RequireAdmin(ctx, c.AuthService, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	effects, err := c.Service.GetRemovedEffects(ctx.Request().Context(), partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertEffectsToDto(effects))
}

// CreateEffect (POST /parties/{partyId}/effects/catalog)
func (c *Controller) CreateEffect(ctx echo.Context, partyId geneffects.PartyId) error {
	err := common.RequireAdmin(ctx, c.AuthService, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var effectDto geneffects.EffectCreate
	err = ctx.Bind(&effectDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = validator.ValidateEffectUseCount(effectDto.UseCount)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = validator.ValidateEffectDurationInSeconds(effectDto.DurationInSeconds)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var duration *time.Duration
	if effectDto.DurationInSeconds != nil {
		effectDuration := time.Duration(*effectDto.DurationInSeconds) * time.Second
		duration = &effectDuration
	}

	effect, err := c.Service.CreateEffect(
		ctx.Request().Context(),
		partyId,
		effectDto.Name,
		effectDto.Description,
		effectDto.UseCount,
		duration,
		convertDtoToNamedChangeEntries(effectDto.Entries),
		convertDtoToNamedModifiers(effectDto.Modifiers))

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, convertEffectToDto(effect))
}

// RemoveEffect (DELETE /parties/{partyId}/effects/catalog/{name})
func (c *Controller) RemoveEffect(ctx echo.Context, partyId geneffects.PartyId, name geneffects.Name) error {
	err := common.RequireAdmin(ctx, c.AuthService, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.RemoveEffect(ctx.Request().Context(), partyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetUserEffects (GET /parties/{partyId}/effects/{login})
func (c *Controller) GetUserEffects(ctx echo.Context, partyId geneffects.PartyId, login geneffects.Login) error {
	_, userId, err := c.userIdFromLogin(ctx, partyId, login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	userEffects, err := c.Service.GetNamedUserEffects(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	userEffectsDto := make(geneffects.UserEffects, len(userEffects))
	for i, userEffect := range userEffects {
		userEffectsDto[i] = geneffects.UserEffect{
			Name:        userEffect.Name,
			Description: userEffect.Description,
			UseCount:    userEffect.UseCount,
			UsesLeft:    userEffect.UsesLeft,
			Duration:    convertDurationToDto(userEffect.Duration),
			StartedDate: userEffect.StartedDate,
			Modifiers:   convertNamedModifiersToDto(userEffect.Modifiers),
		}
	}

	return ctx.JSON(http.StatusOK, userEffectsDto)
}

// UseUserEffect (POST /parties/{partyId}/effects/{login}/{name}/use)
func (c *Controller) UseUserEffect(ctx echo.Context, partyId geneffects.PartyId, login geneffects.Login, name geneffects.Name) error {
	actorUserId, userId, err := c.userIdFromLogin(ctx, partyId, login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.UseEffect(ctx.Request().Context(), actorUserId, userId, partyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// EndUserEffect (DELETE /parties/{partyId}/effects/{login}/{name})
func (c *Controller) EndUserEffect(ctx echo.Context, partyId geneffects.PartyId, login geneffects.Login, name geneffects.Name) error {
	actorUserId, userId, err := c.userIdFromLogin(ctx, partyId, login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.EndUserEffect(ctx.Request().Context(), actorUserId, userId, partyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetUserEffectHistory (GET /parties/{partyId}/effects/{login}/history)
func (c *Controller) GetUserEffectHistory(ctx echo.Context, partyId geneffects.PartyId, login geneffects.Login) error {
	_, userId, err := c.userIdFromLogin(ctx, partyId, login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	history, err := c.Service.GetEffectHistory(ctx.Request().Context(), userId, partyId)

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

	historyDto := make(geneffects.EffectHistoryEntries, len(history))
	for i, entry := range history {
		historyDto[i] = geneffects.EffectHistoryEntry{
			Name:        entry.Name,
			Description: entry.Description,
			Action:      entry.Action,
			ActorLogin:  loginsByUserId[entry.ActorUserId],
			UseCount:    entry.UseCount,
			UsesLeft:    entry.UsesLeft,
			CreatedDate: entry.CreatedDate,
		}
	}

	return ctx.JSON(http.StatusOK, historyDto)
}

// userIdFromLogin resolves a login path parameter, rejecting the request when it carries no session
// or comes from a user outside the party. It returns the user the request comes from along with the
// user the login names.
func (c *Controller) userIdFromLogin(ctx echo.Context, partyId geneffects.PartyId, login string) (actorUserId int, userId int, err error) {
	actorUserId, err = c.AuthService.GetUserId(ctx)

	if err != nil {
		return
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), actorUserId, partyId)

	if err != nil {
		return
	}

	userId, err = c.AuthService.GetUserIdByLogin(ctx.Request().Context(), login)

	return
}

func convertDurationToDto(duration *time.Duration) *geneffects.Duration {
	if duration == nil {
		return nil
	}

	converted := geneffects.Duration(common.DurationToISO8601(*duration))

	return &converted
}

func convertEffectToDto(effect typeeffects.NamedEffect) geneffects.Effect {
	return geneffects.Effect{
		Name:        effect.Name,
		Description: effect.Description,
		UseCount:    effect.UseCount,
		Duration:    convertDurationToDto(effect.Duration),
		Modifiers:   convertNamedModifiersToDto(effect.Modifiers),
	}
}

func convertNamedModifiersToDto(modifiers []typeeffects.NamedPointModifier) []geneffects.PointModifier {
	modifiersDto := make([]geneffects.PointModifier, len(modifiers))

	for i, modifier := range modifiers {
		modifiersDto[i] = geneffects.PointModifier{
			PointTypeName: modifier.PointTypeName,
			Amount:        modifier.Amount,
		}
	}

	return modifiersDto
}

func convertDtoToNamedModifiers(modifiersDto *[]geneffects.PointModifier) []typeeffects.NamedPointModifier {
	if modifiersDto == nil {
		return nil
	}

	modifiers := make([]typeeffects.NamedPointModifier, len(*modifiersDto))

	for i, modifierDto := range *modifiersDto {
		modifiers[i] = typeeffects.NamedPointModifier{
			PointTypeName: modifierDto.PointTypeName,
			Amount:        modifierDto.Amount,
		}
	}

	return modifiers
}

func convertEffectsToDto(effects []typeeffects.NamedEffect) geneffects.Effects {
	effectsDto := make(geneffects.Effects, len(effects))

	for i, effect := range effects {
		effectsDto[i] = convertEffectToDto(effect)
	}

	return effectsDto
}

func convertDtoToNamedChangeEntries(entriesDto geneffects.NamedChangeEntries) []typechanges.NamedChangeEntry {
	inputs := make([]typechanges.NamedChangeEntry, len(entriesDto))

	for i, entryDto := range entriesDto {
		inputs[i] = typechanges.NamedChangeEntry{
			PointTypeName: entryDto.PointTypeName,
			ItemName:      entryDto.ItemName,
			PerkName:      entryDto.PerkName,
			EffectName:    entryDto.EffectName,
			Amount:        entryDto.Amount,
		}
	}

	return inputs
}
