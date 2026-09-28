package validator

import (
	"FGG-Service/src/common"
	"fmt"
	"math"
	"regexp"
)

var userNameRegex = regexp.MustCompile(`^\w+$`)

func ValidateUserLogin(login string) error {
	if len(login) < 3 {
		return common.NewUserNameUnprocessableError(
			login,
			"The login is too short, it should be at least 3 characters long.")
	}

	if len(login) > 35 {
		return common.NewUserNameUnprocessableError(
			login,
			"The login is too long, it should be less than 35 characters long.")
	}

	if !userNameRegex.MatchString(login) {
		return common.NewUserNameUnprocessableError(
			login,
			"The login must contain only Latin letters, numbers, and underscores")
	}

	return nil
}

var nameRegex = regexp.MustCompile(`^.+$`)

func ValidateName(name string) error {
	if len(name) < 1 {
		return common.NewNameUnprocessableError(
			name,
			"The name is too short, it should be at least 1 characters long.")
	}

	if len(name) > 70 {
		return common.NewNameUnprocessableError(
			name,
			"The name is too long, it should be less than 70 characters long.")
	}

	if !nameRegex.MatchString(name) {
		return common.NewNameUnprocessableError(
			name,
			"The name must contain only UTF-8 characters")
	}

	return nil
}

func ValidateRating(rating int) error {
	if rating < 1 || rating > 10 {
		return common.NewRatingUnprocessableError(
			rating,
			"The rating should be between 1 and 10.")
	}

	return nil
}

// ValidatePointTypeBounds checks what the schema requires of a point type: every value fits the
// INTEGER columns, the minimum does not exceed the maximum and the start value lies between them.
// A nil bound does not limit.
func ValidatePointTypeBounds(startValue int, minimum *int, maximum *int) error {
	values := []struct {
		name  string
		value *int
	}{
		{"start value", &startValue},
		{"minimum", minimum},
		{"maximum", maximum},
	}

	for _, v := range values {
		if v.value != nil && !fitsInteger(*v.value) {
			return common.NewPointTypeBoundsUnprocessableError(fmt.Sprintf(
				"The %s (%d) should be between %d and %d.", v.name, *v.value, math.MinInt32, math.MaxInt32))
		}
	}

	if minimum != nil && maximum != nil && *minimum > *maximum {
		return common.NewPointTypeBoundsUnprocessableError(fmt.Sprintf(
			"The minimum (%d) should not be greater than the maximum (%d).", *minimum, *maximum))
	}

	if minimum != nil && startValue < *minimum {
		return common.NewPointTypeBoundsUnprocessableError(fmt.Sprintf(
			"The start value (%d) should not be less than the minimum (%d).", startValue, *minimum))
	}

	if maximum != nil && startValue > *maximum {
		return common.NewPointTypeBoundsUnprocessableError(fmt.Sprintf(
			"The start value (%d) should not be greater than the maximum (%d).", startValue, *maximum))
	}

	return nil
}

// ValidateChangeAmount checks that a change to a point value fits the INTEGER columns it is stored
// and recorded in.
func ValidateChangeAmount(amount int) error {
	if !fitsInteger(amount) {
		return common.NewChangeAmountUnprocessableError(amount, math.MinInt32, math.MaxInt32)
	}

	return nil
}

// fitsInteger reports whether a value fits a Postgres INTEGER column.
func fitsInteger(value int) bool {
	return value >= math.MinInt32 && value <= math.MaxInt32
}

var emailRegex = regexp.MustCompile(`^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$`)

func ValidateEmail(email string) error {
	if len(email) < 6 {
		return common.NewEmailUnprocessableError(
			email,
			"The email is too short, it should be at least 6 characters long.")
	}

	if len(email) > 100 {
		return common.NewEmailUnprocessableError(
			email,
			"The email is too long, it should be less than 100 characters long.")
	}

	if !emailRegex.MatchString(email) {
		return common.NewEmailUnprocessableError(
			email,
			"The email must follow the pattern (e.g. email@example.com)")
	}

	return nil
}

var passwordLetterRegex = regexp.MustCompile(`^[\w\d !"#$%&'()*+,-.=/:;<?>@\[\\\]^_|{}~]+$`)

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return common.NewPasswordUnprocessableError(
			"The password is too short, it should be at least 8 characters long.")
	}

	if len(password) > 35 {
		return common.NewPasswordUnprocessableError(
			"The password is too long, it should be less than 35 characters long.")
	}

	if !passwordLetterRegex.MatchString(password) {
		return common.NewPasswordUnprocessableError(
			"The password must contain only Latin letters, numbers, and special symbols ( !\"#$%&'()*+,-./:;<=>?@[]\\^_{|}~).")
	}

	return nil
}
