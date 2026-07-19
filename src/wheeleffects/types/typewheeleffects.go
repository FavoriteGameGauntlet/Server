package typewheeleffects

import (
	typepoints "FGG-Service/src/points/type"
	"time"
)

type WheelEffect struct {
	Id          int
	Name        string
	Description *string
}

type WheelEffects = []WheelEffect

type RolledWheelEffect struct {
	Id          int
	Name        string
	Description *string
	RollDate    time.Time
	Position    int
	IsApplied   bool
}

type RolledWheelEffects = []RolledWheelEffect

type RolledWheelEffectHistory struct {
	Id          int
	Name        string
	Description *string
	RollDate    time.Time
}

type RolledWheelEffectHistories = []RolledWheelEffectHistory

type WheelEffectRollApply struct {
	PointChangeByUserIds typepoints.PointChangeByUserIds
	WheelEffectName      string
}

type WheelGroup struct {
	Id      int
	PartyId int
	Name    string
}

type CreatedWheelRow struct {
	Id          int
	PartyId     int
	Name        string
	Description string
	ChangeId    int
	GroupId     int
}

type WheelRow struct {
	Id             int
	PartyId        int
	Name           string
	Description    string
	ChangeId       int
	GroupId        int
	IsManualChange bool
}

type RolledWheelRow struct {
	Id             int
	UserId         int
	PartyId        int
	WheelRowId     int
	WheelPosition  int
	RolledDate     time.Time
	IsManualChange bool
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
