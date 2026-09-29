package ctrlpoints

import (
	"FGG-Service/api/generated/points"
	"FGG-Service/src/auth/service"
	"FGG-Service/src/common"
	"FGG-Service/src/points/service"
	"FGG-Service/src/points/type"
	"net/http"

	"github.com/labstack/echo/v4"
)

// defaultPartyId is a stopgap until real party context exists (see project plan) — every point is
// scoped to this one hardcoded party.
const defaultPartyId = 1

type Controller struct {
	Service     srvpoints.Service
	AuthService srvauth.IService
}

func NewController() *Controller {
	s := srvpoints.NewService()
	as := srvauth.NewService()

	return &Controller{
		*s,
		as,
	}
}

// GetPointTypes (GET /points/types)
func (c *Controller) GetPointTypes(ctx echo.Context) error {
	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	pointTypes, err := c.Service.GetVisiblePointTypes(actorUserId, defaultPartyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertPointTypesToDto(pointTypes))
}

// CreatePointType (POST /points/types)
func (c *Controller) CreatePointType(ctx echo.Context) error {
	err := common.RequireAdmin(ctx, c.AuthService)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var pointTypeDto genpoints.PointTypeCreate
	err = ctx.Bind(&pointTypeDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	created, err := c.Service.CreatePointType(defaultPartyId, typepoints.PointType{
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

// ChangePointType (PATCH /points/types/{name})
func (c *Controller) ChangePointType(ctx echo.Context, name genpoints.Name) error {
	err := common.RequireAdmin(ctx, c.AuthService)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var pointTypeDto genpoints.PointTypeChange
	err = ctx.Bind(&pointTypeDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.ChangePointType(defaultPartyId, name, typepoints.PointType{
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

// RemovePointType (DELETE /points/types/{name})
func (c *Controller) RemovePointType(ctx echo.Context, name genpoints.Name) error {
	err := common.RequireAdmin(ctx, c.AuthService)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.RemovePointType(defaultPartyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}
// GetUserPoints (GET /points/users/{login})
func (c *Controller) GetUserPoints(ctx echo.Context, login genpoints.Login) error {
	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	affectedUserId, err := c.AuthService.GetUserIdByLogin(login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	values, err := c.Service.GetUserPoints(actorUserId, affectedUserId, defaultPartyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertUserPointsToDto(values))
}

// GetAllUserPoints (GET /points/users)
func (c *Controller) GetAllUserPoints(ctx echo.Context) error {
	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	byLogin, err := c.Service.GetAllUserPoints(actorUserId, defaultPartyId)

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

// GetUserPointValue (GET /points/users/{login}/types/{name})
func (c *Controller) GetUserPointValue(ctx echo.Context, login genpoints.Login, name genpoints.Name) error {
	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	affectedUserId, err := c.AuthService.GetUserIdByLogin(login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	value, err := c.Service.GetUserPointValueByTypeName(actorUserId, affectedUserId, defaultPartyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, value)
}

// ChangeUserPointValue (PATCH /points/users/{login}/types/{name})
func (c *Controller) ChangeUserPointValue(ctx echo.Context, login genpoints.Login, name genpoints.Name) error {
	actorUserId, err := common.RequireAdminUserId(ctx, c.AuthService)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var changeDto genpoints.PointChange
	err = ctx.Bind(&changeDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	affectedUserId, err := c.AuthService.GetUserIdByLogin(login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	result, err := c.Service.ChangeUserPointByTypeName(actorUserId, affectedUserId, defaultPartyId, name, changeDto.DesiredChangeValue)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertChangeResultToDto(result))
}

// GetUserPointHistory (GET /points/users/{login}/types/{name}/history)
func (c *Controller) GetUserPointHistory(ctx echo.Context, login genpoints.Login, name genpoints.Name) error {
	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	affectedUserId, err := c.AuthService.GetUserIdByLogin(login)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	history, err := c.Service.GetUserPointHistoryByTypeName(actorUserId, affectedUserId, defaultPartyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	historyDto, err := c.convertHistoryToDto(history)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, historyDto)
}
// GetAllPartyPoints (GET /points/party)
func (c *Controller) GetAllPartyPoints(ctx echo.Context) error {
	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	values, err := c.Service.GetAllPartyPoints(actorUserId, defaultPartyId)

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

// GetPartyPointValue (GET /points/party/types/{name})
func (c *Controller) GetPartyPointValue(ctx echo.Context, name genpoints.Name) error {
	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	value, err := c.Service.GetPartyPointValueByTypeName(actorUserId, defaultPartyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, value)
}

// ChangePartyPointValue (PATCH /points/party/types/{name})
func (c *Controller) ChangePartyPointValue(ctx echo.Context, name genpoints.Name) error {
	actorUserId, err := common.RequireAdminUserId(ctx, c.AuthService)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var changeDto genpoints.PointChange
	err = ctx.Bind(&changeDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	result, err := c.Service.ChangePartyPointByTypeName(actorUserId, defaultPartyId, name, changeDto.DesiredChangeValue)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.JSON(http.StatusOK, convertChangeResultToDto(result))
}

// GetPartyPointHistory (GET /points/party/types/{name}/history)
func (c *Controller) GetPartyPointHistory(ctx echo.Context, name genpoints.Name) error {
	actorUserId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	history, err := c.Service.GetPartyPointHistoryByTypeName(actorUserId, defaultPartyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	historyDto, err := c.convertHistoryToDto(history)

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

// convertHistoryToDto names the user behind each recorded change, resolving each id once.
func (c *Controller) convertHistoryToDto(history []typepoints.PointHistoryEntry) (genpoints.PointHistoryEntries, error) {
	historyDto := make(genpoints.PointHistoryEntries, len(history))
	loginsByUserId := make(map[int]string)

	for i, entry := range history {
		login, isKnown := loginsByUserId[entry.ActorUserId]

		if !isKnown {
			resolved, err := c.AuthService.GetLoginByUserId(entry.ActorUserId)

			if err != nil {
				return nil, err
			}

			login = resolved
			loginsByUserId[entry.ActorUserId] = login
		}

		actorLogin := login

		historyDto[i] = genpoints.PointHistoryEntry{
			DesiredChangeValue: entry.DesiredChangeValue,
			ActualChangeValue:  entry.ActualChangeValue,
			FinalValue:         entry.FinalValue,
			ActorLogin:        &actorLogin,
			ChangedDate:        entry.ChangedDate,
		}
	}

	return historyDto, nil
}