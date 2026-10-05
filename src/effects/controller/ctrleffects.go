package ctrleffects

import (
	"FGG-Service/api/generated/effects"
	"FGG-Service/src/auth/service"
	"FGG-Service/src/changes/types"
	"FGG-Service/src/common"
	"FGG-Service/src/effects/service"
	"FGG-Service/src/effects/types"
	"FGG-Service/src/parties/service"
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

	err = c.PartyService.RequireMember(userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	effects, err := c.Service.GetEffects(partyId)

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

	effects, err := c.Service.GetRemovedEffects(partyId)

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

	var duration *time.Duration
	if effectDto.DurationInSeconds != nil {
		effectDuration := time.Duration(*effectDto.DurationInSeconds) * time.Second
		duration = &effectDuration
	}

	effect, err := c.Service.CreateEffect(
		partyId,
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

// RemoveEffect (DELETE /parties/{partyId}/effects/catalog/{name})
func (c *Controller) RemoveEffect(ctx echo.Context, partyId geneffects.PartyId, name geneffects.Name) error {
	err := common.RequireAdmin(ctx, c.AuthService, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.RemoveEffect(partyId, name)

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

	views, err := c.Service.GetUserEffectViews(userId, partyId)

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

// UseUserEffect (POST /parties/{partyId}/effects/{login}/{name}/use)
func (c *Controller) UseUserEffect(ctx echo.Context, partyId geneffects.PartyId, login geneffects.Login, name geneffects.Name) error {
	actorUserId, userId, err := c.userIdFromLogin(ctx, partyId, login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.UseEffect(actorUserId, userId, partyId, name)

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

	err = c.Service.EndUserEffect(actorUserId, userId, partyId, name)

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

	history, err := c.Service.GetEffectHistory(userId, partyId)

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

	err = c.PartyService.RequireMember(actorUserId, partyId)

	if err != nil {
		return
	}

	userId, err = c.AuthService.GetUserIdByLogin(login)

	return
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