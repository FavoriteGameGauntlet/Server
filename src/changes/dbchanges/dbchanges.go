package dbchanges

import (
	"FGG-Service/src/changes/typechanges"
	"FGG-Service/src/dbaccess"
	"context"
	"encoding/json"
)

type IDatabase interface {
	ChangeEntryToJsonbCommand(ctx context.Context, entryId *int, amount int, pointTypeId *int, itemId *int, perkId *int, effectId *int, userId *int) (entry typechanges.ChangeEntry, err error)
	ChangeToJsonbCommand(ctx context.Context, changeId *int, shouldApplyToAll bool, isManualChange bool, entries []typechanges.ChangeEntry) (change typechanges.Change, err error)
	CreateChangeFromJsonbCommand(ctx context.Context, partyId int, change typechanges.Change) (created typechanges.Change, err error)
	CreateUserChangeFromJsonbCommand(ctx context.Context, partyId int, entries []typechanges.ChangeEntry) (created typechanges.UserChange, err error)
	GetChangeEntriesJsonbCommand(ctx context.Context, partyId int, changeId int) (entries []typechanges.ChangeEntry, err error)
}

type Database struct{}

var changeEntryToJsonbQuery = dbaccess.Query{Name: "ChangeEntryToJsonbQuery", SQL: `SELECT change_entry_to_jsonb($1::integer, $2::integer, $3::integer, $4::integer, $5::integer, $6::integer, $7::integer)`}

func (db *Database) ChangeEntryToJsonbCommand(
	ctx context.Context,
	entryId *int,
	amount int,
	pointTypeId *int,
	itemId *int,
	perkId *int,
	effectId *int,
	userId *int) (entry typechanges.ChangeEntry, err error) {

	row := dbaccess.QueryRow(ctx, changeEntryToJsonbQuery, entryId, amount, pointTypeId, itemId, perkId, effectId, userId)

	var raw []byte
	err = row.Scan(&raw)
	if err == nil {
		err = json.Unmarshal(raw, &entry)
	}

	dbaccess.LogDbResult(changeEntryToJsonbQuery, entry, err)

	return
}

var changeToJsonbQuery = dbaccess.Query{Name: "ChangeToJsonbQuery", SQL: `SELECT change_to_jsonb($1::integer, $2::boolean, $3::boolean, $4::jsonb)`}

func (db *Database) ChangeToJsonbCommand(
	ctx context.Context,
	changeId *int,
	shouldApplyToAll bool,
	isManualChange bool,
	entries []typechanges.ChangeEntry) (change typechanges.Change, err error) {

	entriesJson, err := json.Marshal(entries)
	if err != nil {
		return
	}

	row := dbaccess.QueryRow(ctx, changeToJsonbQuery, changeId, shouldApplyToAll, isManualChange, entriesJson)

	var raw []byte
	err = row.Scan(&raw)
	if err == nil {
		err = json.Unmarshal(raw, &change)
	}

	dbaccess.LogDbResult(changeToJsonbQuery, change, err)

	return
}

var createChangeFromJsonbQuery = dbaccess.Query{Name: "CreateChangeFromJsonbQuery", SQL: `SELECT create_change_from_jsonb($1::integer, $2::jsonb)`}

func (db *Database) CreateChangeFromJsonbCommand(ctx context.Context, partyId int, change typechanges.Change) (created typechanges.Change, err error) {
	changeJson, err := json.Marshal(change)
	if err != nil {
		return
	}

	row := dbaccess.QueryRow(ctx, createChangeFromJsonbQuery, partyId, changeJson)

	var raw []byte
	err = row.Scan(&raw)
	if err == nil {
		err = json.Unmarshal(raw, &created)
	}

	dbaccess.LogDbResult(createChangeFromJsonbQuery, created, err)

	return
}

var createUserChangeFromJsonbQuery = dbaccess.Query{Name: "CreateUserChangeFromJsonbQuery", SQL: `SELECT create_user_change_from_jsonb($1::integer, $2::jsonb)`}

func (db *Database) CreateUserChangeFromJsonbCommand(ctx context.Context, partyId int, entries []typechanges.ChangeEntry) (created typechanges.UserChange, err error) {
	entriesJson, err := json.Marshal(entries)
	if err != nil {
		return
	}

	row := dbaccess.QueryRow(ctx, createUserChangeFromJsonbQuery, partyId, entriesJson)

	var raw []byte
	err = row.Scan(&raw)
	if err == nil {
		err = json.Unmarshal(raw, &created)
	}

	dbaccess.LogDbResult(createUserChangeFromJsonbQuery, created, err)

	return
}

var getChangeEntriesJsonbQuery = dbaccess.Query{Name: "GetChangeEntriesJsonbQuery", SQL: `SELECT get_change_entries_jsonb($1::integer, $2::integer)`}

func (db *Database) GetChangeEntriesJsonbCommand(ctx context.Context, partyId int, changeId int) (entries []typechanges.ChangeEntry, err error) {
	row := dbaccess.QueryRow(ctx, getChangeEntriesJsonbQuery, partyId, changeId)

	var raw []byte
	err = row.Scan(&raw)
	if err == nil && raw != nil {
		err = json.Unmarshal(raw, &entries)
	}

	dbaccess.LogDbResult(getChangeEntriesJsonbQuery, entries, err)

	return
}
