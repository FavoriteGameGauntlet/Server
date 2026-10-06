package ctrlpoints

import (
	"FGG-Service/api/generated/points"
	"FGG-Service/src/auth/service"
	"FGG-Service/src/common"
	"FGG-Service/src/parties/service"
	"FGG-Service/src/points/service"
	"FGG-Service/src/points/type"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Controller struct {
	Service      srvpoints.Service
	AuthService  srvauth.IService
	PartyService srvparties.IService
}

func NewController() *Controller {
	s := srvpoints.NewService()
	as := srvauth.NewService()
	ps := srvparties.NewService()

	return &Controller{
		*s,
		as,
		ps,
	}
}

// GetPointTypes (GET /parties/{partyId}/points/types)
func (c *Controller) GetPointTypes(ctx echo.Context, partyId genpoints.PartyId) error {
	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), actorUserId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	pointTypes, err := c.Service.GetVisiblePointTypes(ctx.Request().Context(), actorUserId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertPointTypesToDto(pointTypes))
}

// CreatePointType (POST /parties/{partyId}/points/types)
func (c *Controller) CreatePointType(ctx echo.Context, partyId genpoints.PartyId) error {
	err := common.RequireAdmin(ctx, c.AuthService, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var pointTypeDto genpoints.PointTypeCreate
	err = ctx.Bind(&pointTypeDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	created, err := c.Service.CreatePointType(ctx.Request().Context(), partyId, typepoints.PointType{
		Name:        pointTypeDto.Name,
		Description: pointTypeDto.Description,
		StartValue:  pointTypeDto.StartValue,
		IsPublic:    pointTypeDto.IsPublic,
		IsShared:    pointTypeDto.IsShared,
		Minimum:     pointTypeDto.Minimum,
		Maximum:     pointTypeDto.Maximum,
	})

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, genpoints.PointType{
		Name:        created.Name,
		Description: created.Description,
		StartValue:  created.StartValue,
		IsPublic:    created.IsPublic,
		IsShared:    created.IsShared,
		Minimum:     created.Minimum,
		Maximum:     created.Maximum,
	})
}

// ChangePointType (PATCH /parties/{partyId}/points/types/{name})
func (c *Controller) ChangePointType(ctx echo.Context, partyId genpoints.PartyId, name genpoints.Name) error {
	err := common.RequireAdmin(ctx, c.AuthService, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var pointTypeDto genpoints.PointTypeChange
	err = ctx.Bind(&pointTypeDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.ChangePointType(ctx.Request().Context(), partyId, name, typepoints.PointType{
		Name:        pointTypeDto.Name,
		Description: pointTypeDto.Description,
		StartValue:  pointTypeDto.StartValue,
		IsPublic:    pointTypeDto.IsPublic,
		IsShared:    pointTypeDto.IsShared,
		Minimum:     pointTypeDto.Minimum,
		Maximum:     pointTypeDto.Maximum,
	})

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// RemovePointType (DELETE /parties/{partyId}/points/types/{name})
func (c *Controller) RemovePointType(ctx echo.Context, partyId genpoints.PartyId, name genpoints.Name) error {
	err := common.RequireAdmin(ctx, c.AuthService, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.RemovePointType(ctx.Request().Context(), partyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetUserPoints (GET /parties/{partyId}/points/users/{login})
func (c *Controller) GetUserPoints(ctx echo.Context, partyId genpoints.PartyId, login genpoints.Login) error {
	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), actorUserId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	affectedUserId, err := c.AuthService.GetUserIdByLogin(ctx.Request().Context(), login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	values, err := c.Service.GetUserPoints(ctx.Request().Context(), actorUserId, affectedUserId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertUserPointsToDto(values))
}

// GetAllUserPoints (GET /parties/{partyId}/points/users)
func (c *Controller) GetAllUserPoints(ctx echo.Context, partyId genpoints.PartyId) error {
	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), actorUserId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	byLogin, err := c.Service.GetAllUserPoints(ctx.Request().Context(), actorUserId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	byLoginDto := make(genpoints.UserPointsByLogins, len(byLogin))
	for i, entry := range byLogin {
		byLoginDto[i].Login = entry.Login
		byLoginDto[i].Points = convertUserPointsToDto(entry.Points)
	}

	return ctx.JSON(http.StatusOK, byLoginDto)
}

// GetUserPointValue (GET /parties/{partyId}/points/users/{login}/types/{name})
func (c *Controller) GetUserPointValue(ctx echo.Context, partyId genpoints.PartyId, login genpoints.Login, name genpoints.Name) error {
	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), actorUserId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	affectedUserId, err := c.AuthService.GetUserIdByLogin(ctx.Request().Context(), login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	value, err := c.Service.GetUserPointValueByTypeName(ctx.Request().Context(), actorUserId, affectedUserId, partyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, value)
}

// ChangeUserPointValue (PATCH /parties/{partyId}/points/users/{login}/types/{name})
func (c *Controller) ChangeUserPointValue(ctx echo.Context, partyId genpoints.PartyId, login genpoints.Login, name genpoints.Name) error {
	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), actorUserId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var changeDto genpoints.PointChange
	err = ctx.Bind(&changeDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	affectedUserId, err := c.AuthService.GetUserIdByLogin(ctx.Request().Context(), login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	result, err := c.Service.ChangeUserPointByTypeName(ctx.Request().Context(), actorUserId, affectedUserId, partyId, name, changeDto.DesiredChangeValue)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertChangeResultToDto(result))
}

// GetUserPointHistory (GET /parties/{partyId}/points/users/{login}/types/{name}/history)
func (c *Controller) GetUserPointHistory(ctx echo.Context, partyId genpoints.PartyId, login genpoints.Login, name genpoints.Name) error {
	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), actorUserId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	affectedUserId, err := c.AuthService.GetUserIdByLogin(ctx.Request().Context(), login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	history, err := c.Service.GetUserPointHistoryByTypeName(ctx.Request().Context(), actorUserId, affectedUserId, partyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	historyDto, err := c.convertHistoryToDto(ctx, history)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, historyDto)
}

// GetAllPartyPoints (GET /parties/{partyId}/points/party)
func (c *Controller) GetAllPartyPoints(ctx echo.Context, partyId genpoints.PartyId) error {
	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), actorUserId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	values, err := c.Service.GetAllPartyPoints(ctx.Request().Context(), actorUserId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	valuesDto := make(genpoints.PartyPoints, len(values))
	for i, value := range values {
		valuesDto[i] = genpoints.PartyPoint{
			PointType: convertPointTypeToDto(value.PointType),
			Value:     value.Value,
		}
	}

	return ctx.JSON(http.StatusOK, valuesDto)
}

// GetPartyPointValue (GET /parties/{partyId}/points/party/types/{name})
func (c *Controller) GetPartyPointValue(ctx echo.Context, partyId genpoints.PartyId, name genpoints.Name) error {
	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), actorUserId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	value, err := c.Service.GetPartyPointValueByTypeName(ctx.Request().Context(), actorUserId, partyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, value)
}

// ChangePartyPointValue (PATCH /parties/{partyId}/points/party/types/{name})
func (c *Controller) ChangePartyPointValue(ctx echo.Context, partyId genpoints.PartyId, name genpoints.Name) error {
	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), actorUserId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var changeDto genpoints.PointChange
	err = ctx.Bind(&changeDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	result, err := c.Service.ChangePartyPointByTypeName(ctx.Request().Context(), actorUserId, partyId, name, changeDto.DesiredChangeValue)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertChangeResultToDto(result))
}

// GetPartyPointHistory (GET /parties/{partyId}/points/party/types/{name}/history)
func (c *Controller) GetPartyPointHistory(ctx echo.Context, partyId genpoints.PartyId, name genpoints.Name) error {
	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(ctx.Request().Context(), actorUserId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	history, err := c.Service.GetPartyPointHistoryByTypeName(ctx.Request().Context(), actorUserId, partyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	historyDto, err := c.convertHistoryToDto(ctx, history)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, historyDto)
}

func convertPointTypeToDto(pointType typepoints.PointTypeInfo) genpoints.PointType {
	return genpoints.PointType{
		Name:        pointType.Name,
		Description: pointType.Description,
		StartValue:  pointType.StartValue,
		IsPublic:    pointType.IsPublic,
		IsShared:    pointType.IsShared,
		Minimum:     pointType.Minimum,
		Maximum:     pointType.Maximum,
	}
}

func convertPointTypesToDto(pointTypes []typepoints.PointTypeInfo) genpoints.PointTypes {
	pointTypesDto := make(genpoints.PointTypes, len(pointTypes))

	for i, pointType := range pointTypes {
		pointTypesDto[i] = convertPointTypeToDto(pointType)
	}

	return pointTypesDto
}

func convertUserPointsToDto(values []typepoints.PointValue) genpoints.UserPoints {
	valuesDto := make(genpoints.UserPoints, len(values))

	for i, value := range values {
		valuesDto[i] = genpoints.UserPoint{
			PointType: convertPointTypeToDto(value.PointType),
			Value:     value.Value,
		}
	}

	return valuesDto
}

func convertChangeResultToDto(result typepoints.PointChangeResult) genpoints.PointChangeResult {
	return genpoints.PointChangeResult{
		DesiredChangeValue: result.DesiredChangeValue,
		ActualChangeValue:  result.ActualChangeValue,
		FinalValue:         result.FinalValue,
	}
}

// convertHistoryToDto names the user behind each recorded change.
func (c *Controller) convertHistoryToDto(ctx echo.Context, history []typepoints.PointHistoryEntry) (genpoints.PointHistoryEntries, error) {
	actorUserIds := make([]int, len(history))
	for i, entry := range history {
		actorUserIds[i] = entry.ActorUserId
	}

	loginsByUserId, err := common.GetLoginsByUserIds(ctx.Request().Context(), c.AuthService, actorUserIds)

	if err != nil {
		return nil, err
	}

	historyDto := make(genpoints.PointHistoryEntries, len(history))

	for i, entry := range history {
		actorLogin := loginsByUserId[entry.ActorUserId]

		historyDto[i] = genpoints.PointHistoryEntry{
			DesiredChangeValue: entry.DesiredChangeValue,
			ActualChangeValue:  entry.ActualChangeValue,
			FinalValue:         entry.FinalValue,
			ActorLogin:         &actorLogin,
			ChangedDate:        entry.ChangedDate,
		}
	}

	return historyDto, nil
}
