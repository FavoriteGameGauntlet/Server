package dbwheeleffects

import (
	"FGG-Service/src/dbaccess"
	dbpoints "FGG-Service/src/points/database"
	typepoints "FGG-Service/src/points/type"
	"FGG-Service/src/wheeleffects/types"
	"database/sql"
	"encoding/json"
	"errors"
)

// availableRollsDefaultPartyId is a stopgap until real party-context resolution exists (see project
// plan). GetAvailableRollsCountCommand keeps its historical userId-only signature so callers without
// party context yet (e.g. srvtimers) don't need it either.
const availableRollsDefaultPartyId = 1

type IDatabase interface {
	// GetAvailableRollsCountCommand reads the user's AvailableRolls PointType value, under
	// availableRollsDefaultPartyId. See its doc comment for why this stays userId-only.
	GetAvailableRollsCountCommand(userId int) (count int, err error)
	GetAvailableWheelRowsCommand(userId int, partyId int, collectionId int) (rows []typewheeleffects.WheelRow, err error)
	GetEffectHistoryCommand(userId int, partyId int) (history []typewheeleffects.WheelRowHistory, err error)
	GetEffectHistoryByEffectNameCommand(userId int, partyId int, wheelRowName string) (history typewheeleffects.WheelRowHistory, err error)
	ClearLastWheelEffectsCommand(userId int, partyId int) error
	AddLastRolledWheelEffectsCommand(userId int, partyId int, rows []typewheeleffects.RolledWheelRowInput) (created []typewheeleffects.CreatedLastWheelRow, err error)
	GetLastRolledWheelEffectsCommand(userId int, partyId int) (rows []typewheeleffects.LastWheelRow, err error)
	AddWheelEffectHistoryCommand(userId int, partyId int, wheelRowId int, actorUserId int, sourceEventId *int) (created typewheeleffects.CreatedWheelRowHistory, err error)
	CreateWheelCollectionCommand(partyId int, name string, shouldCheckHistory bool) (collection typewheeleffects.WheelCollection, err error)
	GetWheelCollectionsCommand(partyId int) (collections []typewheeleffects.WheelCollection, err error)
	CreateWheelRowCommand(partyId int, name string, description string, changeId int, collectionId int) (row typewheeleffects.CreatedWheelRow, err error)
	GetWheelRowCommand(partyId int, wheelRowId int) (row typewheeleffects.WheelRow, err error)
	GetWheelRowsCommand(partyId int) (rows []typewheeleffects.WheelRow, err error)
	DeleteWheelRowCommand(partyId int, wheelRowId int) error
}

type Database struct {
}

func (db *Database) GetAvailableRollsCountCommand(userId int) (count int, err error) {
	pointsDb := new(dbpoints.Database)

	pointTypes, err := pointsDb.GetPointTypesCommand(availableRollsDefaultPartyId)

	if err != nil {
		return
	}

	for _, pointType := range pointTypes {
		if pointType.Name != typepoints.PointTypeAvailableRolls {
			continue
		}

		var point typepoints.UserPoint
		point, err = pointsDb.GetUserPointCommand(userId, availableRollsDefaultPartyId, pointType.Id)

		if errors.Is(err, sql.ErrNoRows) {
			err = nil
			return
		}

		if err != nil {
			return
		}

		count = point.Value
		return
	}

	return
}

var getAvailableWheelRowsQuery = dbaccess.Query{Name: "GetAvailableWheelRowsQuery", SQL: `SELECT * FROM get_available_wheel_rows($1::integer, $2::integer, $3::integer)`}

func (db *Database) GetAvailableWheelRowsCommand(userId int, partyId int, collectionId int) (rows []typewheeleffects.WheelRow, err error) {
	rs, err := dbaccess.QueryRows(getAvailableWheelRowsQuery, userId, partyId, collectionId)

	if err != nil {
		return
	}

	for rs.Next() {
		row := typewheeleffects.WheelRow{}
		err = rs.Scan(&row.Id, &row.PartyId, &row.Name, &row.Description, &row.ChangeId, &row.CollectionId, &row.IsManualChange)

		if err != nil {
			_ = rs.Close()
			return
		}

		rows = append(rows, row)
	}

	dbaccess.LogDbResult(getAvailableWheelRowsQuery, rows, err)

	_ = rs.Close()
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

var clearLastWheelEffectsQuery = dbaccess.Query{Name: "ClearLastWheelEffectsQuery", SQL: `SELECT clear_last_wheel_rows($1::integer, $2::integer)`}

func (db *Database) ClearLastWheelEffectsCommand(userId int, partyId int) error {
	_, err := dbaccess.Exec(clearLastWheelEffectsQuery, userId, partyId)

	dbaccess.LogDbResult(clearLastWheelEffectsQuery, nil, err)

	return err
}

var addLastRolledWheelEffectsQuery = dbaccess.Query{Name: "AddLastRolledWheelEffectsQuery", SQL: `SELECT * FROM create_last_wheel_rows($1::integer, $2::integer, $3::jsonb)`}

func (db *Database) AddLastRolledWheelEffectsCommand(userId int, partyId int, rows []typewheeleffects.RolledWheelRowInput) (created []typewheeleffects.CreatedLastWheelRow, err error) {
	type rowEntry struct {
		WheelRowId int `json:"wheel_row_id"`
		Position   int `json:"position"`
	}

	entries := make([]rowEntry, len(rows))
	for i, row := range rows {
		entries[i] = rowEntry{WheelRowId: row.WheelRowId, Position: row.Position}
	}

	entriesJson, err := json.Marshal(entries)
	if err != nil {
		return
	}

	rs, err := dbaccess.QueryRows(addLastRolledWheelEffectsQuery, userId, partyId, entriesJson)

	if err != nil {
		return
	}

	for rs.Next() {
		row := typewheeleffects.CreatedLastWheelRow{}
		err = rs.Scan(&row.Id, &row.UserId, &row.PartyId, &row.WheelRowId, &row.WheelPosition, &row.RolledDate)

		if err != nil {
			_ = rs.Close()
			return
		}

		created = append(created, row)
	}

	dbaccess.LogDbResult(addLastRolledWheelEffectsQuery, created, err)

	_ = rs.Close()
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

var addWheelEffectHistoryQuery = dbaccess.Query{Name: "AddWheelEffectHistoryQuery", SQL: `SELECT * FROM create_wheel_row_history($1::integer, $2::integer, $3::integer, $4::integer, $5::integer)`}

func (db *Database) AddWheelEffectHistoryCommand(userId int, partyId int, wheelRowId int, actorUserId int, sourceEventId *int) (created typewheeleffects.CreatedWheelRowHistory, err error) {
	row := dbaccess.QueryRow(addWheelEffectHistoryQuery, userId, partyId, wheelRowId, actorUserId, sourceEventId)

	err = row.Scan(&created.Id, &created.UserId, &created.PartyId, &created.WheelRowId, &created.AppliedDate)

	dbaccess.LogDbResult(addWheelEffectHistoryQuery, created, err)

	return
}

var createWheelCollectionQuery = dbaccess.Query{Name: "CreateWheelCollectionQuery", SQL: `SELECT * FROM create_wheel_collection($1::integer, $2::text, $3::boolean)`}

func (db *Database) CreateWheelCollectionCommand(partyId int, name string, shouldCheckHistory bool) (collection typewheeleffects.WheelCollection, err error) {
	row := dbaccess.QueryRow(createWheelCollectionQuery, partyId, name, shouldCheckHistory)

	err = row.Scan(&collection.Id, &collection.PartyId, &collection.Name, &collection.ShouldCheckHistory)

	dbaccess.LogDbResult(createWheelCollectionQuery, collection, err)

	return
}

var getWheelCollectionsQuery = dbaccess.Query{Name: "GetWheelCollectionsQuery", SQL: `SELECT * FROM get_wheel_collections($1::integer)`}

func (db *Database) GetWheelCollectionsCommand(partyId int) (collections []typewheeleffects.WheelCollection, err error) {
	rows, err := dbaccess.QueryRows(getWheelCollectionsQuery, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		collection := typewheeleffects.WheelCollection{}
		err = rows.Scan(&collection.Id, &collection.PartyId, &collection.Name, &collection.ShouldCheckHistory)

		if err != nil {
			_ = rows.Close()
			return
		}

		collections = append(collections, collection)
	}

	dbaccess.LogDbResult(getWheelCollectionsQuery, collections, err)

	_ = rows.Close()
	return
}

var createWheelRowQuery = dbaccess.Query{Name: "CreateWheelRowQuery", SQL: `SELECT * FROM create_wheel_row($1::integer, $2::text, $3::text, $4::integer, $5::integer)`}

func (db *Database) CreateWheelRowCommand(partyId int, name string, description string, changeId int, collectionId int) (row typewheeleffects.CreatedWheelRow, err error) {
	r := dbaccess.QueryRow(createWheelRowQuery, partyId, name, description, changeId, collectionId)

	err = r.Scan(&row.Id, &row.PartyId, &row.Name, &row.Description, &row.ChangeId, &row.CollectionId)

	dbaccess.LogDbResult(createWheelRowQuery, row, err)

	return
}

var getWheelRowQuery = dbaccess.Query{Name: "GetWheelRowQuery", SQL: `SELECT * FROM get_wheel_row($1::integer, $2::integer)`}

func (db *Database) GetWheelRowCommand(partyId int, wheelRowId int) (row typewheeleffects.WheelRow, err error) {
	r := dbaccess.QueryRow(getWheelRowQuery, partyId, wheelRowId)

	err = r.Scan(&row.Id, &row.PartyId, &row.Name, &row.Description, &row.ChangeId, &row.CollectionId, &row.IsManualChange)

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
		err = rs.Scan(&row.Id, &row.PartyId, &row.Name, &row.Description, &row.ChangeId, &row.CollectionId, &row.IsManualChange)

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
