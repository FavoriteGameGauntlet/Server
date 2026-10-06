package typewheeleffects

import "time"

type WheelCollection struct {
	Id                 int
	PartyId            int
	Name               string
	ShouldCheckHistory bool
}

type WheelRow struct {
	Id             int
	PartyId        int
	Name           string
	Description    string
	ChangeId       int
	CollectionId   int
	IsManualChange bool
}

type CreatedWheelRow struct {
	Id           int
	PartyId      int
	Name         string
	Description  string
	ChangeId     int
	CollectionId int
}

// RolledWheelRowInput is one row to persist as a rolled-but-not-yet-applied pick (see
// AddLastRolledWheelEffectsCommand / create_last_wheel_rows), at the given position in the roll batch.
type RolledWheelRowInput struct {
	WheelRowId int
	Position   int
}

type CreatedLastWheelRow struct {
	Id            int
	UserId        int
	PartyId       int
	WheelRowId    int
	WheelPosition int
	RolledDate    time.Time
}

type LastWheelRow struct {
	Id            int
	Name          string
	Description   string
	RolledDate    time.Time
	WheelPosition int
}

type CreatedWheelRowHistory struct {
	Id          int
	UserId      int
	PartyId     int
	WheelRowId  int
	AppliedDate time.Time
}

type WheelRowHistory struct {
	Name        string
	Description string
	AppliedDate time.Time
}
