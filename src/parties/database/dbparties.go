package dbparties

import (
	"FGG-Service/src/dbaccess"
	"FGG-Service/src/parties/types"
)

type IDatabase interface {
	CreatePartyCommand(name string) (party typeparties.Party, err error)
	GetPartyCommand(partyId int) (party typeparties.Party, err error)
	GetPartiesCommand() (parties []typeparties.Party, err error)
	ChangePartyNameCommand(partyId int, name string) error
	CreateMemberCommand(userId int, partyId int, displayName *string, isAdmin bool) (member typeparties.Member, err error)
	GetMemberCommand(userId int, partyId int) (member typeparties.MemberWithLogin, err error)
	GetMembersCommand(partyId int) (members []typeparties.MemberWithLogin, err error)
	ChangeMemberAdminStatusCommand(userId int, partyId int, isAdmin bool) error
	ChangeMemberDisplayNameCommand(userId int, partyId int, displayName string) error
	RemoveMemberCommand(userId int, partyId int) error
}

type Database struct{}

var createPartyQuery = dbaccess.Query{Name: "CreatePartyQuery", SQL: `SELECT * FROM create_party($1::text)`}

func (db *Database) CreatePartyCommand(name string) (party typeparties.Party, err error) {
	row := dbaccess.QueryRow(createPartyQuery, name)

	err = row.Scan(&party.Id, &party.Name, &party.CreatedDate)

	dbaccess.LogDbResult(createPartyQuery, party, err)

	return
}

var getPartyQuery = dbaccess.Query{Name: "GetPartyQuery", SQL: `SELECT * FROM get_party($1::integer)`}

func (db *Database) GetPartyCommand(partyId int) (party typeparties.Party, err error) {
	row := dbaccess.QueryRow(getPartyQuery, partyId)

	err = row.Scan(&party.Id, &party.Name, &party.CreatedDate)

	dbaccess.LogDbResult(getPartyQuery, party, err)

	return
}

var getPartiesQuery = dbaccess.Query{Name: "GetPartiesQuery", SQL: `SELECT * FROM get_parties()`}

func (db *Database) GetPartiesCommand() (parties []typeparties.Party, err error) {
	rows, err := dbaccess.QueryRows(getPartiesQuery)

	if err != nil {
		return
	}

	for rows.Next() {
		party := typeparties.Party{}
		err = rows.Scan(&party.Id, &party.Name, &party.CreatedDate)

		if err != nil {
			_ = rows.Close()
			return
		}

		parties = append(parties, party)
	}

	dbaccess.LogDbResult(getPartiesQuery, parties, err)

	_ = rows.Close()
	return
}

var changePartyNameQuery = dbaccess.Query{Name: "ChangePartyNameQuery", SQL: `SELECT change_party_name($1::integer, $2::text)`}

func (db *Database) ChangePartyNameCommand(partyId int, name string) error {
	_, err := dbaccess.Exec(changePartyNameQuery, partyId, name)

	dbaccess.LogDbResult(changePartyNameQuery, nil, err)

	return err
}

var createMemberQuery = dbaccess.Query{Name: "CreateMemberQuery", SQL: `SELECT * FROM create_member($1::integer, $2::integer, $3::text, $4::boolean)`}

func (db *Database) CreateMemberCommand(userId int, partyId int, displayName *string, isAdmin bool) (member typeparties.Member, err error) {
	row := dbaccess.QueryRow(createMemberQuery, userId, partyId, displayName, isAdmin)

	err = row.Scan(
		&member.Id,
		&member.UserId,
		&member.PartyId,
		&member.DisplayName,
		&member.IsAdmin,
		&member.JoinedDate,
		&member.LeftDate)

	dbaccess.LogDbResult(createMemberQuery, member, err)

	return
}

var getMemberQuery = dbaccess.Query{Name: "GetMemberQuery", SQL: `SELECT * FROM get_member($1::integer, $2::integer)`}

func (db *Database) GetMemberCommand(userId int, partyId int) (member typeparties.MemberWithLogin, err error) {
	row := dbaccess.QueryRow(getMemberQuery, userId, partyId)

	err = row.Scan(
		&member.Id,
		&member.UserId,
		&member.PartyId,
		&member.Login,
		&member.DisplayName,
		&member.IsAdmin,
		&member.JoinedDate,
		&member.LeftDate)

	dbaccess.LogDbResult(getMemberQuery, member, err)

	return
}

var getMembersQuery = dbaccess.Query{Name: "GetMembersQuery", SQL: `SELECT * FROM get_members($1::integer)`}

func (db *Database) GetMembersCommand(partyId int) (members []typeparties.MemberWithLogin, err error) {
	rows, err := dbaccess.QueryRows(getMembersQuery, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		member := typeparties.MemberWithLogin{}
		err = rows.Scan(
			&member.Id,
			&member.UserId,
			&member.PartyId,
			&member.Login,
			&member.DisplayName,
			&member.IsAdmin,
			&member.JoinedDate,
			&member.LeftDate)

		if err != nil {
			_ = rows.Close()
			return
		}

		members = append(members, member)
	}

	dbaccess.LogDbResult(getMembersQuery, members, err)

	_ = rows.Close()
	return
}

var changeMemberAdminStatusQuery = dbaccess.Query{Name: "ChangeMemberAdminStatusQuery", SQL: `SELECT change_member_admin_status($1::integer, $2::integer, $3::boolean)`}

func (db *Database) ChangeMemberAdminStatusCommand(userId int, partyId int, isAdmin bool) error {
	_, err := dbaccess.Exec(changeMemberAdminStatusQuery, userId, partyId, isAdmin)

	dbaccess.LogDbResult(changeMemberAdminStatusQuery, nil, err)

	return err
}

var changeMemberDisplayNameQuery = dbaccess.Query{Name: "ChangeMemberDisplayNameQuery", SQL: `SELECT change_member_display_name($1::integer, $2::integer, $3::text)`}

func (db *Database) ChangeMemberDisplayNameCommand(userId int, partyId int, displayName string) error {
	_, err := dbaccess.Exec(changeMemberDisplayNameQuery, userId, partyId, displayName)

	dbaccess.LogDbResult(changeMemberDisplayNameQuery, nil, err)

	return err
}

var removeMemberQuery = dbaccess.Query{Name: "RemoveMemberQuery", SQL: `SELECT remove_member($1::integer, $2::integer)`}

func (db *Database) RemoveMemberCommand(userId int, partyId int) error {
	_, err := dbaccess.Exec(removeMemberQuery, userId, partyId)

	dbaccess.LogDbResult(removeMemberQuery, nil, err)

	return err
}
