package typeitems

import (
	"FGG-Service/src/changes/typechanges"
	"time"
)

type Item struct {
	Id          int
	PartyId     int
	Name        string
	Description string
	UseCount    int
	ChangeId    int
}

type ItemWithChange struct {
	Id          int
	PartyId     int
	Name        string
	Description string
	UseCount    int
	Change      typechanges.Change
}

// ItemWithEntries is a catalogue item along with what using it grants.
type ItemWithEntries struct {
	Item
	Entries []typechanges.NamedChangeEntry
}

// UserItem is one copy of an item a user holds. Id identifies the copy, ItemId the catalogue item it
// is a copy of.
type UserItem struct {
	Id           int
	UserId       int
	PartyId      int
	ItemId       int
	UsesLeft     int
	ReceivedDate time.Time
}

// UserItemWithChangeId is a copy of an item a user holds, along with the change using it grants. The
// change belongs to the catalogue item, which may have been removed since the user received the copy.
type UserItemWithChangeId struct {
	UserItem
	ChangeId int
}

type ItemHistory struct {
	Id            int
	UserId        int
	ActorUserId   int
	PartyId       int
	ItemId        int
	UserItemId    int
	Name          string
	UseCount      int
	Action        string
	UsesLeft      *int
	SourceEventId *int
	CreatedDate   time.Time
}

// UserItemDetail is a copy of an item a user holds, carrying the catalog details the API returns with
// it. ChangeId points at what using the item grants, which the service reads into Entries.
type UserItemDetail struct {
	Id           int
	ItemId       int
	Name         string
	Description  string
	UsesLeft     int
	ReceivedDate time.Time
	ChangeId     int
	Entries      []typechanges.NamedChangeEntry
}
