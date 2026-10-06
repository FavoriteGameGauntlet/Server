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

type UserItem struct {
	Id           int
	UserId       int
	PartyId      int
	ItemId       int
	UsesLeft     int
	ReceivedDate time.Time
}

type ItemHistory struct {
	Id            int
	UserId        int
	ActorUserId   int
	PartyId       int
	ItemId        int
	Name          string
	UseCount      int
	Action        string
	UsesLeft      *int
	SourceEventId *int
	CreatedDate   time.Time
}

// UserItemDetail is an item a user holds, carrying the catalog details the API returns with it.
type UserItemDetail struct {
	Name         string
	Description  string
	UsesLeft     int
	ReceivedDate time.Time
}
