package validator

import (
	"FGG-Service/src/common"
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
