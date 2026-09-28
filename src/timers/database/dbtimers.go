package dbtimers

import (
	"FGG-Service/src/dbaccess"
	"FGG-Service/src/timers/types"
	"time"
)

type IDatabase interface {
	GetCurrentTimerCommand(userId int, partyId int) (timer typetimers.CurrentTimer, err error)
	CreateCurrentTimerCommand(userId int, partyId int, gameId int, duration time.Duration) (timer typetimers.CreatedTimer, err error)
	ActTimerCommand(timerId int, timerState typetimers.TimerStateType, timeSpent time.Duration) error
	GetCompletedTimerUsersCommand() (timers []typetimers.EndedTimer, err error)
	DeleteCurrentTimerCommand(userId int, partyId int) (timer typetimers.EndedTimer, err error)
}

type Database struct {
}

var getCurrentTimerQuery = dbaccess.Query{Name: "GetCurrentTimerQuery", SQL: `SELECT * FROM get_timer($1::integer, $2::integer)`}

func (db *Database) GetCurrentTimerCommand(userId int, partyId int) (timer typetimers.CurrentTimer, err error) {
	row := dbaccess.QueryRow(getCurrentTimerQuery, userId, partyId)

	var durationRaw, timeSpentRaw string
	err = row.Scan(&timer.Id, &timer.GameId, &timer.State, &durationRaw, &timer.LastActionDate, &timeSpentRaw)

	if err == nil {
		timer.Duration, err = dbaccess.ScanInterval(durationRaw)
	}
	if err == nil {
		timer.TimeSpent, err = dbaccess.ScanInterval(timeSpentRaw)
	}

	dbaccess.LogDbResult(getCurrentTimerQuery, timer, err)

	return
}

var createCurrentTimerQuery = dbaccess.Query{Name: "CreateCurrentTimerQuery", SQL: `SELECT * FROM create_timer($1::integer, $2::integer, $3::integer, $4::interval)`}

func (db *Database) CreateCurrentTimerCommand(userId int, partyId int, gameId int, duration time.Duration) (timer typetimers.CreatedTimer, err error) {
	row := dbaccess.QueryRow(createCurrentTimerQuery, userId, partyId, gameId, duration)

	var durationRaw string
	err = row.Scan(&timer.Id, &timer.UserId, &timer.PartyId, &timer.GameId, &timer.State, &durationRaw, &timer.CreatedDate)

	if err == nil {
		timer.Duration, err = dbaccess.ScanInterval(durationRaw)
	}

	dbaccess.LogDbResult(createCurrentTimerQuery, timer, err)

	return
}

var actTimerQuery = dbaccess.Query{Name: "ActTimerQuery", SQL: `SELECT act_timer($1::integer, $2::text, $3::interval)`}

func (db *Database) ActTimerCommand(
	timerId int,
	timerState typetimers.TimerStateType,
	timeSpent time.Duration) error {

	_, err := dbaccess.Exec(
		actTimerQuery,
		timerId,
		timerState,
		timeSpent)

	dbaccess.LogDbResult(actTimerQuery, nil, err)

	return err
}

var getCompletedTimerUsersQuery = dbaccess.Query{Name: "GetCompletedTimerUsersQuery", SQL: `SELECT * FROM delete_ended_timers()`, IsSilent: true}

func (db *Database) GetCompletedTimerUsersCommand() (timers []typetimers.EndedTimer, err error) {
	rows, err := dbaccess.QueryRows(getCompletedTimerUsersQuery)

	if err != nil {
		return
	}

	for rows.Next() {
		timer := typetimers.EndedTimer{}
		var durationRaw, timeSpentRaw string
		err = rows.Scan(
			&timer.Id,
			&timer.UserId,
			&timer.PartyId,
			&timer.GameId,
			&timer.State,
			&durationRaw,
			&timeSpentRaw,
			&timer.LastActionDate)

		if err == nil {
			timer.Duration, err = dbaccess.ScanInterval(durationRaw)
		}
		if err == nil {
			timer.TimeSpent, err = dbaccess.ScanInterval(timeSpentRaw)
		}

		if err != nil {
			_ = rows.Close()
			return
		}

		timers = append(timers, timer)
	}

	dbaccess.LogDbResult(getCompletedTimerUsersQuery, timers, err)

	_ = rows.Close()
	return
}

var deleteCurrentTimerQuery = dbaccess.Query{Name: "DeleteCurrentTimerQuery", SQL: `SELECT * FROM delete_timer($1::integer, $2::integer)`}

func (db *Database) DeleteCurrentTimerCommand(userId int, partyId int) (timer typetimers.EndedTimer, err error) {
	row := dbaccess.QueryRow(deleteCurrentTimerQuery, userId, partyId)

	var durationRaw, timeSpentRaw string
	err = row.Scan(
		&timer.Id,
		&timer.UserId,
		&timer.PartyId,
		&timer.GameId,
		&timer.State,
		&durationRaw,
		&timeSpentRaw,
		&timer.LastActionDate)

	if err == nil {
		timer.Duration, err = dbaccess.ScanInterval(durationRaw)
	}
	if err == nil {
		timer.TimeSpent, err = dbaccess.ScanInterval(timeSpentRaw)
	}

	dbaccess.LogDbResult(deleteCurrentTimerQuery, timer, err)

	return
}
