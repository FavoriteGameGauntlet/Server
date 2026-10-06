package dbgrants

import (
	"FGG-Service/src/dbaccess"
	"FGG-Service/src/grants/types"
	"context"
)

type IDatabase interface {
	CreateManualHistoryCommand(ctx context.Context, userId int, partyId int, changeId int, actorUserId int, sourceEventId *int) (entry typegrants.ManualHistoryEntry, err error)
}

type Database struct{}

var createManualHistoryQuery = dbaccess.Query{Name: "CreateManualHistoryQuery", SQL: `SELECT * FROM create_manual_history($1::integer, $2::integer, $3::integer, $4::integer, $5::integer)`}

func (db *Database) CreateManualHistoryCommand(ctx context.Context, userId int, partyId int, changeId int, actorUserId int, sourceEventId *int) (entry typegrants.ManualHistoryEntry, err error) {
	row := dbaccess.QueryRow(ctx, createManualHistoryQuery, userId, partyId, changeId, actorUserId, sourceEventId)

	err = row.Scan(&entry.Id, &entry.UserId, &entry.PartyId, &entry.ActorUserId, &entry.ChangeId, &entry.CreatedDate)

	dbaccess.LogDbResult(createManualHistoryQuery, entry, err)

	return
}
