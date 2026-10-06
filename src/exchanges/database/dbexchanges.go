package dbexchanges

import (
	"FGG-Service/src/changes/types"
	"FGG-Service/src/dbaccess"
	"FGG-Service/src/exchanges/types"
	"context"
	"encoding/json"
)

type IDatabase interface {
	CreateExchangeCommand(ctx context.Context, partyId int, name string, description string, sourceChange typechanges.Change, targetChange typechanges.Change) (exchange typeexchanges.ExchangeWithChanges, err error)
	GetExchangeWithEntriesCommand(ctx context.Context, partyId int, exchangeId int) (exchange typeexchanges.ExchangeWithChanges, err error)
	GetActualExchangesCommand(ctx context.Context, partyId int) (exchanges []typeexchanges.Exchange, err error)
	GetRemovedExchangesCommand(ctx context.Context, partyId int) (exchanges []typeexchanges.Exchange, err error)
	RemoveExchangeCommand(ctx context.Context, partyId int, exchangeId int) error
	CreateExchangeHistoryCommand(ctx context.Context, userId int, partyId int, exchangeId int, actorUserId int, sourceEventId *int) (entry typeexchanges.ExchangeHistoryEntry, err error)
	GetExchangeHistoryCommand(ctx context.Context, userId int, partyId int) (history []typeexchanges.ExchangeHistory, err error)
}

type Database struct{}

func scanExchangeWithChanges(row interface {
	Scan(dest ...any) error
}) (exchange typeexchanges.ExchangeWithChanges, err error) {
	var sourceRaw, targetRaw []byte
	err = row.Scan(&exchange.Id, &exchange.Name, &exchange.Description, &sourceRaw, &targetRaw)

	if err == nil {
		err = json.Unmarshal(sourceRaw, &exchange.SourceChange)
	}
	if err == nil {
		err = json.Unmarshal(targetRaw, &exchange.TargetChange)
	}

	return
}

var createExchangeQuery = dbaccess.Query{Name: "CreateExchangeQuery", SQL: `SELECT * FROM create_exchange($1::integer, $2::text, $3::text, $4::jsonb, $5::jsonb)`}

func (db *Database) CreateExchangeCommand(ctx context.Context, partyId int, name string, description string, sourceChange typechanges.Change, targetChange typechanges.Change) (exchange typeexchanges.ExchangeWithChanges, err error) {
	sourceJson, err := json.Marshal(sourceChange)
	if err != nil {
		return
	}

	targetJson, err := json.Marshal(targetChange)
	if err != nil {
		return
	}

	row := dbaccess.QueryRow(ctx, createExchangeQuery, partyId, name, description, sourceJson, targetJson)

	exchange, err = scanExchangeWithChanges(row)

	dbaccess.LogDbResult(createExchangeQuery, exchange, err)

	return
}

var getExchangeWithEntriesQuery = dbaccess.Query{Name: "GetExchangeWithEntriesQuery", SQL: `SELECT * FROM get_exchange_with_entries($1::integer, $2::integer)`}

func (db *Database) GetExchangeWithEntriesCommand(ctx context.Context, partyId int, exchangeId int) (exchange typeexchanges.ExchangeWithChanges, err error) {
	row := dbaccess.QueryRow(ctx, getExchangeWithEntriesQuery, partyId, exchangeId)

	exchange, err = scanExchangeWithChanges(row)

	dbaccess.LogDbResult(getExchangeWithEntriesQuery, exchange, err)

	return
}

func scanExchanges(ctx context.Context, q dbaccess.Query, partyId int) (exchanges []typeexchanges.Exchange, err error) {
	rows, err := dbaccess.QueryRows(ctx, q, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		exchange := typeexchanges.Exchange{}
		err = rows.Scan(&exchange.Id, &exchange.PartyId, &exchange.Name, &exchange.Description)

		if err != nil {
			_ = rows.Close()
			return
		}

		exchanges = append(exchanges, exchange)
	}

	err = rows.Err()

	dbaccess.LogDbResult(q, exchanges, err)

	_ = rows.Close()
	return
}

var getActualExchangesQuery = dbaccess.Query{Name: "GetActualExchangesQuery", SQL: `SELECT * FROM get_actual_exchanges($1::integer)`}

func (db *Database) GetActualExchangesCommand(ctx context.Context, partyId int) (exchanges []typeexchanges.Exchange, err error) {
	return scanExchanges(ctx, getActualExchangesQuery, partyId)
}

var getRemovedExchangesQuery = dbaccess.Query{Name: "GetRemovedExchangesQuery", SQL: `SELECT * FROM get_removed_exchanges($1::integer)`}

func (db *Database) GetRemovedExchangesCommand(ctx context.Context, partyId int) (exchanges []typeexchanges.Exchange, err error) {
	return scanExchanges(ctx, getRemovedExchangesQuery, partyId)
}

var removeExchangeQuery = dbaccess.Query{Name: "RemoveExchangeQuery", SQL: `SELECT remove_exchange($1::integer, $2::integer)`}

func (db *Database) RemoveExchangeCommand(ctx context.Context, partyId int, exchangeId int) error {
	_, err := dbaccess.Exec(ctx, removeExchangeQuery, partyId, exchangeId)

	dbaccess.LogDbResult(removeExchangeQuery, nil, err)

	return err
}

var createExchangeHistoryQuery = dbaccess.Query{Name: "CreateExchangeHistoryQuery", SQL: `SELECT * FROM create_exchange_history($1::integer, $2::integer, $3::integer, $4::integer, $5::integer)`}

func (db *Database) CreateExchangeHistoryCommand(ctx context.Context, userId int, partyId int, exchangeId int, actorUserId int, sourceEventId *int) (entry typeexchanges.ExchangeHistoryEntry, err error) {
	row := dbaccess.QueryRow(ctx, createExchangeHistoryQuery, userId, partyId, exchangeId, actorUserId, sourceEventId)

	err = row.Scan(&entry.Id, &entry.UserId, &entry.PartyId, &entry.ExchangeId, &entry.UsedDate)

	dbaccess.LogDbResult(createExchangeHistoryQuery, entry, err)

	return
}

var getExchangeHistoryQuery = dbaccess.Query{Name: "GetExchangeHistoryQuery", SQL: `SELECT * FROM get_exchange_history($1::integer, $2::integer)`}

func (db *Database) GetExchangeHistoryCommand(ctx context.Context, userId int, partyId int) (history []typeexchanges.ExchangeHistory, err error) {
	rows, err := dbaccess.QueryRows(ctx, getExchangeHistoryQuery, userId, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		entry := typeexchanges.ExchangeHistory{}
		err = rows.Scan(&entry.Id, &entry.UserId, &entry.PartyId, &entry.ExchangeId, &entry.Name, &entry.Description, &entry.UsedDate)

		if err != nil {
			_ = rows.Close()
			return
		}

		history = append(history, entry)
	}

	err = rows.Err()

	dbaccess.LogDbResult(getExchangeHistoryQuery, history, err)

	_ = rows.Close()
	return
}
