package typehistory

import "time"

type ManualHistoryEntry struct {
	Id          int
	UserId      int
	PartyId     int
	ActorUserId int
	ChangeId    int
	CreatedDate time.Time
}
