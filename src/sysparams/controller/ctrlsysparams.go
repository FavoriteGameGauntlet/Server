package ctrlsysparams

import (
	"FGG-Service/api/generated/system_parameters"
	"FGG-Service/src/auth/service"
	"FGG-Service/src/common"
	"FGG-Service/src/sysparams/service"
	"FGG-Service/src/sysparams/types"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Controller struct {
	Service     srvsysparams.Service
	AuthService srvauth.Service
}

func NewController() *Controller {
	s := srvsysparams.NewService()
	as := srvauth.NewService()

	return &Controller{
		*s,
		*as,
	}
}

// GetSystemParameters (GET /system-parameters)
func (c *Controller) GetSystemParameters(ctx echo.Context) error {
	_, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	parameters, err := c.Service.GetAll()

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	parametersDto := convertParametersToDto(parameters)

	return ctx.JSON(http.StatusOK, parametersDto)
}

// GetSystemParameter (GET /system-parameters/{name})
func (c *Controller) GetSystemParameter(ctx echo.Context, name gensysparams.Name) error {
	_, err := c.AuthService.GetUserId(ctx)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	parameter, err := c.Service.GetParameter(name)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	parameterDto := convertParameterToDto(parameter)

	return ctx.JSON(http.StatusOK, parameterDto)
}

// ChangeSystemParameter (POST /system-parameters/{name})
func (c *Controller) ChangeSystemParameter(ctx echo.Context, name gensysparams.Name) error {
	err := common.RequireAdmin(ctx, &c.AuthService)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	var parameterDto gensysparams.SystemParameterChange
	err = ctx.Bind(&parameterDto)

	if err != nil {
		err = common.NewBadRequestError(err.Error())
		return common.SendJSONErrorResponse(ctx, err)
	}

	created, err := c.Service.ChangeValue(name, parameterDto.Value)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	if created {
		return ctx.NoContent(http.StatusCreated)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// ResetSystemParameter (DELETE /system-parameters/{name})
func (c *Controller) ResetSystemParameter(ctx echo.Context, name gensysparams.Name) error {
	err := common.RequireAdmin(ctx, &c.AuthService)

	if err != nil {
		return common.SendJSONErrorResponse(ctx, err)
	}

	err = c.Service.ResetParameter(name)

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
