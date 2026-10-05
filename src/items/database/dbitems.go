package dbitems

import (
	"FGG-Service/src/changes/types"
	"FGG-Service/src/dbaccess"
	"FGG-Service/src/items/types"
	"encoding/json"
)

type IDatabase interface {
	CreateItemCommand(partyId int, name string, description string, useCount int, change typechanges.Change) (item typeitems.ItemWithChange, err error)
	GetItemCommand(partyId int, itemId int) (item typeitems.ItemWithChange, err error)
	GetActualItemsCommand(partyId int) (items []typeitems.Item, err error)
	GetRemovedItemsCommand(partyId int) (items []typeitems.Item, err error)
	RemoveItemCommand(partyId int, itemId int) error
	CreateUserItemCommand(userId int, partyId int, itemId int, actorUserId int, sourceEventId int) (userItem typeitems.UserItem, err error)
	GetUserItemCommand(userId int, partyId int, itemId int) (userItem typeitems.UserItem, err error)
	GetUserItemsCommand(userId int, partyId int) (userItems []typeitems.UserItem, err error)
	ChangeUserItemUsesLeftCommand(userId int, partyId int, itemId int, usesLeft int, actorUserId int, sourceEventId *int) (historyEventId int, err error)
	GetItemHistoryCommand(userId int, partyId int) (history []typeitems.ItemHistory, err error)
	DeleteUserItemCommand(userId int, partyId int, itemId int, actorUserId int, sourceEventId *int) error
}

type Database struct{}

var createItemQuery = dbaccess.Query{Name: "CreateItemQuery", SQL: `SELECT * FROM create_item($1::integer, $2::text, $3::text, $4::integer, $5::jsonb)`}

func (db *Database) CreateItemCommand(partyId int, name string, description string, useCount int, change typechanges.Change) (item typeitems.ItemWithChange, err error) {
	changeJson, err := json.Marshal(change)
	if err != nil {
		return
	}

	row := dbaccess.QueryRow(createItemQuery, partyId, name, description, useCount, changeJson)

	var changeRaw []byte
	err = row.Scan(&item.Id, &item.PartyId, &item.Name, &item.Description, &item.UseCount, &changeRaw)
	if err == nil {
		err = json.Unmarshal(changeRaw, &item.Change)
	}

	dbaccess.LogDbResult(createItemQuery, item, err)

	return
}

var getItemQuery = dbaccess.Query{Name: "GetItemQuery", SQL: `SELECT * FROM get_item($1::integer, $2::integer)`}

func (db *Database) GetItemCommand(partyId int, itemId int) (item typeitems.ItemWithChange, err error) {
	row := dbaccess.QueryRow(getItemQuery, partyId, itemId)

	var changeRaw []byte
	err = row.Scan(&item.Id, &item.PartyId, &item.Name, &item.Description, &item.UseCount, &changeRaw)
	if err == nil {
		err = json.Unmarshal(changeRaw, &item.Change)
	}

	dbaccess.LogDbResult(getItemQuery, item, err)

	return
}

func scanItems(q dbaccess.Query, partyId int) (items []typeitems.Item, err error) {
	rows, err := dbaccess.QueryRows(q, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		item := typeitems.Item{}
		err = rows.Scan(&item.Id, &item.PartyId, &item.Name, &item.Description, &item.UseCount, &item.ChangeId)

		if err != nil {
			_ = rows.Close()
			return
		}

		items = append(items, item)
	}

	dbaccess.LogDbResult(q, items, err)

	_ = rows.Close()
	return
}

var getActualItemsQuery = dbaccess.Query{Name: "GetActualItemsQuery", SQL: `SELECT * FROM get_actual_items($1::integer)`}

func (db *Database) GetActualItemsCommand(partyId int) (items []typeitems.Item, err error) {
	return scanItems(getActualItemsQuery, partyId)
}

var getRemovedItemsQuery = dbaccess.Query{Name: "GetRemovedItemsQuery", SQL: `SELECT * FROM get_removed_items($1::integer)`}

func (db *Database) GetRemovedItemsCommand(partyId int) (items []typeitems.Item, err error) {
	return scanItems(getRemovedItemsQuery, partyId)
}

var removeItemQuery = dbaccess.Query{Name: "RemoveItemQuery", SQL: `SELECT remove_item($1::integer, $2::integer)`}

func (db *Database) RemoveItemCommand(partyId int, itemId int) error {
	_, err := dbaccess.Exec(removeItemQuery, partyId, itemId)

	dbaccess.LogDbResult(removeItemQuery, nil, err)

	return err
}

var createUserItemQuery = dbaccess.Query{Name: "CreateUserItemQuery", SQL: `SELECT * FROM create_user_item($1::integer, $2::integer, $3::integer, $4::integer, $5::integer)`}

func (db *Database) CreateUserItemCommand(userId int, partyId int, itemId int, actorUserId int, sourceEventId int) (userItem typeitems.UserItem, err error) {
	row := dbaccess.QueryRow(createUserItemQuery, userId, partyId, itemId, actorUserId, sourceEventId)

	err = row.Scan(&userItem.Id, &userItem.UserId, &userItem.PartyId, &userItem.ItemId, &userItem.UsesLeft, &userItem.ReceivedDate)

	dbaccess.LogDbResult(createUserItemQuery, userItem, err)

	return
}

var getUserItemQuery = dbaccess.Query{Name: "GetUserItemQuery", SQL: `SELECT * FROM get_user_item($1::integer, $2::integer, $3::integer)`}

func (db *Database) GetUserItemCommand(userId int, partyId int, itemId int) (userItem typeitems.UserItem, err error) {
	row := dbaccess.QueryRow(getUserItemQuery, userId, partyId, itemId)

	err = row.Scan(&userItem.Id, &userItem.UserId, &userItem.PartyId, &userItem.ItemId, &userItem.UsesLeft, &userItem.ReceivedDate)

	dbaccess.LogDbResult(getUserItemQuery, userItem, err)

	return
}

var getUserItemsQuery = dbaccess.Query{Name: "GetUserItemsQuery", SQL: `SELECT * FROM get_user_items($1::integer, $2::integer)`}

func (db *Database) GetUserItemsCommand(userId int, partyId int) (userItems []typeitems.UserItem, err error) {
	rows, err := dbaccess.QueryRows(getUserItemsQuery, userId, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		userItem := typeitems.UserItem{}
		err = rows.Scan(&userItem.Id, &userItem.UserId, &userItem.PartyId, &userItem.ItemId, &userItem.UsesLeft, &userItem.ReceivedDate)

		if err != nil {
			_ = rows.Close()
			return
		}

		userItems = append(userItems, userItem)
	}

	dbaccess.LogDbResult(getUserItemsQuery, userItems, err)

	_ = rows.Close()
	return
}

var changeUserItemUsesLeftQuery = dbaccess.Query{Name: "ChangeUserItemUsesLeftQuery", SQL: `SELECT change_user_item_uses_left($1::integer, $2::integer, $3::integer, $4::integer, $5::integer, $6::integer)`}

func (db *Database) ChangeUserItemUsesLeftCommand(userId int, partyId int, itemId int, usesLeft int, actorUserId int, sourceEventId *int) (historyEventId int, err error) {
	row := dbaccess.QueryRow(changeUserItemUsesLeftQuery, userId, partyId, itemId, usesLeft, actorUserId, sourceEventId)

	err = row.Scan(&historyEventId)

	dbaccess.LogDbResult(changeUserItemUsesLeftQuery, historyEventId, err)

	return
}

var getItemHistoryQuery = dbaccess.Query{Name: "GetItemHistoryQuery", SQL: `SELECT * FROM get_item_history($1::integer, $2::integer)`}

func (db *Database) GetItemHistoryCommand(userId int, partyId int) (history []typeitems.ItemHistory, err error) {
	rows, err := dbaccess.QueryRows(getItemHistoryQuery, userId, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		entry := typeitems.ItemHistory{}
		err = rows.Scan(&entry.Id, &entry.UserId, &entry.ActorUserId, &entry.PartyId, &entry.ItemId, &entry.Action, &entry.UsesLeft, &entry.SourceEventId, &entry.CreatedDate)

		if err != nil {
			_ = rows.Close()
			return
		}

		history = append(history, entry)
	}

	dbaccess.LogDbResult(getItemHistoryQuery, history, err)

	_ = rows.Close()
	return
}

var deleteUserItemQuery = dbaccess.Query{Name: "DeleteUserItemQuery", SQL: `SELECT delete_user_item($1::integer, $2::integer, $3::integer, $4::integer, $5::integer)`}

func (db *Database) DeleteUserItemCommand(userId int, partyId int, itemId int, actorUserId int, sourceEventId *int) error {
	_, err := dbaccess.Exec(deleteUserItemQuery, userId, partyId, itemId, actorUserId, sourceEventId)

	dbaccess.LogDbResult(deleteUserItemQuery, nil, err)

	return err
}
