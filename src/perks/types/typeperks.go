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
