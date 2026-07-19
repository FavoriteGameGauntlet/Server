package typeitems

import (
	"FGG-Service/src/changes/types"
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

type ItemHistoryEntry struct {
	Id       int
	UserId   int
	PartyId  int
	ItemId   int
	UsesLeft int
	UsedDate time.Time
}

type ItemHistory struct {
	Id            int
	UserId        int
	PartyId       int
	ItemId        int
	Action        string
	UsesLeft      int
	SourceEventId *int
	CreatedDate   time.Time
}
