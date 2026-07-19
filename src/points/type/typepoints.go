package typepoints

import "time"

type PointChangeResult struct {
	ActualChangeValue  int
	ChangeSource       string
	DesiredChangeValue int
	FinalValue         int
}

type PointChangeResultByTypes = map[string]PointChangeResult

type PointChangeResultByUserId struct {
	Login         string
	UserId        int
	ChangeResults PointChangeResultByTypes
}

type PointChangeResultByUserIds = []PointChangeResultByUserId

type PointChange struct {
	ChangeSource       string
	DesiredChangeValue int
}

type TerritoryHourChange struct {
	ChangeSource       string
	DesiredChangeValue int
	IsSomeones         bool
}

type FreePointChange struct {
	SourceUserId       int
	ChangeSource       string
	DesiredChangeValue int
}

type PointChangeByUserId struct {
	Login               string
	UserId              int
	FreePointChange     *FreePointChange
	AvailableRollChange *PointChange
}

type PointChangeByUserIds = []PointChangeByUserId

type TerritoryPointChange struct {
	SourceUserId       int
	ChangeSource       string
	DesiredChangeValue int
}

type FreePointChangeHistory struct {
	ActualChangeValue  int
	ChangeDate         time.Time
	ChangeSource       string
	DesiredChangeValue int
	FinalValue         int
	SourceLogin        *string
	WheelEffectName    *string
}

type FreePointChangeHistories = []FreePointChangeHistory

type TerritoryPointChangeHistory struct {
	ActualChangeValue  int
	ChangeDate         time.Time
	ChangeSource       string
	DesiredChangeValue int
	FinalValue         int
	SourceLogin        *string
}

type TerritoryPointChangeHistories = []TerritoryPointChangeHistory

type PointInfo struct {
	TerritoryPoints int
	FreePoints      int
}

type PointInfoByLogin struct {
	Login     string
	PointInfo PointInfo
}

type PointInfoByLogins = []PointInfoByLogin

const (
	PointTypeTerritoryHours  = "territoryHours"
	PointTypeTerritoryPoints = "territoryPoints"
	PointTypeFreePoints      = "freePoints"
	PointTypeAvailableRolls  = "availableRolls"
)

const (
	ExperienceChangeSourceLevelUp = "level-up"
	ExperienceChangeSourceOther   = "other"

	TerritoryHourChangeSourceSeize = "seize"
	TerritoryHourChangeSourceOther = "other"

	FreePointsChangeSourceQuestCompletion = "quest"
	FreePointsChangeSourceWheelEffect     = "wheel-effect"
	FreePointsChangeSourceBaseTeleport    = "base-teleport"
	FreePointsChangeSourceSandStorm       = "sandstorm"
	FreePointsChangeSourceOther           = "other"

	TerritoryPointChangeSourceObtaining = "territory-obtaining"
	TerritoryPointChangeSourceLoss      = "territory-loss"
	TerritoryPointChangeSourceOther     = "other"
)

var ExperienceChangeSourceSlice = []string{ExperienceChangeSourceLevelUp, ExperienceChangeSourceOther}

var TerritoryHourChangeSourceSlice = []string{TerritoryHourChangeSourceSeize, TerritoryHourChangeSourceOther}

var FreePointsChangeSourceSlice = []string{
	FreePointsChangeSourceQuestCompletion,
	FreePointsChangeSourceWheelEffect,
	FreePointsChangeSourceBaseTeleport,
	FreePointsChangeSourceSandStorm,
	FreePointsChangeSourceOther,
}

var TerritoryPointChangeSourceSlice = []string{
	TerritoryPointChangeSourceObtaining,
	TerritoryPointChangeSourceLoss,
	TerritoryPointChangeSourceOther,
}

type PointType struct {
	Id          int
	PartyId     int
	Name        string
	Description string
	StartValue  int
	IsPublic    bool
	IsShared    bool
	Minimum     int
	Maximum     int
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
	Minimum     int
	Maximum     int
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
	SourceUserId       int
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
	SourceUserId       int
	DesiredChangeValue int
	ActualChangeValue  int
	FinalValue         int
	SourceEventId      int
	ChangedDate        time.Time
}
