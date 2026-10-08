package dbperks

import (
	"FGG-Service/src/dbaccess"
	"FGG-Service/src/perks/typeperks"
	"context"
)

type IDatabase interface {
	CreatePerkCommand(ctx context.Context, partyId int, name string, description string, effectId int) (perk typeperks.PerkWithRemoved, err error)
	GetPerkCommand(ctx context.Context, partyId int, perkId int) (perk typeperks.Perk, err error)
	GetActualPerksCommand(ctx context.Context, partyId int) (perks []typeperks.Perk, err error)
	GetRemovedPerksCommand(ctx context.Context, partyId int) (perks []typeperks.Perk, err error)
	RemovePerkCommand(ctx context.Context, partyId int, perkId int) error
	CreateUserPerkCommand(ctx context.Context, userId int, partyId int, perkId int, actorUserId int, sourceEventId int) (userPerk typeperks.UserPerk, err error)
	GetUserPerkCommand(ctx context.Context, userId int, partyId int, perkId int) (userPerk typeperks.UserPerk, err error)
	GetUserPerksCommand(ctx context.Context, userId int, partyId int) (userPerks []typeperks.UserPerk, err error)
	DeleteUserPerkCommand(ctx context.Context, userId int, partyId int, perkId int, actorUserId int, sourceEventId *int) error
	GetPerkHistoryCommand(ctx context.Context, userId int, partyId int) (history []typeperks.PerkHistory, err error)
}

type Database struct{}

var createPerkQuery = dbaccess.Query{Name: "CreatePerkQuery", SQL: `SELECT * FROM create_perk($1::integer, $2::text, $3::text, $4::integer)`}

func (db *Database) CreatePerkCommand(ctx context.Context, partyId int, name string, description string, effectId int) (perk typeperks.PerkWithRemoved, err error) {
	row := dbaccess.QueryRow(ctx, createPerkQuery, partyId, name, description, effectId)

	err = row.Scan(&perk.Id, &perk.PartyId, &perk.Name, &perk.Description, &perk.EffectId, &perk.IsRemoved)

	dbaccess.LogDbResult(createPerkQuery, perk, err)

	return
}

var getPerkQuery = dbaccess.Query{Name: "GetPerkQuery", SQL: `SELECT * FROM get_perk($1::integer, $2::integer)`}

func (db *Database) GetPerkCommand(ctx context.Context, partyId int, perkId int) (perk typeperks.Perk, err error) {
	row := dbaccess.QueryRow(ctx, getPerkQuery, partyId, perkId)

	err = row.Scan(&perk.Id, &perk.PartyId, &perk.Name, &perk.Description, &perk.EffectId)

	dbaccess.LogDbResult(getPerkQuery, perk, err)

	return
}

func scanPerks(ctx context.Context, q dbaccess.Query, partyId int) (perks []typeperks.Perk, err error) {
	rows, err := dbaccess.QueryRows(ctx, q, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		perk := typeperks.Perk{}
		err = rows.Scan(&perk.Id, &perk.PartyId, &perk.Name, &perk.Description, &perk.EffectId)

		if err != nil {
			_ = rows.Close()
			return
		}

		perks = append(perks, perk)
	}

	err = rows.Err()

	dbaccess.LogDbResult(q, perks, err)

	_ = rows.Close()
	return
}

var getActualPerksQuery = dbaccess.Query{Name: "GetActualPerksQuery", SQL: `SELECT * FROM get_actual_perks($1::integer)`}

func (db *Database) GetActualPerksCommand(ctx context.Context, partyId int) (perks []typeperks.Perk, err error) {
	return scanPerks(ctx, getActualPerksQuery, partyId)
}

var getRemovedPerksQuery = dbaccess.Query{Name: "GetRemovedPerksQuery", SQL: `SELECT * FROM get_removed_perks($1::integer)`}

func (db *Database) GetRemovedPerksCommand(ctx context.Context, partyId int) (perks []typeperks.Perk, err error) {
	return scanPerks(ctx, getRemovedPerksQuery, partyId)
}

var removePerkQuery = dbaccess.Query{Name: "RemovePerkQuery", SQL: `SELECT remove_perk($1::integer, $2::integer)`}

func (db *Database) RemovePerkCommand(ctx context.Context, partyId int, perkId int) error {
	_, err := dbaccess.Exec(ctx, removePerkQuery, partyId, perkId)

	dbaccess.LogDbResult(removePerkQuery, nil, err)

	return err
}

var createUserPerkQuery = dbaccess.Query{Name: "CreateUserPerkQuery", SQL: `SELECT * FROM create_user_perk($1::integer, $2::integer, $3::integer, $4::integer, $5::integer)`}

func (db *Database) CreateUserPerkCommand(ctx context.Context, userId int, partyId int, perkId int, actorUserId int, sourceEventId int) (userPerk typeperks.UserPerk, err error) {
	row := dbaccess.QueryRow(ctx, createUserPerkQuery, userId, partyId, perkId, actorUserId, sourceEventId)

	err = row.Scan(&userPerk.Id, &userPerk.UserId, &userPerk.PartyId, &userPerk.PerkId, &userPerk.UserEffectId, &userPerk.ReceivedDate)

	dbaccess.LogDbResult(createUserPerkQuery, userPerk, err)

	return
}

var getUserPerkQuery = dbaccess.Query{Name: "GetUserPerkQuery", SQL: `SELECT * FROM get_user_perk($1::integer, $2::integer, $3::integer)`}

func (db *Database) GetUserPerkCommand(ctx context.Context, userId int, partyId int, perkId int) (userPerk typeperks.UserPerk, err error) {
	row := dbaccess.QueryRow(ctx, getUserPerkQuery, userId, partyId, perkId)

	err = row.Scan(&userPerk.Id, &userPerk.UserId, &userPerk.PartyId, &userPerk.PerkId, &userPerk.UserEffectId, &userPerk.ReceivedDate)

	dbaccess.LogDbResult(getUserPerkQuery, userPerk, err)

	return
}

var getUserPerksQuery = dbaccess.Query{Name: "GetUserPerksQuery", SQL: `SELECT * FROM get_user_perks($1::integer, $2::integer)`}

func (db *Database) GetUserPerksCommand(ctx context.Context, userId int, partyId int) (userPerks []typeperks.UserPerk, err error) {
	rows, err := dbaccess.QueryRows(ctx, getUserPerksQuery, userId, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		userPerk := typeperks.UserPerk{}
		err = rows.Scan(&userPerk.Id, &userPerk.UserId, &userPerk.PartyId, &userPerk.PerkId, &userPerk.UserEffectId, &userPerk.ReceivedDate)

		if err != nil {
			_ = rows.Close()
			return
		}

		userPerks = append(userPerks, userPerk)
	}

	err = rows.Err()

	dbaccess.LogDbResult(getUserPerksQuery, userPerks, err)

	_ = rows.Close()
	return
}

var deleteUserPerkQuery = dbaccess.Query{Name: "DeleteUserPerkQuery", SQL: `SELECT delete_user_perk($1::integer, $2::integer, $3::integer, $4::integer, $5::integer)`}

func (db *Database) DeleteUserPerkCommand(ctx context.Context, userId int, partyId int, perkId int, actorUserId int, sourceEventId *int) error {
	_, err := dbaccess.Exec(ctx, deleteUserPerkQuery, userId, partyId, perkId, actorUserId, sourceEventId)

	dbaccess.LogDbResult(deleteUserPerkQuery, nil, err)

	return err
}

var getPerkHistoryQuery = dbaccess.Query{Name: "GetPerkHistoryQuery", SQL: `SELECT * FROM get_perk_history($1::integer, $2::integer)`}

func (db *Database) GetPerkHistoryCommand(ctx context.Context, userId int, partyId int) (history []typeperks.PerkHistory, err error) {
	rows, err := dbaccess.QueryRows(ctx, getPerkHistoryQuery, userId, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		entry := typeperks.PerkHistory{}
		err = rows.Scan(&entry.Id, &entry.UserId, &entry.ActorUserId, &entry.PartyId, &entry.PerkId, &entry.Action, &entry.SourceEventId, &entry.CreatedDate)

		if err != nil {
			_ = rows.Close()
			return
		}

		history = append(history, entry)
	}

	err = rows.Err()

	dbaccess.LogDbResult(getPerkHistoryQuery, history, err)

	_ = rows.Close()
	return
}
