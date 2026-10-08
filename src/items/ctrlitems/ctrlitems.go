package ctrlitems

import (
	"FGG-Service/api/generated/items"
	"FGG-Service/src/auth/srvauth"
	"FGG-Service/src/changes/typechanges"
	"FGG-Service/src/common"
	"FGG-Service/src/items/srvitems"
	"FGG-Service/src/items/typeitems"
	"FGG-Service/src/parties/srvparties"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Controller struct {
	Service      srvitems.IService
	AuthService  srvauth.IService
	PartyService srvparties.IService
}

func NewController() *Controller {
	s := srvitems.NewService()
	as := srvauth.NewService()
	ps := srvparties.NewService()

	return &Controller{
		s,
		as,
		ps,
	}
}

// GetItems (GET /parties/{partyId}/items/catalog)
func (c *Controller) GetItems(ctx echo.Context, partyId genitems.PartyId) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	items, err := c.Service.GetItems(ctx.Request().Context(), partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertItemsToDto(items))
}

// GetRemovedItems (GET /parties/{partyId}/items/catalog/removed)
func (c *Controller) GetRemovedItems(ctx echo.Context, partyId genitems.PartyId) error {
	err := common.RequireAdmin(ctx, c.AuthService, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	items, err := c.Service.GetRemovedItems(ctx.Request().Context(), partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertItemsToDto(items))
}

// CreateItem (POST /parties/{partyId}/items/catalog)
func (c *Controller) CreateItem(ctx echo.Context, partyId genitems.PartyId) error {
	err := common.RequireAdmin(ctx, c.AuthService, partyId)

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
		ctx.Request().Context(),
		partyId,
		itemDto.Name,
		itemDto.Description,
		itemDto.UseCount,
		convertDtoToNamedChangeEntries(itemDto.Entries))

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, convertItemToDto(item))
}

// RemoveItem (DELETE /parties/{partyId}/items/catalog/{name})
func (c *Controller) RemoveItem(ctx echo.Context, partyId genitems.PartyId, name genitems.Name) error {
	err := common.RequireAdmin(ctx, c.AuthService, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.RemoveItem(ctx.Request().Context(), partyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetUserItems (GET /parties/{partyId}/items/{login})
func (c *Controller) GetUserItems(ctx echo.Context, partyId genitems.PartyId, login genitems.Login) error {
	_, userId, err := c.userIdFromLogin(ctx, partyId, login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	userItems, err := c.Service.GetUserItems(ctx.Request().Context(), userId, partyId)

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
			Entries:      convertNamedChangeEntriesToDto(userItem.Entries),
		}
	}

	return ctx.JSON(http.StatusOK, userItemsDto)
}

// UseUserItem (POST /parties/{partyId}/items/{login}/{name}/use)
func (c *Controller) UseUserItem(ctx echo.Context, partyId genitems.PartyId, login genitems.Login, name genitems.Name) error {
	actorUserId, userId, err := c.userIdFromLogin(ctx, partyId, login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.UseItem(ctx.Request().Context(), actorUserId, userId, partyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// DiscardUserItem (DELETE /parties/{partyId}/items/{login}/{name})
func (c *Controller) DiscardUserItem(ctx echo.Context, partyId genitems.PartyId, login genitems.Login, name genitems.Name) error {
	actorUserId, userId, err := c.userIdFromLogin(ctx, partyId, login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.DiscardUserItem(ctx.Request().Context(), actorUserId, userId, partyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetUserItemHistory (GET /parties/{partyId}/items/{login}/history)
func (c *Controller) GetUserItemHistory(ctx echo.Context, partyId genitems.PartyId, login genitems.Login) error {
	_, userId, err := c.userIdFromLogin(ctx, partyId, login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	history, err := c.Service.GetItemHistory(ctx.Request().Context(), userId, partyId)

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

	historyDto := make(genitems.ItemHistoryEntries, len(history))
	for i, entry := range history {
		historyDto[i] = genitems.ItemHistoryEntry{
			Name:        entry.Name,
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
func (c *Controller) userIdFromLogin(ctx echo.Context, partyId genitems.PartyId, login string) (actorUserId int, userId int, err error) {
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

func convertItemToDto(item typeitems.ItemWithEntries) genitems.Item {
	return genitems.Item{
		Name:        item.Name,
		Description: item.Description,
		UseCount:    item.UseCount,
		Entries:     convertNamedChangeEntriesToDto(item.Entries),
	}
}

func convertItemsToDto(items []typeitems.ItemWithEntries) genitems.Items {
	itemsDto := make(genitems.Items, len(items))

	for i, item := range items {
		itemsDto[i] = convertItemToDto(item)
	}

	return itemsDto
}

func convertNamedChangeEntriesToDto(inputs []typechanges.NamedChangeEntry) genitems.NamedChangeEntries {
	entriesDto := make(genitems.NamedChangeEntries, len(inputs))

	for i, input := range inputs {
		entriesDto[i] = genitems.NamedChangeEntry{
			PointTypeName: input.PointTypeName,
			ItemName:      input.ItemName,
			PerkName:      input.PerkName,
			EffectName:    input.EffectName,
			Amount:        input.Amount,
		}
	}

	return entriesDto
}

func convertDtoToNamedChangeEntries(entriesDto genitems.NamedChangeEntries) []typechanges.NamedChangeEntry {
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
