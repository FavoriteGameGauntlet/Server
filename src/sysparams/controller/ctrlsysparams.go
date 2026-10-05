package ctrlsysparams

import (
	"FGG-Service/api/generated/system_parameters"
	"FGG-Service/src/auth/service"
	"FGG-Service/src/common"
	"FGG-Service/src/parties/service"
	"FGG-Service/src/sysparams/service"
	"FGG-Service/src/sysparams/types"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Controller struct {
	Service      srvsysparams.Service
	AuthService  srvauth.Service
	PartyService srvparties.IService
}

func NewController() *Controller {
	s := srvsysparams.NewService()
	as := srvauth.NewService()
	ps := srvparties.NewService()

	return &Controller{
		*s,
		*as,
		ps,
	}
}

// GetSystemParameters (GET /parties/{partyId}/system-parameters)
func (c *Controller) GetSystemParameters(ctx echo.Context, partyId gensysparams.PartyId) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	parameters, err := c.Service.GetAll(partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	parametersDto := convertParametersToDto(parameters)

	return ctx.JSON(http.StatusOK, parametersDto)
}

// GetSystemParameter (GET /parties/{partyId}/system-parameters/{name})
func (c *Controller) GetSystemParameter(ctx echo.Context, partyId gensysparams.PartyId, name gensysparams.Name) error {
	userId, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.PartyService.RequireMember(userId, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	parameter, err := c.Service.GetParameter(partyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	parameterDto := convertParameterToDto(parameter)

	return ctx.JSON(http.StatusOK, parameterDto)
}

// ChangeSystemParameter (POST /parties/{partyId}/system-parameters/{name})
func (c *Controller) ChangeSystemParameter(ctx echo.Context, partyId gensysparams.PartyId, name gensysparams.Name) error {
	err := common.RequireAdmin(ctx, &c.AuthService, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var parameterDto gensysparams.SystemParameterChange
	err = ctx.Bind(&parameterDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	created, err := c.Service.ChangeValue(partyId, name, parameterDto.Value)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	if created {
		return ctx.NoContent(http.StatusCreated)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// ResetSystemParameter (DELETE /parties/{partyId}/system-parameters/{name})
func (c *Controller) ResetSystemParameter(ctx echo.Context, partyId gensysparams.PartyId, name gensysparams.Name) error {
	err := common.RequireAdmin(ctx, &c.AuthService, partyId)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.ResetParameter(partyId, name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

func convertParameterToDto(parameter typesysparams.SystemParameter) gensysparams.SystemParameter {
	return gensysparams.SystemParameter{
		Name:      parameter.Name,
		Value:     parameter.Value,
		IsDefault: parameter.IsDefault,
	}
}

func convertParametersToDto(parameters []typesysparams.SystemParameter) gensysparams.SystemParameters {
	dtos := make(gensysparams.SystemParameters, len(parameters))

	for i, parameter := range parameters {
		dtos[i] = convertParameterToDto(parameter)
	}

	return dtos
}
