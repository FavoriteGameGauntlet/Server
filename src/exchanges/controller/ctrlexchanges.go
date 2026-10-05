package ctrlexchanges

import (
	"FGG-Service/api/generated/exchanges"
	"FGG-Service/src/auth/service"
	"FGG-Service/src/changes/types"
	"FGG-Service/src/common"
	"FGG-Service/src/exchanges/service"
	"FGG-Service/src/exchanges/types"
	"FGG-Service/src/parties/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Controller struct {
	Service      srvexchanges.IService
	AuthService  srvauth.IService
	PartyService srvparties.IService
}

func NewController() *Controller {
	s := srvexchanges.NewService()
	as := srvauth.NewService()
	ps := srvparties.NewService()

	return &Controller{
		s,
		as,
		ps,
	}
}

// GetExchanges (GET /parties/{partyId}/exchanges/catalog)
func (c *Controller) GetExchanges(ctx echo.Context, partyId genexchanges.PartyId) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	exchanges, err := c.Service.GetExchanges(partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertExchangesToDto(exchanges))
}

// GetRemovedExchanges (GET /parties/{partyId}/exchanges/catalog/removed)
func (c *Controller) GetRemovedExchanges(ctx echo.Context, partyId genexchanges.PartyId) error {
	err := common.RequireAdmin(ctx, c.AuthService, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	exchanges, err := c.Service.GetRemovedExchanges(partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertExchangesToDto(exchanges))
}

// CreateExchange (POST /parties/{partyId}/exchanges/catalog)
func (c *Controller) CreateExchange(ctx echo.Context, partyId genexchanges.PartyId) error {
	err := common.RequireAdmin(ctx, c.AuthService, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var exchangeDto genexchanges.ExchangeCreate
	err = ctx.Bind(&exchangeDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	exchange, err := c.Service.CreateExchange(
		partyId,
		exchangeDto.Name,
		exchangeDto.Description,
		convertDtoToChangeEntryInputs(exchangeDto.SourceEntries),
		convertDtoToChangeEntryInputs(exchangeDto.TargetEntries))

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, convertExchangeToDto(exchange))
}

// RemoveExchange (DELETE /parties/{partyId}/exchanges/catalog/{name})
func (c *Controller) RemoveExchange(ctx echo.Context, partyId genexchanges.PartyId, name genexchanges.Name) error {
	err := common.RequireAdmin(ctx, c.AuthService, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.RemoveExchange(partyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// UseExchange (POST /parties/{partyId}/exchanges/catalog/{name}/use)
func (c *Controller) UseExchange(ctx echo.Context, partyId genexchanges.PartyId, name genexchanges.Name) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var useDto genexchanges.ExchangeUse
	err = ctx.Bind(&useDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	var targetUserIds []int
	if useDto.TargetLogins != nil {
		targetUserIds = make([]int, len(*useDto.TargetLogins))

		for i, login := range *useDto.TargetLogins {
			targetUserIds[i], err = c.AuthService.GetUserIdByLogin(login)

			if err != nil {
				return common.SendJSONErrorResponse(ctx, err)
			}
		}
	}

	err = c.Service.UseExchange(userId, partyId, name, targetUserIds)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetUserExchangeHistory (GET /parties/{partyId}/exchanges/{login}/history)
func (c *Controller) GetUserExchangeHistory(ctx echo.Context, partyId genexchanges.PartyId, login genexchanges.Login) error {
	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(actorUserId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	userId, err := c.AuthService.GetUserIdByLogin(login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	history, err := c.Service.GetExchangeHistory(userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	historyDto := make(genexchanges.ExchangeHistoryEntries, len(history))
	for i, entry := range history {
		historyDto[i] = genexchanges.ExchangeHistoryEntry{
			Name:        entry.Name,
			Description: entry.Description,
			UsedDate:    entry.UsedDate,
		}
	}

	return ctx.JSON(http.StatusOK, historyDto)
}

func convertExchangeToDto(exchange typeexchanges.Exchange) genexchanges.Exchange {
	return genexchanges.Exchange{
		Name:        exchange.Name,
		Description: exchange.Description,
	}
}

func convertExchangesToDto(exchanges []typeexchanges.Exchange) genexchanges.Exchanges {
	exchangesDto := make(genexchanges.Exchanges, len(exchanges))

	for i, exchange := range exchanges {
		exchangesDto[i] = convertExchangeToDto(exchange)
	}

	return exchangesDto
}

func convertDtoToChangeEntryInputs(entriesDto genexchanges.ChangeEntryInputs) []typechanges.ChangeEntryInput {
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
