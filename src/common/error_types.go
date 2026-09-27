package common

import (
	"FGG-Service/src/timers/types"
	"fmt"
)

type AppError interface {
	GetCode() string
	GetMessage() string
}

type BaseError struct {
	Code    string
	Message string
}

func (e *BaseError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *BaseError) GetCode() string {
	return e.Code
}

func (e *BaseError) GetMessage() string {
	return e.Message
}

type BadRequestError struct {
	*BaseError
}

func NewBadRequestError(message string) error {
	return &BadRequestError{
		&BaseError{
			Code:    "BAD_REQUEST",
			Message: message,
		},
	}
}

type UnauthorizedError struct {
	*BaseError
}

func NewCookieNotFoundUnauthorizedError() error {
	return &UnauthorizedError{
		&BaseError{
			Code:    "COOKIE_NOT_FOUND",
			Message: "Unable to retrieve the authentication cookie. Try logging in.",
		},
	}
}

func NewActiveSessionNotFoundUnauthorizedError() error {
	return &UnauthorizedError{
		&BaseError{
			Code:    "ACTIVE_SESSION_NOT_FOUND",
			Message: "An authentication session couldn't be found. Try logging in.",
		},
	}
}

func NewNotAdminUnauthorizedError() error {
	return &UnauthorizedError{
		&BaseError{
			Code:    "NOT_ADMIN",
			Message: "This action is only available to administrators.",
		},
	}
}

type NotFoundError struct {
	*BaseError
}

func NewCurrentGameNotFoundError() error {
	return &NotFoundError{
		&BaseError{
			Code:    "CURRENT_GAME_NOT_FOUND",
			Message: "The user doesn't have a current game. Roll the game to get one.",
		},
	}
}

func NewPlayedGameNotFoundError(name string) error {
	message := fmt.Sprintf(
		"The user hasn't played the game \"%s\". Only a finished or cancelled game can be rated.",
		name)

	return &NotFoundError{
		&BaseError{
			Code:    "PLAYED_GAME_NOT_FOUND",
			Message: message,
		},
	}
}

func NewGameReviewNotFoundError(name string) error {
	message := fmt.Sprintf(
		"The user hasn't rated the game \"%s\".",
		name)

	return &NotFoundError{
		&BaseError{
			Code:    "GAME_REVIEW_NOT_FOUND",
			Message: message,
		},
	}
}

func NewCompletedTimersNotFoundError() error {
	return &NotFoundError{
		&BaseError{
			Code:    "COMPLETED_TIMERS_NOT_FOUND",
			Message: "The user doesn't have completed timers. Complete at least one timer to finish the game.",
		},
	}
}

func NewUnplayedGamesNotFoundError(minimum int) error {
	message := fmt.Sprintf(
		"The user doesn't have unplayed games. Add at least %d to roll the game.",
		minimum)

	return &NotFoundError{
		&BaseError{
			Code:    "UNPLAYED_GAMES_NOT_FOUND",
			Message: message,
		},
	}
}

func NewCurrentTimerNotFoundError() error {
	return &NotFoundError{
		&BaseError{
			Code:    "CURRENT_TIMER_NOT_FOUND",
			Message: "The user doesn't have a current timer. Create a timer so you can control it.",
		},
	}
}

func NewAvailableRollsNotFoundError() error {
	return &NotFoundError{
		&BaseError{
			Code:    "AVAILABLE_ROLLS_NOT_FOUND",
			Message: "The user doesn't have available rolls. Complete the timer to get one.",
		},
	}
}

func NewUserLoginNotFoundError(userLogin string) error {
	message := fmt.Sprintf(
		"The user login \"%s\" wasn't found.",
		userLogin)

	return &NotFoundError{
		&BaseError{
			Code:    "USER_LOGIN_NOT_FOUND",
			Message: message,
		},
	}
}

func NewLastWheelEffectsNotFoundError() error {
	return &NotFoundError{
		&BaseError{
			Code:    "LAST_WHEEL_EFFECTS_NOT_FOUND",
			Message: "The last wheel effects not found. Try to roll wheel effects.",
		},
	}
}

func NewDisplayNameNotFoundError() error {
	return &NotFoundError{
		&BaseError{
			Code:    "DISPLAY_NAME_NOT_FOUND",
			Message: "The user have not added a display name.",
		},
	}
}

func NewSystemParameterNotFoundError(name string) error {
	message := fmt.Sprintf(
		"The system parameter \"%s\" wasn't found.",
		name)

	return &NotFoundError{
		&BaseError{
			Code:    "SYSTEM_PARAMETER_NOT_FOUND",
			Message: message,
		},
	}
}

func NewPointTypeNotFoundError(name string) error {
	message := fmt.Sprintf(
		"The point type \"%s\" wasn't found.",
		name)

	return &NotFoundError{
		&BaseError{
			Code:    "POINT_TYPE_NOT_FOUND",
			Message: message,
		},
	}
}

func NewPointTypeAlreadyExistsConflictError(name string) error {
	message := fmt.Sprintf(
		"The point type \"%s\" already exists.",
		name)

	return &ConflictError{
		&BaseError{
			Code:    "POINT_TYPE_ALREADY_EXISTS",
			Message: message,
		},
	}
}

func NewSharedPointTypeConflictError(name string) error {
	message := fmt.Sprintf(
		"The point type \"%s\" is shared by the party. Change it through the party points endpoint.",
		name)

	return &ConflictError{
		&BaseError{
			Code:    "SHARED_POINT_TYPE",
			Message: message,
		},
	}
}

func NewNotSharedPointTypeConflictError(name string) error {
	message := fmt.Sprintf(
		"The point type \"%s\" isn't shared by the party. Change it for a user instead.",
		name)

	return &ConflictError{
		&BaseError{
			Code:    "NOT_SHARED_POINT_TYPE",
			Message: message,
		},
	}
}
func NewItemNotFoundError(name string) error {
	message := fmt.Sprintf(
		"The item \"%s\" wasn't found.",
		name)

	return &NotFoundError{
		&BaseError{
			Code:    "ITEM_NOT_FOUND",
			Message: message,
		},
	}
}

func NewPerkNotFoundError(name string) error {
	message := fmt.Sprintf(
		"The perk \"%s\" wasn't found.",
		name)

	return &NotFoundError{
		&BaseError{
			Code:    "PERK_NOT_FOUND",
			Message: message,
		},
	}
}

func NewEffectNotFoundError(name string) error {
	message := fmt.Sprintf(
		"The effect \"%s\" wasn't found.",
		name)

	return &NotFoundError{
		&BaseError{
			Code:    "EFFECT_NOT_FOUND",
			Message: message,
		},
	}
}

func NewChangeEntryUnprocessableError() error {
	return &UnprocessableError{
		&BaseError{
			Code:    "CHANGE_ENTRY_WITHOUT_TARGET",
			Message: "A change entry has to name exactly one point type, item, perk or effect.",
		},
	}
}
func NewItemNotOwnedConflictError(name string) error {
	message := fmt.Sprintf(
		"The item \"%s\" isn't owned.",
		name)

	return &ConflictError{
		&BaseError{
			Code:    "ITEM_NOT_OWNED",
			Message: message,
		},
	}
}

func NewItemUsedUpConflictError(name string) error {
	message := fmt.Sprintf(
		"The item \"%s\" has no uses left.",
		name)

	return &ConflictError{
		&BaseError{
			Code:    "ITEM_USED_UP",
			Message: message,
		},
	}
}

func NewEffectNotActiveConflictError(name string) error {
	message := fmt.Sprintf(
		"The effect \"%s\" isn't active.",
		name)

	return &ConflictError{
		&BaseError{
			Code:    "EFFECT_NOT_ACTIVE",
			Message: message,
		},
	}
}

func NewEffectUsedUpConflictError(name string) error {
	message := fmt.Sprintf(
		"The effect \"%s\" has no uses left.",
		name)

	return &ConflictError{
		&BaseError{
			Code:    "EFFECT_USED_UP",
			Message: message,
		},
	}
}
func NewPerkNotOwnedConflictError(name string) error {
	message := fmt.Sprintf(
		"The perk \"%s\" isn't owned.",
		name)

	return &ConflictError{
		&BaseError{
			Code:    "PERK_NOT_OWNED",
			Message: message,
		},
	}
}
func NewExchangeNotFoundError(name string) error {
	message := fmt.Sprintf(
		"The exchange \"%s\" wasn't found.",
		name)

	return &NotFoundError{
		&BaseError{
			Code:    "EXCHANGE_NOT_FOUND",
			Message: message,
		},
	}
}

func NewNotEnoughPointsConflictError(pointTypeName string, requiredPoints int) error {
	message := fmt.Sprintf(
		"At least %d of \"%s\" is required.",
		requiredPoints,
		pointTypeName)

	return &ConflictError{
		&BaseError{
			Code:    "NOT_ENOUGH_POINTS",
			Message: message,
		},
	}
}
func NewPartyNotFoundError(partyId int) error {
	message := fmt.Sprintf(
		"The party %d wasn't found.",
		partyId)

	return &NotFoundError{
		&BaseError{
			Code:    "PARTY_NOT_FOUND",
			Message: message,
		},
	}
}

func NewMemberNotFoundError() error {
	return &NotFoundError{
		&BaseError{
			Code:    "MEMBER_NOT_FOUND",
			Message: "This user isn't a member of the party.",
		},
	}
}

func NewMemberAlreadyExistsConflictError() error {
	return &ConflictError{
		&BaseError{
			Code:    "MEMBER_ALREADY_EXISTS",
			Message: "This user is already a member of the party.",
		},
	}
}
func NewWheelEffectNameNotFoundError() error {
	return &NotFoundError{
		&BaseError{
			Code:    "WHEEL_EFFECT_NAME_NOT_FOUND",
			Message: "The wheel effect name not found. Change it or try again later.",
		},
	}
}

type ConflictError struct {
	*BaseError
}

func NewSessionAlreadyExistsConflictError() error {
	return &ConflictError{
		&BaseError{
			Code:    "SESSION_ALREADY_EXISTS",
			Message: "You're already logged in.",
		},
	}
}

func NewCurrentTimerIncorrectStateConflictError(timerState typetimers.TimerStateType) error {
	message := fmt.Sprintf(
		"This action cannot be performed. The current timer is in the \"%s\" state.",
		timerState)

	return &ConflictError{
		&BaseError{
			Code:    "CURRENT_TIMER_INCORRECT_STATE",
			Message: message,
		},
	}
}

func NewWishlistGameAlreadyExistsConflictError(gameName string) error {
	message := fmt.Sprintf(
		"The unplayed game \"%s\" has already been added.",
		gameName)

	return &ConflictError{
		&BaseError{
			Code:    "UNPLAYED_GAME_ALREADY_EXISTS",
			Message: message,
		},
	}
}

func NewCurrentGameAlreadyExistsConflictError() error {
	return &ConflictError{
		&BaseError{
			Code:    "CURRENT_GAME_ALREADY_EXISTS",
			Message: "The current game has already been rolled. Complete it before you can roll a new one.",
		},
	}
}

func NewUserNameAlreadyExistsConflictError() error {
	return &ConflictError{
		&BaseError{
			Code:    "USER_NAME_ALREADY_EXISTS",
			Message: "This username is already taken. Try another one.",
		},
	}
}

func NewUserEmailAlreadyExistsConflictError() error {
	return &ConflictError{
		&BaseError{
			Code:    "USER_EMAIL_ALREADY_EXISTS",
			Message: "This email is already taken. Try another one.",
		},
	}
}

func NewAvailableRollsExistConflictError() error {
	return &ConflictError{
		&BaseError{
			Code:    "AVAILABLE_ROLLS_EXIST",
			Message: "You have available rolls. You need to use them.",
		},
	}
}

func NewNotEnoughAvailableWheelEffectsConflictError() error {
	return &ConflictError{
		&BaseError{
			Code:    "NOT_ENOUGH_AVAILABLE_WHEEL_EFFECTS",
			Message: "You have not enough available wheel effects to roll. Contact with the administrator.",
		},
	}
}

func NewWheelEffectRollAlreadyAppliedConflictError() error {
	return &ConflictError{
		&BaseError{
			Code:    "WHEEL_EFFECT_ROLL_ALREADY_APPLIED",
			Message: "This wheel effect roll has already been applied.",
		},
	}
}

type UnprocessableError struct {
	*BaseError
}

func NewWrongDataUnprocessableError() error {
	return &UnprocessableError{
		&BaseError{
			Code:    "WRONG_AUTH_DATA",
			Message: "Incorrect login or password. Try again.",
		},
	}
}

func NewUserNameUnprocessableError(name string, messageDetails string) error {
	message := fmt.Sprintf(
		"'%s' does not match the format. %s",
		name,
		messageDetails)

	return &UnprocessableError{
		&BaseError{
			Code:    "INCORRECT_USER_NAME_FORMAT",
			Message: message,
		},
	}
}

func NewNameUnprocessableError(name string, messageDetails string) error {
	message := fmt.Sprintf(
		"'%s' does not match the format. %s",
		name,
		messageDetails)

	return &UnprocessableError{
		&BaseError{
			Code:    "INCORRECT_GAME_NAME_FORMAT",
			Message: message,
		},
	}
}

func NewRatingUnprocessableError(rating int, messageDetails string) error {
	message := fmt.Sprintf(
		"'%d' does not match the format. %s",
		rating,
		messageDetails)

	return &UnprocessableError{
		&BaseError{
			Code:    "INCORRECT_RATING_FORMAT",
			Message: message,
		},
	}
}

func NewEmailUnprocessableError(email string, messageDetails string) error {
	message := fmt.Sprintf(
		"'%s' does not match the format. %s",
		email,
		messageDetails)

	return &UnprocessableError{
		&BaseError{
			Code:    "INCORRECT_EMAIL_FORMAT",
			Message: message,
		},
	}
}

func NewPasswordUnprocessableError(messageDetails string) error {
	message := fmt.Sprintf(
		"The password does not match the format. %s",
		messageDetails)

	return &UnprocessableError{
		&BaseError{
			Code:    "INCORRECT_PASSWORD_FORMAT",
			Message: message,
		},
	}
}

