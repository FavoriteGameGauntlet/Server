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
	ActorUserId   int
	PartyId       int
	PerkId        int
	Action        string
	SourceEventId *int
	CreatedDate   time.Time
}

// UserPerkDetail is a perk a user holds, carrying the catalogue details the API returns with it.
type UserPerkDetail struct {
	Name         string
	Description  string
	ReceivedDate time.Time
}

// PerkHistoryEntry is one recorded perk event, named by the perk it concerns.
type PerkHistoryEntry struct {
	Name        string
	Action      string
	ActorUserId int
	CreatedDate time.Time
}
