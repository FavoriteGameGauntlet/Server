package common

import (
	"FGG-Service/api/generated/auth"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func SendJSONErrorResponse(ctx echo.Context, err error) error {
	var badRequestError *BadRequestError
	var unauthorizedError *UnauthorizedError
	var notFoundError *NotFoundError
	var conflictError *ConflictError
	var unprocessableError *UnprocessableError

	apiCode := http.StatusInternalServerError

	switch {
	case errors.As(err, &badRequestError):
		apiCode = http.StatusBadRequest
	case errors.As(err, &unauthorizedError):
		apiCode = http.StatusUnauthorized
	case errors.As(err, &notFoundError):
		apiCode = http.StatusNotFound
	case errors.As(err, &conflictError):
		apiCode = http.StatusConflict
	case errors.As(err, &unprocessableError):
		apiCode = http.StatusUnprocessableEntity
	}

	apiError := convertToError(err)

	return ctx.JSON(apiCode, apiError)
}

func convertToError(err error) genauth.Error {
	var appError AppError
	if errors.As(err, &appError) {
		return genauth.Error{
			Code:    appError.GetCode(),
			Message: appError.GetMessage(),
		}
	}

	return genauth.Error{
		Code:    "UNEXPECTED",
		Message: err.Error(),
	}
}

func DurationToISO8601(duration time.Duration) string {
	if duration == 0 {
		return "PT0S"
	}

	sign := ""
	if duration < 0 {
		sign = "-"
		duration = -duration
	}

	totalSeconds := int64(duration.Seconds())
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	result := fmt.Sprintf("%sPT", sign)

	if hours > 0 {
		result += fmt.Sprintf("%dH", hours)
	}
	if minutes > 0 {
		result += fmt.Sprintf("%dM", minutes)
	}
	if seconds > 0 {
		result += fmt.Sprintf("%dS", seconds)
	}

	return result
}

func ConvertIntSliceToString(intSlice []int) string {
	stringSlice := make([]string, len(intSlice))
	for i, v := range intSlice {
		stringSlice[i] = strconv.Itoa(v)
	}

	return strings.Join(stringSlice, ", ")
}

// AdminChecker is the part of the auth service needed to authorize admin-only endpoints. It is
// declared here instead of imported so that common keeps no dependency on the auth service.
type AdminChecker interface {
	GetUserId(ctx echo.Context) (userId int, err error)
	IsAdmin(userId int) (isAdmin bool, err error)
}

// RequireAdmin returns an error unless the request comes from an admin of the party.
func RequireAdmin(ctx echo.Context, authService AdminChecker) error {
	userId, err := authService.GetUserId(ctx)

	if err != nil {
		return err
	}

	isAdmin, err := authService.IsAdmin(userId)

	if err != nil {
		return err
	}

	if !isAdmin {
		return NewNotAdminUnauthorizedError()
	}

	return nil
}
