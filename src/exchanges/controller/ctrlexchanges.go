package ctrlexchanges

import (
	"FGG-Service/api/generated/exchanges"
	"FGG-Service/src/auth/service"
	"FGG-Service/src/changes/types"
	"FGG-Service/src/common"
	"FGG-Service/src/exchanges/service"
	"FGG-Service/src/exchanges/types"
	"net/http"

	"github.com/labstack/echo/v4"
)

// defaultPartyId is a stopgap until real party context exists (see project plan) — every exchange is
// scoped to this one hardcoded party.
const defaultPartyId = 1

type Controller struct {
	Service     srvexchanges.IService
	AuthService srvauth.IService
}

func NewController() *Controller {
	s := srvexchanges.NewService()
	as := srvauth.NewService()

	return &Controller{
		s,
		as,
	}
}

// GetExchanges (GET /exchanges/catalog)
func (c *Controller) GetExchanges(ctx echo.Context) error {
	_, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	exchanges, err := c.Service.GetExchanges(defaultPartyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertExchangesToDto(exchanges))
}

// GetRemovedExchanges (GET /exchanges/catalog/removed)
func (c *Controller) GetRemovedExchanges(ctx echo.Context) error {
	err := common.RequireAdmin(ctx, c.AuthService)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	exchanges, err := c.Service.GetRemovedExchanges(defaultPartyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertExchangesToDto(exchanges))
}

// CreateExchange (POST /exchanges/catalog)
func (c *Controller) CreateExchange(ctx echo.Context) error {
	err := common.RequireAdmin(ctx, c.AuthService)

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
		defaultPartyId,
		exchangeDto.Name,
		exchangeDto.Description,
		convertDtoToChangeEntryInputs(exchangeDto.SourceEntries),
		convertDtoToChangeEntryInputs(exchangeDto.TargetEntries))

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertExchangeToDto(exchange))
}

// RemoveExchange (DELETE /exchanges/catalog/{name})
func (c *Controller) RemoveExchange(ctx echo.Context, name genexchanges.Name) error {
	err := common.RequireAdmin(ctx, c.AuthService)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.RemoveExchange(defaultPartyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// UseExchange (POST /exchanges/catalog/{name}/use)
func (c *Controller) UseExchange(ctx echo.Context, name genexchanges.Name) error {
	userId, err := c.AuthService.GetUserId(ctx)

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

	err = c.Service.UseExchange(userId, defaultPartyId, name, targetUserIds)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetUserExchangeHistory (GET /exchanges/{login}/history)
func (c *Controller) GetUserExchangeHistory(ctx echo.Context, login genexchanges.Login) error {
	_, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	userId, err := c.AuthService.GetUserIdByLogin(login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	history, err := c.Service.GetExchangeHistory(userId, defaultPartyId)

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