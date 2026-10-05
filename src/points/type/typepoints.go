package typepoints

import "time"

type PointType struct {
	Id          int
	PartyId     int
	Name        string
	Description string
	StartValue  int
	IsPublic    bool
	IsShared    bool
	Minimum     *int
	Maximum     *int
	IsRemoved   bool
}

type PointTypeInfo struct {
	Id          int
	PartyId     int
	Name        string
	Description string
	StartValue  int
	IsPublic    bool
	IsShared    bool
	Minimum     *int
	Maximum     *int
}

type UserPoint struct {
	Id          int
	UserId      int
	PartyId     int
	PointTypeId int
	Value       int
}

type PartyPoint struct {
	Id          int
	PartyId     int
	PointTypeId int
	Value       int
}

type UserPointHistoryEntry struct {
	Id                 int
	UserId             int
	PartyId            int
	PointTypeId        int
	ActorUserId        int
	DesiredChangeValue int
	ActualChangeValue  int
	FinalValue         int
	SourceEventId      int
	ChangedDate        time.Time
}

type PartyPointHistoryEntry struct {
	Id                 int
	PartyId            int
	PointTypeId        int
	ActorUserId        int
	DesiredChangeValue int
	ActualChangeValue  int
	FinalValue         int
	SourceEventId      int
	ChangedDate        time.Time
}

// PointChangeResult is the outcome of a point change once clamped to its point type's bounds.
type PointChangeResult struct {
	DesiredChangeValue int
	ActualChangeValue  int
	FinalValue         int
}

// PointValue pairs a point type with the value held for it, for reads that list every type.
type PointValue struct {
	PointType PointTypeInfo
	Value     int
}

// UserPointValuesByLogin is one member's point values, for reads across the whole party.
type UserPointValuesByLogin struct {
	Login  string
	Points []PointValue
}

// PointHistoryEntry is one recorded change to a point value, whether user- or party-scoped.
type PointHistoryEntry struct {
	DesiredChangeValue int
	ActualChangeValue  int
	FinalValue         int
	ActorUserId        int
	ChangedDate        time.Time
}
