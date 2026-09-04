package typegames

import "time"

type CurrentGame struct {
	Id         int
	Name       string
	State      CurrentGameState
	TimeSpent  time.Duration
	StartDate  time.Time
	FinishDate *time.Time
}

type CurrentGames = []CurrentGame

type CurrentGameWithLogin struct {
	Login string
	Game  CurrentGame
}

type CurrentGameState string

const (
	GameStateCancelled CurrentGameState = "cancelled"
	GameStateFinished  CurrentGameState = "finished"
	GameStateStarted   CurrentGameState = "started"
)

type WishlistGame struct {
	Id     int
	GameId int
	Name   string
}

type WishlistGames = []WishlistGame

type CreatedWishlistGame struct {
	Id          int
	UserId      int
	PartyId     int
	GameId      int
	CreatedDate time.Time
}

type Game struct {
	Id      int
	PartyId int
	Name    string
}

type CreatedUserGame struct {
	Id          int
	UserId      int
	PartyId     int
	GameId      int
	TimeSpent   time.Duration
	StartedDate time.Time
}

type UserGame struct {
	Id        int
	Name      string
	TimeSpent time.Duration
}

type UserGameWithLogin struct {
	Id        int
	Name      string
	TimeSpent time.Duration
	Login     string
}

type GameHistoryEntry struct {
	Id            int
	GameId        int
	Name          string
	Action        string
	TimeSpent     time.Duration
	EndState      *string
	SourceEventId *int
	CreatedDate   time.Time
}

type GameRating struct {
	Id            int
	UserId        int
	PartyId       int
	GameId        int
	Rating        int
	ReviewComment *string
	CreatedDate   time.Time
	UpdatedDate   time.Time
}

type GameReview struct {
	Rating        int
	ReviewComment *string
}
