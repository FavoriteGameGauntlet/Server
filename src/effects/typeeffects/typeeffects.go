package typeeffects

import (
	"FGG-Service/src/changes/typechanges"
	"time"
)

type Effect struct {
	Id          int
	PartyId     int
	Name        string
	Description string
	UseCount    int
	Duration    *time.Duration
}

type EffectWithChange struct {
	Id          int
	PartyId     int
	Name        string
	Description string
	UseCount    int
	Duration    *time.Duration
	Change      typechanges.Change
}

type PointModifier struct {
	Id          int `json:"id"`
	PointTypeId int `json:"point_type_id"`
	Amount      int `json:"amount"`
}

type UserEffect struct {
	Id              int
	UserId          int
	PartyId         int
	EffectId        int
	UsesLeft        int
	EffectHistoryId int
}

type UserEffectDetail struct {
	Id          int
	UserId      int
	PartyId     int
	EffectId    int
	Name        string
	Description string
	UseCount    int
	UsesLeft    int
	Duration    *time.Duration
	StartedDate time.Time
	Modifiers   []PointModifier
}

type EffectHistory struct {
	Id            int
	UserId        int
	ActorUserId   int
	PartyId       int
	EffectId      int
	Name          string
	Description   string
	UseCount      int
	Duration      *time.Duration
	Action        string
	UsesLeft      *int
	SourceEventId *int
	CreatedDate   time.Time
}

type EndedUserEffect struct {
	Id       int
	UserId   int
	PartyId  int
	EffectId int
	UsesLeft int
}

// PointModifierView is a passive point modifier named the way the API returns it.
type PointModifierView struct {
	PointTypeName string
	Amount        int
}

// UserEffectView is an active effect of a user, with its passive modifiers named.
type UserEffectView struct {
	Name        string
	Description string
	UseCount    int
	UsesLeft    int
	Duration    *time.Duration
	StartedDate time.Time
	Modifiers   []PointModifierView
}
