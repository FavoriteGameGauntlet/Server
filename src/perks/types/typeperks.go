package typeperks

import "time"

type Perk struct {
	Id          int
	PartyId     int
	Name        string
	Description string
	EffectId    int
}

type PerkWithRemoved struct {
	Id          int
	PartyId     int
	Name        string
	Description string
	EffectId    int
	IsRemoved   bool
}

type UserPerk struct {
	Id           int
	UserId       int
	PartyId      int
	PerkId       int
	UserEffectId int
	ReceivedDate time.Time
}

type PerkHistory struct {
	Id            int
	UserId        int
	PartyId       int
	PerkId        int
	Action        string
	SourceEventId *int
	CreatedDate   time.Time
}

// UserPerkView is a perk a user holds, carrying the catalogue details the API returns with it.
type UserPerkView struct {
	Name         string
	Description  string
	ReceivedDate time.Time
}

// PerkHistoryView is one recorded perk event, named by the perk it concerns.
type PerkHistoryView struct {
	Name        string
	Action      string
	CreatedDate time.Time
}