package ctrlparties

import (
	"FGG-Service/api/generated/parties"
	"FGG-Service/src/auth/service"
	"FGG-Service/src/common"
	"FGG-Service/src/parties/service"
	"FGG-Service/src/parties/types"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Controller struct {
	Service     srvparties.IService
	AuthService srvauth.IService
}

func NewController() *Controller {
	s := srvparties.NewService()
	as := srvauth.NewService()

	return &Controller{
		s,
		as,
	}
}

// CreateParty (POST /parties)
//
// Creating a party is not admin-only: admin rights come from party membership, so requiring them
// here would leave a party with nobody able to create it. The creator becomes its first admin.
func (c *Controller) CreateParty(ctx echo.Context) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var partyDto genparties.PartyCreate
	err = ctx.Bind(&partyDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	party, err := c.Service.CreateParty(userId, partyDto.Name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, convertPartyToDto(party))
}

// GetParties (GET /parties)
func (c *Controller) GetParties(ctx echo.Context) error {
	err := common.RequireAdmin(ctx, c.AuthService)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	parties, err := c.Service.GetParties()

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	partiesDto := make(genparties.Parties, len(parties))
	for i, party := range parties {
		partiesDto[i] = convertPartyToDto(party)
	}

	return ctx.JSON(http.StatusOK, partiesDto)
}

// GetParty (GET /parties/{partyId})
func (c *Controller) GetParty(ctx echo.Context, partyId genparties.PartyId) error {
	_, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	party, err := c.Service.GetParty(partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertPartyToDto(party))
}

// ChangeParty (PATCH /parties/{partyId})
func (c *Controller) ChangeParty(ctx echo.Context, partyId genparties.PartyId) error {
	err := common.RequireAdmin(ctx, c.AuthService)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var partyDto genparties.PartyChange
	err = ctx.Bind(&partyDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.ChangePartyName(partyId, partyDto.Name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetMembers (GET /parties/{partyId}/members)
func (c *Controller) GetMembers(ctx echo.Context, partyId genparties.PartyId) error {
	_, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	members, err := c.Service.GetMembers(partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	membersDto := make(genparties.Members, 0, len(members))
	for _, member := range members {
		if member.LeftDate != nil {
			continue
		}

		membersDto = append(membersDto, genparties.Member{
			Login:       member.Login,
			DisplayName: member.DisplayName,
			IsAdmin:     member.IsAdmin,
			JoinedDate:  member.JoinedDate,
		})
	}

	return ctx.JSON(http.StatusOK, membersDto)
}

// AddMember (POST /parties/{partyId}/members)
func (c *Controller) AddMember(ctx echo.Context, partyId genparties.PartyId) error {
	err := common.RequireAdmin(ctx, c.AuthService)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var memberDto genparties.MemberCreate
	err = ctx.Bind(&memberDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	userId, err := c.AuthService.GetUserIdByLogin(memberDto.Login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	member, err := c.Service.AddMember(userId, partyId, memberDto.DisplayName, memberDto.IsAdmin)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, genparties.Member{
		Login:       memberDto.Login,
		DisplayName: member.DisplayName,
		IsAdmin:     member.IsAdmin,
		JoinedDate:  member.JoinedDate,
	})
}

// ChangeMember (PATCH /parties/{partyId}/members/{login})
func (c *Controller) ChangeMember(ctx echo.Context, partyId genparties.PartyId, login genparties.Login) error {
	err := common.RequireAdmin(ctx, c.AuthService)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var memberDto genparties.MemberChange
	err = ctx.Bind(&memberDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	userId, err := c.AuthService.GetUserIdByLogin(login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	if memberDto.IsAdmin != nil && !*memberDto.IsAdmin {
		currentUserId, err := c.AuthService.GetUserId(ctx)

		if err != nil {
			return common.SendJSONErrorResponse(ctx, err)
		}

		if currentUserId == userId {
			return common.SendJSONErrorResponse(ctx, common.NewOwnAdminRightsRevokeConflictError())
		}
	}

	err = c.Service.ChangeMember(userId, partyId, memberDto.DisplayName, memberDto.IsAdmin)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// RemoveMember (DELETE /parties/{partyId}/members/{login})
func (c *Controller) RemoveMember(ctx echo.Context, partyId genparties.PartyId, login genparties.Login) error {
	currentUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	userId, err := c.AuthService.GetUserIdByLogin(login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	if userId != currentUserId {
		err = common.RequireAdmin(ctx, c.AuthService)

		if err != nil {
			return common.SendJSONErrorResponse(ctx, err)
		}
	}

	err = c.Service.RemoveMember(userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

func convertPartyToDto(party typeparties.Party) genparties.Party {
	return genparties.Party{
		Id:          party.Id,
		Name:        party.Name,
		CreatedDate: party.CreatedDate,
	}
}
