package ctrlitems

import (
	"FGG-Service/api/generated/items"
	"FGG-Service/src/auth/service"
	"FGG-Service/src/changes/types"
	"FGG-Service/src/common"
	"FGG-Service/src/items/service"
	"FGG-Service/src/items/types"
	"net/http"

	"github.com/labstack/echo/v4"
)

// defaultPartyId is a stopgap until real party context exists (see project plan) — every item is
// scoped to this one hardcoded party.
const defaultPartyId = 1

type Controller struct {
	Service     srvitems.IService
	AuthService srvauth.IService
}

func NewController() *Controller {
	s := srvitems.NewService()
	as := srvauth.NewService()

	return &Controller{
		s,
		as,
	}
}

// GetItems (GET /items/catalog)
func (c *Controller) GetItems(ctx echo.Context) error {
	_, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	items, err := c.Service.GetItems(defaultPartyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertItemsToDto(items))
}

// GetRemovedItems (GET /items/catalog/removed)
func (c *Controller) GetRemovedItems(ctx echo.Context) error {
	err := common.RequireAdmin(ctx, c.AuthService)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	items, err := c.Service.GetRemovedItems(defaultPartyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertItemsToDto(items))
}

// CreateItem (POST /items/catalog)
func (c *Controller) CreateItem(ctx echo.Context) error {
	err := common.RequireAdmin(ctx, c.AuthService)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var itemDto genitems.ItemCreate
	err = ctx.Bind(&itemDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	item, err := c.Service.CreateItem(
		defaultPartyId,
		itemDto.Name,
		itemDto.Description,
		itemDto.UseCount,
		convertDtoToChangeEntryInputs(itemDto.Entries))

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, convertItemToDto(item))
}

// RemoveItem (DELETE /items/catalog/{name})
func (c *Controller) RemoveItem(ctx echo.Context, name genitems.Name) error {
	err := common.RequireAdmin(ctx, c.AuthService)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.RemoveItem(defaultPartyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}
// GetUserItems (GET /items/{login})
func (c *Controller) GetUserItems(ctx echo.Context, login genitems.Login) error {
	userId, err := c.userIdFromLogin(ctx, login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	userItems, err := c.Service.GetUserItems(userId, defaultPartyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	userItemsDto := make(genitems.UserItems, len(userItems))
	for i, userItem := range userItems {
		userItemsDto[i] = genitems.UserItem{
			Name:         userItem.Name,
			Description:  userItem.Description,
			UsesLeft:     userItem.UsesLeft,
			ReceivedDate: userItem.ReceivedDate,
		}
	}

	return ctx.JSON(http.StatusOK, userItemsDto)
}

// UseUserItem (POST /items/{login}/{name}/use)
func (c *Controller) UseUserItem(ctx echo.Context, login genitems.Login, name genitems.Name) error {
	userId, err := c.userIdFromLogin(ctx, login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.UseItem(userId, defaultPartyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// DiscardUserItem (DELETE /items/{login}/{name})
func (c *Controller) DiscardUserItem(ctx echo.Context, login genitems.Login, name genitems.Name) error {
	userId, err := c.userIdFromLogin(ctx, login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.DiscardUserItem(userId, defaultPartyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetUserItemHistory (GET /items/{login}/history)
func (c *Controller) GetUserItemHistory(ctx echo.Context, login genitems.Login) error {
	userId, err := c.userIdFromLogin(ctx, login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	history, err := c.Service.GetItemHistory(userId, defaultPartyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	historyDto := make(genitems.ItemHistoryEntries, len(history))
	for i, entry := range history {
		historyDto[i] = genitems.ItemHistoryEntry{
			Name:        entry.Name,
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

func convertItemToDto(item typeitems.Item) genitems.Item {
	return genitems.Item{
		Name:        item.Name,
		Description: item.Description,
		UseCount:    item.UseCount,
	}
}

func convertItemsToDto(items []typeitems.Item) genitems.Items {
	itemsDto := make(genitems.Items, len(items))

	for i, item := range items {
		itemsDto[i] = convertItemToDto(item)
	}

	return itemsDto
}

func convertDtoToChangeEntryInputs(entriesDto genitems.ChangeEntryInputs) []typechanges.ChangeEntryInput {
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