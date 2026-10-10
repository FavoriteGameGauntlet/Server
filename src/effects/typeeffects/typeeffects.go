package typeeffects

import (
	"FGG-Service/src/changes/typechanges"
	"time"
)

// Effect is a catalogue entry. A nil UseCount stands for an effect that can be used without limit.
type Effect struct {
	Id          int
	PartyId     int
	Name        string
	Description string
	UseCount    *int
	Duration    *time.Duration
	Modifiers   []PointModifier
}

type EffectWithChange struct {
	Id          int
	PartyId     int
	Name        string
	Description string
	UseCount    *int
	Duration    *time.Duration
	Change      typechanges.Change
}

type PointModifier struct {
	Id          int `json:"id"`
	PointTypeId int `json:"point_type_id"`
	Amount      int `json:"amount"`
}

// UserEffect is an effect held by a user. A nil UsesLeft stands for an effect without a use limit.
type UserEffect struct {
	Id              int
	UserId          int
	PartyId         int
	EffectId        int
	UsesLeft        *int
	EffectHistoryId int
}

type UserEffectDetail struct {
	Id          int
	UserId      int
	PartyId     int
	EffectId    int
	Name        string
	Description string
	UseCount    *int
	UsesLeft    *int
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
	UseCount      *int
	Duration      *time.Duration
	Action        string
	UsesLeft      *int
	SourceEventId *int
	CreatedDate   time.Time
	// Entries is filled by the service, since the schema does not return it with the history rows.
	Entries []typechanges.NamedChangeEntry
}

type EndedUserEffect struct {
	Id       int
	UserId   int
	PartyId  int
	EffectId int
	UsesLeft *int
}

// NamedPointModifier is a passive point modifier that names its point type the way the API does,
// instead of carrying its id.
type NamedPointModifier struct {
	PointTypeName string
	Amount        int
}

// NamedEffect is a catalogue effect, with its passive modifiers named and the entries using it grants.
type NamedEffect struct {
	Name        string
	Description string
	UseCount    *int
	Duration    *time.Duration
	Modifiers   []NamedPointModifier
	Entries     []typechanges.NamedChangeEntry
}

// NamedUserEffect is an active effect of a user, with its passive modifiers named and the entries
// using it grants.
type NamedUserEffect struct {
	Name        string
	Description string
	UseCount    *int
	UsesLeft    *int
	Duration    *time.Duration
	StartedDate time.Time
	Modifiers   []NamedPointModifier
	Entries     []typechanges.NamedChangeEntry
}
