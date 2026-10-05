package dbgrants

import (
	"FGG-Service/src/changes/types"
	"FGG-Service/src/dbaccess"
	"FGG-Service/src/grants/types"
	"encoding/json"
)

type IDatabase interface {
	CreateManualHistoryCommand(partyId int, actorUserId int, entries []typechanges.ChangeEntry, sourceEventId *int) (created []typegrants.ManualHistoryEntry, err error)
}

type Database struct{}

var createManualHistoryQuery = dbaccess.Query{Name: "CreateManualHistoryQuery", SQL: `SELECT * FROM create_manual_history($1::integer, $2::integer, $3::jsonb, $4::integer)`}

func (db *Database) CreateManualHistoryCommand(partyId int, actorUserId int, entries []typechanges.ChangeEntry, sourceEventId *int) (created []typegrants.ManualHistoryEntry, err error) {
	entriesJson, err := json.Marshal(entries)
	if err != nil {
		return
	}

	rows, err := dbaccess.QueryRows(createManualHistoryQuery, partyId, actorUserId, entriesJson, sourceEventId)

	if err != nil {
		return
	}

	for rows.Next() {
		entry := typegrants.ManualHistoryEntry{}
		err = rows.Scan(&entry.Id, &entry.UserId, &entry.PartyId, &entry.ActorUserId, &entry.ChangeId, &entry.CreatedDate)

		if err != nil {
			_ = rows.Close()
			return
		}

		created = append(created, entry)
	}

	dbaccess.LogDbResult(createManualHistoryQuery, created, err)

	_ = rows.Close()
	return
}
