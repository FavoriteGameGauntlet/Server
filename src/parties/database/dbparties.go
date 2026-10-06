package dbparties

import (
	"FGG-Service/src/dbaccess"
	"FGG-Service/src/parties/types"
	"context"
)

type IDatabase interface {
	CreatePartyCommand(ctx context.Context, name string) (party typeparties.Party, err error)
	GetPartyCommand(ctx context.Context, partyId int) (party typeparties.Party, err error)
	GetUserPartiesCommand(ctx context.Context, userId int) (parties []typeparties.Party, err error)
	ChangePartyNameCommand(ctx context.Context, partyId int, name string) error
	CreateMemberCommand(ctx context.Context, userId int, partyId int, displayName *string, isAdmin bool) (member typeparties.Member, err error)
	GetMemberCommand(ctx context.Context, userId int, partyId int) (member typeparties.MemberWithLogin, err error)
	GetMembersCommand(ctx context.Context, partyId int) (members []typeparties.MemberWithLogin, err error)
	ChangeMemberAdminStatusCommand(ctx context.Context, userId int, partyId int, isAdmin bool) error
	ChangeMemberDisplayNameCommand(ctx context.Context, userId int, partyId int, displayName *string) error
	RemoveMemberCommand(ctx context.Context, userId int, partyId int) error
}

type Database struct{}

var createPartyQuery = dbaccess.Query{Name: "CreatePartyQuery", SQL: `SELECT * FROM create_party($1::text)`}

func (db *Database) CreatePartyCommand(ctx context.Context, name string) (party typeparties.Party, err error) {
	row := dbaccess.QueryRow(ctx, createPartyQuery, name)

	err = row.Scan(&party.Id, &party.Name, &party.CreatedDate)

	dbaccess.LogDbResult(createPartyQuery, party, err)

	return
}

var getPartyQuery = dbaccess.Query{Name: "GetPartyQuery", SQL: `SELECT * FROM get_party($1::integer)`}

func (db *Database) GetPartyCommand(ctx context.Context, partyId int) (party typeparties.Party, err error) {
	row := dbaccess.QueryRow(ctx, getPartyQuery, partyId)

	err = row.Scan(&party.Id, &party.Name, &party.CreatedDate)

	dbaccess.LogDbResult(getPartyQuery, party, err)

	return
}

var getUserPartiesQuery = dbaccess.Query{Name: "GetUserPartiesQuery", SQL: `SELECT * FROM get_user_parties($1::integer)`}

func (db *Database) GetUserPartiesCommand(ctx context.Context, userId int) (parties []typeparties.Party, err error) {
	rows, err := dbaccess.QueryRows(ctx, getUserPartiesQuery, userId)

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

	err = rows.Err()

	dbaccess.LogDbResult(getUserPartiesQuery, parties, err)

	_ = rows.Close()
	return
}

var changePartyNameQuery = dbaccess.Query{Name: "ChangePartyNameQuery", SQL: `SELECT change_party_name($1::integer, $2::text)`}

func (db *Database) ChangePartyNameCommand(ctx context.Context, partyId int, name string) error {
	_, err := dbaccess.Exec(ctx, changePartyNameQuery, partyId, name)

	dbaccess.LogDbResult(changePartyNameQuery, nil, err)

	return err
}

var createMemberQuery = dbaccess.Query{Name: "CreateMemberQuery", SQL: `SELECT * FROM create_member($1::integer, $2::integer, $3::text, $4::boolean)`}

func (db *Database) CreateMemberCommand(ctx context.Context, userId int, partyId int, displayName *string, isAdmin bool) (member typeparties.Member, err error) {
	row := dbaccess.QueryRow(ctx, createMemberQuery, userId, partyId, displayName, isAdmin)

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

func (db *Database) GetMemberCommand(ctx context.Context, userId int, partyId int) (member typeparties.MemberWithLogin, err error) {
	row := dbaccess.QueryRow(ctx, getMemberQuery, userId, partyId)

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

func (db *Database) GetMembersCommand(ctx context.Context, partyId int) (members []typeparties.MemberWithLogin, err error) {
	rows, err := dbaccess.QueryRows(ctx, getMembersQuery, partyId)

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

	err = rows.Err()

	dbaccess.LogDbResult(getMembersQuery, members, err)

	_ = rows.Close()
	return
}

var changeMemberAdminStatusQuery = dbaccess.Query{Name: "ChangeMemberAdminStatusQuery", SQL: `SELECT change_member_admin_status($1::integer, $2::integer, $3::boolean)`}

func (db *Database) ChangeMemberAdminStatusCommand(ctx context.Context, userId int, partyId int, isAdmin bool) error {
	_, err := dbaccess.Exec(ctx, changeMemberAdminStatusQuery, userId, partyId, isAdmin)

	dbaccess.LogDbResult(changeMemberAdminStatusQuery, nil, err)

	return err
}

var changeMemberDisplayNameQuery = dbaccess.Query{Name: "ChangeMemberDisplayNameQuery", SQL: `SELECT change_member_display_name($1::integer, $2::integer, $3::text)`}

func (db *Database) ChangeMemberDisplayNameCommand(ctx context.Context, userId int, partyId int, displayName *string) error {
	_, err := dbaccess.Exec(ctx, changeMemberDisplayNameQuery, userId, partyId, displayName)

	dbaccess.LogDbResult(changeMemberDisplayNameQuery, nil, err)

	return err
}

var removeMemberQuery = dbaccess.Query{Name: "RemoveMemberQuery", SQL: `SELECT remove_member($1::integer, $2::integer)`}

func (db *Database) RemoveMemberCommand(ctx context.Context, userId int, partyId int) error {
	_, err := dbaccess.Exec(ctx, removeMemberQuery, userId, partyId)

	dbaccess.LogDbResult(removeMemberQuery, nil, err)

	return err
}
