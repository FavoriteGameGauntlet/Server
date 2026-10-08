package typeauth

import "time"

type User struct {
	Id          int
	Login       string
	DisplayName *string
	Email       string
}

type UserSession struct {
	Id          string
	UserId      int
	CreatedDate time.Time
	ExpiryDate  time.Time
}

type Password struct {
	Value string
}

type LoginUser struct {
	Login    string
	Password Password
}

type SignupUser struct {
	Email    string
	Login    string
	Password Password
}
