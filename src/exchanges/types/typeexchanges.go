package typeexchanges

import (
	"FGG-Service/src/changes/types"
	"time"
)

type Exchange struct {
	Id          int
	PartyId     int
	Name        string
	Description string
}

type ExchangeWithChanges struct {
	Id           int
	Name         string
	Description  string
	SourceChange typechanges.Change
	TargetChange typechanges.Change
}

type ExchangeHistoryEntry struct {
	Id         int
	UserId     int
	PartyId    int
	ExchangeId int
	UsedDate   time.Time
}

type ExchangeHistory struct {
	Id          int
	UserId      int
	PartyId     int
	ExchangeId  int
	Name        string
	Description string
	UsedDate    time.Time
}
