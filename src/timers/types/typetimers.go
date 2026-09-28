package typetimers

import (
	"FGG-Service/src/changes/types"
	"time"
)

type Timer struct {
	Duration       time.Duration
	Id             int
	RemainingTime  time.Duration
	State          TimerStateType
	LastActionDate time.Time
}

type CurrentTimer struct {
	Id             int
	GameId         int
	State          TimerStateType
	Duration       time.Duration
	LastActionDate time.Time
	TimeSpent      time.Duration
}

type CreatedTimer struct {
	Id          int
	UserId      int
	PartyId     int
	GameId      int
	State       TimerStateType
	Duration    time.Duration
	CreatedDate time.Time
}

type EndedTimer struct {
	Id             int
	UserId         int
	PartyId        int
	GameId         int
	State          TimerStateType
	Duration       time.Duration
	TimeSpent      time.Duration
	LastActionDate time.Time
}

// TimerReward is what the party grants for every completed timer.
type TimerReward struct {
	Id     int
	Change typechanges.Change
}

type TimerHistoryEntry struct {
	Id            int
	UserId        int
	PartyId       int
	TimerRewardId *int
	CompletedDate time.Time
}

type TimerStateType string

const (
	TimerStateCreated  TimerStateType = "created"
	TimerStateFinished TimerStateType = "finished"
	TimerStatePaused   TimerStateType = "paused"
	TimerStateRunning  TimerStateType = "running"
)
