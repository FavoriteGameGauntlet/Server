package dbwheeleffects

import (
	"FGG-Service/src/dbaccess"
	"FGG-Service/src/wheeleffects/types"
	"database/sql"
	"errors"
)

type IDatabase interface {
	GetAvailableRollsCountCommand(userId int) (count int, err error)
	GetAvailableEffectsCommand(userId int) (effects typewheeleffects.WheelEffects, err error)
	GetEffectHistoryCommand(userId int, partyId int) (history []typewheeleffects.WheelRowHistory, err error)
	GetEffectHistoryByEffectNameCommand(userId int, partyId int, wheelRowName string) (history typewheeleffects.WheelRowHistory, err error)
	MakeEffectRollCommand(userId int, partyId int, groupId int) (rolled typewheeleffects.RolledWheelRow, err error)
	ClearLastWheelEffectsCommand(userId int, partyId int) error
	AddLastRolledWheelEffectCommand(userId int, partyId int, wheelRowId int, position int) (created typewheeleffects.CreatedLastWheelRow, err error)
	GetLastRolledWheelEffectsCommand(userId int, partyId int) (rows []typewheeleffects.LastWheelRow, err error)
	MarkLastWheelEffectAppliedCommand(userId int, wheelEffectId int) error
	AddWheelEffectHistoryCommand(userId int, partyId int, wheelRowId int, sourceEventId *int) (created typewheeleffects.CreatedWheelRowHistory, err error)
	CreateWheelGroupCommand(partyId int, name string) (group typewheeleffects.WheelGroup, err error)
	GetWheelGroupsCommand(partyId int) (groups []typewheeleffects.WheelGroup, err error)
	CreateWheelRowCommand(partyId int, name string, description string, changeId int, groupId int) (row typewheeleffects.CreatedWheelRow, err error)
	GetWheelRowCommand(partyId int, wheelRowId int) (row typewheeleffects.WheelRow, err error)
	GetWheelRowsCommand(partyId int) (rows []typewheeleffects.WheelRow, err error)
	DeleteWheelRowCommand(partyId int, wheelRowId int) error
}

type Database struct {
}

var getAvailableRollsCountQuery = dbaccess.Query{Name: "GetAvailableRollsCountQuery", SQL: `SELECT * FROM get_available_rolls_count($1::integer)`}

func (db *Database) GetAvailableRollsCountCommand(userId int) (count int, err error) {
	row := dbaccess.QueryRow(getAvailableRollsCountQuery, userId)

	err = row.Scan(&count)

	if errors.Is(err, sql.ErrNoRows) {
		count = 0
		err = nil
	}

	dbaccess.LogDbResult(getAvailableRollsCountQuery, count, err)

	return
}

var getAvailableEffectsQuery = dbaccess.Query{Name: "GetAvailableEffectsQuery", SQL: `SELECT * FROM get_available_effects($1::integer)`}

func (db *Database) GetAvailableEffectsCommand(userId int) (effects typewheeleffects.WheelEffects, err error) {
	rows, err := dbaccess.QueryRows(getAvailableEffectsQuery, userId)

	if err != nil {
		return
	}

	for rows.Next() {
		effect := typewheeleffects.WheelEffect{}
		err = rows.Scan(&effect.Id, &effect.Name, &effect.Description)

		if err != nil {
			dbaccess.LogDbResult(getAvailableEffectsQuery, effects, err)

			_ = rows.Close()
			return
		}

		effects = append(effects, effect)
	}

	dbaccess.LogDbResult(getAvailableEffectsQuery, effects, err)

	_ = rows.Close()
	return
}

var getEffectHistoryQuery = dbaccess.Query{Name: "GetEffectHistoryQuery", SQL: `SELECT * FROM get_wheel_row_history($1::integer, $2::integer)`}

func (db *Database) GetEffectHistoryCommand(userId int, partyId int) (history []typewheeleffects.WheelRowHistory, err error) {
	rows, err := dbaccess.QueryRows(getEffectHistoryQuery, userId, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		entry := typewheeleffects.WheelRowHistory{}
		err = rows.Scan(&entry.Name, &entry.Description, &entry.AppliedDate)

		if err != nil {
			dbaccess.LogDbResult(getEffectHistoryQuery, history, err)

			_ = rows.Close()
			return
		}

		history = append(history, entry)
	}

	dbaccess.LogDbResult(getEffectHistoryQuery, history, err)

	_ = rows.Close()
	return
}

var getEffectHistoryByEffectNameQuery = dbaccess.Query{Name: "GetEffectHistoryByEffectNameQuery", SQL: `SELECT * FROM get_wheel_row_history_by_name($1::integer, $2::integer, $3::text)`}

func (db *Database) GetEffectHistoryByEffectNameCommand(userId int, partyId int, wheelRowName string) (history typewheeleffects.WheelRowHistory, err error) {
	row := dbaccess.QueryRow(getEffectHistoryByEffectNameQuery, userId, partyId, wheelRowName)

	err = row.Scan(&history.Name, &history.Description, &history.AppliedDate)

	dbaccess.LogDbResult(getEffectHistoryByEffectNameQuery, history, err)

	return
}

var makeEffectRollQuery = dbaccess.Query{Name: "MakeEffectRollQuery", SQL: `SELECT * FROM roll_wheel_group($1::integer, $2::integer, $3::integer)`}

func (db *Database) MakeEffectRollCommand(userId int, partyId int, groupId int) (rolled typewheeleffects.RolledWheelRow, err error) {
	row := dbaccess.QueryRow(makeEffectRollQuery, userId, partyId, groupId)

	err = row.Scan(&rolled.Id, &rolled.UserId, &rolled.PartyId, &rolled.WheelRowId, &rolled.WheelPosition, &rolled.RolledDate, &rolled.IsManualChange)

	dbaccess.LogDbResult(makeEffectRollQuery, rolled, err)

	return
}

var clearLastWheelEffectsQuery = dbaccess.Query{Name: "ClearLastWheelEffectsQuery", SQL: `SELECT clear_last_wheel_rows($1::integer, $2::integer)`}

func (db *Database) ClearLastWheelEffectsCommand(userId int, partyId int) error {
	_, err := dbaccess.Exec(clearLastWheelEffectsQuery, userId, partyId)

	dbaccess.LogDbResult(clearLastWheelEffectsQuery, nil, err)

	return err
}

var addLastRolledWheelEffectsQuery = dbaccess.Query{Name: "AddLastRolledWheelEffectsQuery", SQL: `SELECT * FROM create_last_wheel_row($1::integer, $2::integer, $3::integer, $4::integer)`}

func (db *Database) AddLastRolledWheelEffectCommand(userId int, partyId int, wheelRowId int, position int) (created typewheeleffects.CreatedLastWheelRow, err error) {
	row := dbaccess.QueryRow(addLastRolledWheelEffectsQuery, userId, partyId, wheelRowId, position)

	err = row.Scan(&created.Id, &created.UserId, &created.PartyId, &created.WheelRowId, &created.WheelPosition, &created.RolledDate)

	dbaccess.LogDbResult(addLastRolledWheelEffectsQuery, created, err)

	return
}

var getLastRolledWheelEffectsQuery = dbaccess.Query{Name: "GetLastRolledWheelEffectsQuery", SQL: `SELECT * FROM get_last_wheel_rows($1::integer, $2::integer)`}

func (db *Database) GetLastRolledWheelEffectsCommand(userId int, partyId int) (rows []typewheeleffects.LastWheelRow, err error) {
	rs, err := dbaccess.QueryRows(getLastRolledWheelEffectsQuery, userId, partyId)

	if err != nil {
		return
	}

	for rs.Next() {
		row := typewheeleffects.LastWheelRow{}

		err = rs.Scan(&row.Id, &row.Name, &row.Description, &row.RolledDate, &row.WheelPosition)

		if err != nil {
			dbaccess.LogDbResult(getLastRolledWheelEffectsQuery, rows, err)

			_ = rs.Close()
			return
		}

		rows = append(rows, row)
	}

	dbaccess.LogDbResult(getLastRolledWheelEffectsQuery, rows, err)

	_ = rs.Close()
	return
}

var markLastWheelEffectAppliedQuery = dbaccess.Query{Name: "MarkLastWheelEffectAppliedQuery", SQL: `SELECT mark_last_wheel_effect_applied($1::integer, $2::integer)`}

func (db *Database) MarkLastWheelEffectAppliedCommand(userId int, wheelEffectId int) error {
	_, err := dbaccess.Exec(markLastWheelEffectAppliedQuery, userId, wheelEffectId)

	dbaccess.LogDbResult(markLastWheelEffectAppliedQuery, nil, err)

	return err
}

var addWheelEffectHistoryQuery = dbaccess.Query{Name: "AddWheelEffectHistoryQuery", SQL: `SELECT * FROM create_wheel_row_history($1::integer, $2::integer, $3::integer, $4::integer)`}

func (db *Database) AddWheelEffectHistoryCommand(userId int, partyId int, wheelRowId int, sourceEventId *int) (created typewheeleffects.CreatedWheelRowHistory, err error) {
	row := dbaccess.QueryRow(addWheelEffectHistoryQuery, userId, partyId, wheelRowId, sourceEventId)

	err = row.Scan(&created.Id, &created.UserId, &created.PartyId, &created.WheelRowId, &created.AppliedDate)

	dbaccess.LogDbResult(addWheelEffectHistoryQuery, created, err)

	return
}

var createWheelGroupQuery = dbaccess.Query{Name: "CreateWheelGroupQuery", SQL: `SELECT * FROM create_wheel_group($1::integer, $2::text)`}

func (db *Database) CreateWheelGroupCommand(partyId int, name string) (group typewheeleffects.WheelGroup, err error) {
	row := dbaccess.QueryRow(createWheelGroupQuery, partyId, name)

	err = row.Scan(&group.Id, &group.PartyId, &group.Name)

	dbaccess.LogDbResult(createWheelGroupQuery, group, err)

	return
}

var getWheelGroupsQuery = dbaccess.Query{Name: "GetWheelGroupsQuery", SQL: `SELECT * FROM get_wheel_groups($1::integer)`}

func (db *Database) GetWheelGroupsCommand(partyId int) (groups []typewheeleffects.WheelGroup, err error) {
	rows, err := dbaccess.QueryRows(getWheelGroupsQuery, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		group := typewheeleffects.WheelGroup{}
		err = rows.Scan(&group.Id, &group.PartyId, &group.Name)

		if err != nil {
			_ = rows.Close()
			return
		}

		groups = append(groups, group)
	}

	dbaccess.LogDbResult(getWheelGroupsQuery, groups, err)

	_ = rows.Close()
	return
}

var createWheelRowQuery = dbaccess.Query{Name: "CreateWheelRowQuery", SQL: `SELECT * FROM create_wheel_row($1::integer, $2::text, $3::text, $4::integer, $5::integer)`}

func (db *Database) CreateWheelRowCommand(partyId int, name string, description string, changeId int, groupId int) (row typewheeleffects.CreatedWheelRow, err error) {
	r := dbaccess.QueryRow(createWheelRowQuery, partyId, name, description, changeId, groupId)

	err = r.Scan(&row.Id, &row.PartyId, &row.Name, &row.Description, &row.ChangeId, &row.GroupId)

	dbaccess.LogDbResult(createWheelRowQuery, row, err)

	return
}

var getWheelRowQuery = dbaccess.Query{Name: "GetWheelRowQuery", SQL: `SELECT * FROM get_wheel_row($1::integer, $2::integer)`}

func (db *Database) GetWheelRowCommand(partyId int, wheelRowId int) (row typewheeleffects.WheelRow, err error) {
	r := dbaccess.QueryRow(getWheelRowQuery, partyId, wheelRowId)

	err = r.Scan(&row.Id, &row.PartyId, &row.Name, &row.Description, &row.ChangeId, &row.GroupId, &row.IsManualChange)

	dbaccess.LogDbResult(getWheelRowQuery, row, err)

	return
}

var getWheelRowsQuery = dbaccess.Query{Name: "GetWheelRowsQuery", SQL: `SELECT * FROM get_wheel_rows($1::integer)`}

func (db *Database) GetWheelRowsCommand(partyId int) (rows []typewheeleffects.WheelRow, err error) {
	rs, err := dbaccess.QueryRows(getWheelRowsQuery, partyId)

	if err != nil {
		return
	}

	for rs.Next() {
		row := typewheeleffects.WheelRow{}
		err = rs.Scan(&row.Id, &row.PartyId, &row.Name, &row.Description, &row.ChangeId, &row.GroupId, &row.IsManualChange)

		if err != nil {
			_ = rs.Close()
			return
		}

		rows = append(rows, row)
	}

	dbaccess.LogDbResult(getWheelRowsQuery, rows, err)

	_ = rs.Close()
	return
}

var deleteWheelRowQuery = dbaccess.Query{Name: "DeleteWheelRowQuery", SQL: `SELECT delete_wheel_row($1::integer, $2::integer)`}

func (db *Database) DeleteWheelRowCommand(partyId int, wheelRowId int) error {
	_, err := dbaccess.Exec(deleteWheelRowQuery, partyId, wheelRowId)

	dbaccess.LogDbResult(deleteWheelRowQuery, nil, err)

	return err
}
