package dbperks

import (
	"FGG-Service/src/dbaccess"
	"FGG-Service/src/perks/types"
)

type IDatabase interface {
	CreatePerkCommand(partyId int, name string, description string, effectId int) (perk typeperks.PerkWithRemoved, err error)
	GetPerkCommand(partyId int, perkId int) (perk typeperks.Perk, err error)
	GetActualPerksCommand(partyId int) (perks []typeperks.Perk, err error)
	GetRemovedPerksCommand(partyId int) (perks []typeperks.Perk, err error)
	RemovePerkCommand(partyId int, perkId int) error
	CreateUserPerkCommand(userId int, partyId int, perkId int, sourceEventId int) (userPerk typeperks.UserPerk, err error)
	GetUserPerkCommand(userId int, partyId int, perkId int) (userPerk typeperks.UserPerk, err error)
	GetUserPerksCommand(userId int, partyId int) (userPerks []typeperks.UserPerk, err error)
	DeleteUserPerkCommand(userId int, partyId int, perkId int, sourceEventId *int) error
	GetPerkHistoryCommand(userId int, partyId int) (history []typeperks.PerkHistory, err error)
}

type Database struct{}

var createPerkQuery = dbaccess.Query{Name: "CreatePerkQuery", SQL: `SELECT * FROM create_perk($1::integer, $2::text, $3::text, $4::integer)`}

func (db *Database) CreatePerkCommand(partyId int, name string, description string, effectId int) (perk typeperks.PerkWithRemoved, err error) {
	row := dbaccess.QueryRow(createPerkQuery, partyId, name, description, effectId)

	err = row.Scan(&perk.Id, &perk.PartyId, &perk.Name, &perk.Description, &perk.EffectId, &perk.IsRemoved)

	dbaccess.LogDbResult(createPerkQuery, perk, err)

	return
}

var getPerkQuery = dbaccess.Query{Name: "GetPerkQuery", SQL: `SELECT * FROM get_perk($1::integer, $2::integer)`}

func (db *Database) GetPerkCommand(partyId int, perkId int) (perk typeperks.Perk, err error) {
	row := dbaccess.QueryRow(getPerkQuery, partyId, perkId)

	err = row.Scan(&perk.Id, &perk.PartyId, &perk.Name, &perk.Description, &perk.EffectId)

	dbaccess.LogDbResult(getPerkQuery, perk, err)

	return
}

func scanPerks(q dbaccess.Query, partyId int) (perks []typeperks.Perk, err error) {
	rows, err := dbaccess.QueryRows(q, partyId)

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

	dbaccess.LogDbResult(q, perks, err)

	_ = rows.Close()
	return
}

var getActualPerksQuery = dbaccess.Query{Name: "GetActualPerksQuery", SQL: `SELECT * FROM get_actual_perks($1::integer)`}

func (db *Database) GetActualPerksCommand(partyId int) (perks []typeperks.Perk, err error) {
	return scanPerks(getActualPerksQuery, partyId)
}

var getRemovedPerksQuery = dbaccess.Query{Name: "GetRemovedPerksQuery", SQL: `SELECT * FROM get_removed_perks($1::integer)`}

func (db *Database) GetRemovedPerksCommand(partyId int) (perks []typeperks.Perk, err error) {
	return scanPerks(getRemovedPerksQuery, partyId)
}

var removePerkQuery = dbaccess.Query{Name: "RemovePerkQuery", SQL: `SELECT remove_perk($1::integer, $2::integer)`}

func (db *Database) RemovePerkCommand(partyId int, perkId int) error {
	_, err := dbaccess.Exec(removePerkQuery, partyId, perkId)

	dbaccess.LogDbResult(removePerkQuery, nil, err)

	return err
}

var createUserPerkQuery = dbaccess.Query{Name: "CreateUserPerkQuery", SQL: `SELECT * FROM create_user_perk($1::integer, $2::integer, $3::integer, $4::integer)`}

func (db *Database) CreateUserPerkCommand(userId int, partyId int, perkId int, sourceEventId int) (userPerk typeperks.UserPerk, err error) {
	row := dbaccess.QueryRow(createUserPerkQuery, userId, partyId, perkId, sourceEventId)

	err = row.Scan(&userPerk.Id, &userPerk.UserId, &userPerk.PartyId, &userPerk.PerkId, &userPerk.UserEffectId, &userPerk.ReceivedDate)

	dbaccess.LogDbResult(createUserPerkQuery, userPerk, err)

	return
}

var getUserPerkQuery = dbaccess.Query{Name: "GetUserPerkQuery", SQL: `SELECT * FROM get_user_perk($1::integer, $2::integer, $3::integer)`}

func (db *Database) GetUserPerkCommand(userId int, partyId int, perkId int) (userPerk typeperks.UserPerk, err error) {
	row := dbaccess.QueryRow(getUserPerkQuery, userId, partyId, perkId)

	err = row.Scan(&userPerk.Id, &userPerk.UserId, &userPerk.PartyId, &userPerk.PerkId, &userPerk.UserEffectId, &userPerk.ReceivedDate)

	dbaccess.LogDbResult(getUserPerkQuery, userPerk, err)

	return
}

var getUserPerksQuery = dbaccess.Query{Name: "GetUserPerksQuery", SQL: `SELECT * FROM get_user_perks($1::integer, $2::integer)`}

func (db *Database) GetUserPerksCommand(userId int, partyId int) (userPerks []typeperks.UserPerk, err error) {
	rows, err := dbaccess.QueryRows(getUserPerksQuery, userId, partyId)

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

	dbaccess.LogDbResult(getUserPerksQuery, userPerks, err)

	_ = rows.Close()
	return
}

var deleteUserPerkQuery = dbaccess.Query{Name: "DeleteUserPerkQuery", SQL: `SELECT delete_user_perk($1::integer, $2::integer, $3::integer, $4::integer)`}

func (db *Database) DeleteUserPerkCommand(userId int, partyId int, perkId int, sourceEventId *int) error {
	_, err := dbaccess.Exec(deleteUserPerkQuery, userId, partyId, perkId, sourceEventId)

	dbaccess.LogDbResult(deleteUserPerkQuery, nil, err)

	return err
}

var getPerkHistoryQuery = dbaccess.Query{Name: "GetPerkHistoryQuery", SQL: `SELECT * FROM get_perk_history($1::integer, $2::integer)`}

func (db *Database) GetPerkHistoryCommand(userId int, partyId int) (history []typeperks.PerkHistory, err error) {
	rows, err := dbaccess.QueryRows(getPerkHistoryQuery, userId, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		entry := typeperks.PerkHistory{}
		err = rows.Scan(&entry.Id, &entry.UserId, &entry.PartyId, &entry.PerkId, &entry.Action, &entry.SourceEventId, &entry.CreatedDate)

		if err != nil {
			_ = rows.Close()
			return
		}

		history = append(history, entry)
	}

	dbaccess.LogDbResult(getPerkHistoryQuery, history, err)

	_ = rows.Close()
	return
}
