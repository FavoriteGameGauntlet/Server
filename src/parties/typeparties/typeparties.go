package typeparties

import "time"

type Party struct {
	Id          int
	Name        string
	CreatedDate time.Time
}

type Member struct {
	Id          int
	UserId      int
	PartyId     int
	DisplayName *string
	IsAdmin     bool
	JoinedDate  time.Time
	LeftDate    *time.Time
}

type MemberWithLogin struct {
	Id          int
	UserId      int
	PartyId     int
	Login       string
	DisplayName *string
	IsAdmin     bool
	JoinedDate  time.Time
	LeftDate    *time.Time
}
