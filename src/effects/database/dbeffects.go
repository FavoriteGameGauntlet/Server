package dbeffects

import (
	"FGG-Service/src/changes/types"
	"FGG-Service/src/dbaccess"
	"FGG-Service/src/effects/types"
	"database/sql"
	"encoding/json"
	"time"
)

type IDatabase interface {
	CreateEffectCommand(partyId int, name string, description string, useCount int, duration *time.Duration, change typechanges.Change) (effect typeeffects.EffectWithChange, err error)
	GetEffectCommand(partyId int, effectId int) (effect typeeffects.EffectWithChange, err error)
	GetActualEffectsCommand(partyId int) (effects []typeeffects.Effect, err error)
	GetRemovedEffectsCommand(partyId int) (effects []typeeffects.Effect, err error)
	RemoveEffectCommand(partyId int, effectId int) error
	CreateUserEffectCommand(userId int, partyId int, effectId int, actorUserId int, sourceEventId int) (userEffect typeeffects.UserEffect, err error)
	GetUserEffectCommand(userId int, partyId int, effectId int) (userEffect typeeffects.UserEffectDetail, err error)
	GetUserEffectsCommand(userId int, partyId int) (userEffects []typeeffects.UserEffectDetail, err error)
	ChangeUserEffectUsesLeftCommand(userId int, partyId int, effectId int, usesLeft int, actorUserId int, sourceEventId *int) (historyEventId int, err error)
	DeleteUserEffectCommand(userId int, partyId int, effectId int, actorUserId int, sourceEventId *int) error
	DeleteEndedUserEffectsCommand() (deleted []typeeffects.EndedUserEffect, err error)
	GetEffectHistoryCommand(userId int, partyId int) (history []typeeffects.EffectHistory, err error)
	GetUserEffectPointModifiersJsonbCommand(partyId int, userEffectId int) (modifiers []typeeffects.PointModifier, err error)
}

type Database struct{}

func scanNullableInterval(raw sql.NullString) (*time.Duration, error) {
	if !raw.Valid {
		return nil, nil
	}

	duration, err := dbaccess.ScanInterval(raw.String)
	if err != nil {
		return nil, err
	}

	return &duration, nil
}

func durationArg(duration *time.Duration) any {
	if duration == nil {
		return nil
	}

	return *duration
}

var createEffectQuery = dbaccess.Query{Name: "CreateEffectQuery", SQL: `SELECT * FROM create_effect($1::integer, $2::text, $3::text, $4::integer, $5::interval, $6::jsonb)`}

func (db *Database) CreateEffectCommand(partyId int, name string, description string, useCount int, duration *time.Duration, change typechanges.Change) (effect typeeffects.EffectWithChange, err error) {
	changeJson, err := json.Marshal(change)
	if err != nil {
		return
	}

	row := dbaccess.QueryRow(createEffectQuery, partyId, name, description, useCount, durationArg(duration), changeJson)

	var durationRaw sql.NullString
	var changeRaw []byte
	err = row.Scan(&effect.Id, &effect.PartyId, &effect.Name, &effect.Description, &effect.UseCount, &durationRaw, &changeRaw)

	if err == nil {
		effect.Duration, err = scanNullableInterval(durationRaw)
	}
	if err == nil {
		err = json.Unmarshal(changeRaw, &effect.Change)
	}

	dbaccess.LogDbResult(createEffectQuery, effect, err)

	return
}

var getEffectQuery = dbaccess.Query{Name: "GetEffectQuery", SQL: `SELECT * FROM get_effect($1::integer, $2::integer)`}

func (db *Database) GetEffectCommand(partyId int, effectId int) (effect typeeffects.EffectWithChange, err error) {
	row := dbaccess.QueryRow(getEffectQuery, partyId, effectId)

	var durationRaw sql.NullString
	var changeRaw []byte
	err = row.Scan(&effect.Id, &effect.PartyId, &effect.Name, &effect.Description, &effect.UseCount, &durationRaw, &changeRaw)

	if err == nil {
		effect.Duration, err = scanNullableInterval(durationRaw)
	}
	if err == nil {
		err = json.Unmarshal(changeRaw, &effect.Change)
	}

	dbaccess.LogDbResult(getEffectQuery, effect, err)

	return
}

func scanEffects(q dbaccess.Query, partyId int) (effects []typeeffects.Effect, err error) {
	rows, err := dbaccess.QueryRows(q, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		effect := typeeffects.Effect{}
		var durationRaw sql.NullString
		err = rows.Scan(&effect.Id, &effect.PartyId, &effect.Name, &effect.Description, &effect.UseCount, &durationRaw)

		if err == nil {
			effect.Duration, err = scanNullableInterval(durationRaw)
		}

		if err != nil {
			_ = rows.Close()
			return
		}

		effects = append(effects, effect)
	}

	dbaccess.LogDbResult(q, effects, err)

	_ = rows.Close()
	return
}

var getActualEffectsQuery = dbaccess.Query{Name: "GetActualEffectsQuery", SQL: `SELECT * FROM get_actual_effects($1::integer)`}

func (db *Database) GetActualEffectsCommand(partyId int) (effects []typeeffects.Effect, err error) {
	return scanEffects(getActualEffectsQuery, partyId)
}

var getRemovedEffectsQuery = dbaccess.Query{Name: "GetRemovedEffectsQuery", SQL: `SELECT * FROM get_removed_effects($1::integer)`}

func (db *Database) GetRemovedEffectsCommand(partyId int) (effects []typeeffects.Effect, err error) {
	return scanEffects(getRemovedEffectsQuery, partyId)
}

var removeEffectQuery = dbaccess.Query{Name: "RemoveEffectQuery", SQL: `SELECT remove_effect($1::integer, $2::integer)`}

func (db *Database) RemoveEffectCommand(partyId int, effectId int) error {
	_, err := dbaccess.Exec(removeEffectQuery, partyId, effectId)

	dbaccess.LogDbResult(removeEffectQuery, nil, err)

	return err
}

var createUserEffectQuery = dbaccess.Query{Name: "CreateUserEffectQuery", SQL: `SELECT * FROM create_user_effect($1::integer, $2::integer, $3::integer, $4::integer, $5::integer)`}

func (db *Database) CreateUserEffectCommand(userId int, partyId int, effectId int, actorUserId int, sourceEventId int) (userEffect typeeffects.UserEffect, err error) {
	row := dbaccess.QueryRow(createUserEffectQuery, userId, partyId, effectId, actorUserId, sourceEventId)

	err = row.Scan(&userEffect.Id, &userEffect.UserId, &userEffect.PartyId, &userEffect.EffectId, &userEffect.UsesLeft, &userEffect.EffectHistoryId)

	dbaccess.LogDbResult(createUserEffectQuery, userEffect, err)

	return
}

func scanUserEffectDetail(row *sql.Row) (userEffect typeeffects.UserEffectDetail, err error) {
	var durationRaw sql.NullString
	var modifiersRaw []byte
	err = row.Scan(
		&userEffect.Id,
		&userEffect.UserId,
		&userEffect.PartyId,
		&userEffect.EffectId,
		&userEffect.Name,
		&userEffect.Description,
		&userEffect.UseCount,
		&userEffect.UsesLeft,
		&durationRaw,
		&userEffect.StartedDate,
		&modifiersRaw)

	if err == nil {
		userEffect.Duration, err = scanNullableInterval(durationRaw)
	}
	if err == nil {
		err = json.Unmarshal(modifiersRaw, &userEffect.Modifiers)
	}

	return
}

var getUserEffectQuery = dbaccess.Query{Name: "GetUserEffectQuery", SQL: `SELECT * FROM get_user_effect($1::integer, $2::integer, $3::integer)`}

func (db *Database) GetUserEffectCommand(userId int, partyId int, effectId int) (userEffect typeeffects.UserEffectDetail, err error) {
	row := dbaccess.QueryRow(getUserEffectQuery, userId, partyId, effectId)

	userEffect, err = scanUserEffectDetail(row)

	dbaccess.LogDbResult(getUserEffectQuery, userEffect, err)

	return
}

var getUserEffectsQuery = dbaccess.Query{Name: "GetUserEffectsQuery", SQL: `SELECT * FROM get_user_effects($1::integer, $2::integer)`}

func (db *Database) GetUserEffectsCommand(userId int, partyId int) (userEffects []typeeffects.UserEffectDetail, err error) {
	rows, err := dbaccess.QueryRows(getUserEffectsQuery, userId, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		userEffect := typeeffects.UserEffectDetail{}
		var durationRaw sql.NullString
		var modifiersRaw []byte
		err = rows.Scan(
			&userEffect.Id,
			&userEffect.UserId,
			&userEffect.PartyId,
			&userEffect.EffectId,
			&userEffect.Name,
			&userEffect.Description,
			&userEffect.UseCount,
			&userEffect.UsesLeft,
			&durationRaw,
			&userEffect.StartedDate,
			&modifiersRaw)

		if err == nil {
			userEffect.Duration, err = scanNullableInterval(durationRaw)
		}
		if err == nil {
			err = json.Unmarshal(modifiersRaw, &userEffect.Modifiers)
		}

		if err != nil {
			_ = rows.Close()
			return
		}

		userEffects = append(userEffects, userEffect)
	}

	dbaccess.LogDbResult(getUserEffectsQuery, userEffects, err)

	_ = rows.Close()
	return
}

var changeUserEffectUsesLeftQuery = dbaccess.Query{Name: "ChangeUserEffectUsesLeftQuery", SQL: `SELECT change_user_effect_uses_left($1::integer, $2::integer, $3::integer, $4::integer, $5::integer, $6::integer)`}

func (db *Database) ChangeUserEffectUsesLeftCommand(userId int, partyId int, effectId int, usesLeft int, actorUserId int, sourceEventId *int) (historyEventId int, err error) {
	row := dbaccess.QueryRow(changeUserEffectUsesLeftQuery, userId, partyId, effectId, usesLeft, actorUserId, sourceEventId)

	err = row.Scan(&historyEventId)

	dbaccess.LogDbResult(changeUserEffectUsesLeftQuery, historyEventId, err)

	return
}

var deleteUserEffectQuery = dbaccess.Query{Name: "DeleteUserEffectQuery", SQL: `SELECT delete_user_effect($1::integer, $2::integer, $3::integer, $4::integer, $5::integer)`}

func (db *Database) DeleteUserEffectCommand(userId int, partyId int, effectId int, actorUserId int, sourceEventId *int) error {
	_, err := dbaccess.Exec(deleteUserEffectQuery, userId, partyId, effectId, actorUserId, sourceEventId)

	dbaccess.LogDbResult(deleteUserEffectQuery, nil, err)

	return err
}

var deleteEndedUserEffectsQuery = dbaccess.Query{Name: "DeleteEndedUserEffectsQuery", SQL: `SELECT * FROM delete_ended_user_effects()`}

func (db *Database) DeleteEndedUserEffectsCommand() (deleted []typeeffects.EndedUserEffect, err error) {
	rows, err := dbaccess.QueryRows(deleteEndedUserEffectsQuery)

	if err != nil {
		return
	}

	for rows.Next() {
		entry := typeeffects.EndedUserEffect{}
		err = rows.Scan(&entry.Id, &entry.UserId, &entry.PartyId, &entry.EffectId, &entry.UsesLeft)

		if err != nil {
			_ = rows.Close()
			return
		}

		deleted = append(deleted, entry)
	}

	dbaccess.LogDbResult(deleteEndedUserEffectsQuery, deleted, err)

	_ = rows.Close()
	return
}

var getEffectHistoryQuery = dbaccess.Query{Name: "GetEffectHistoryQuery", SQL: `SELECT * FROM get_effect_history($1::integer, $2::integer)`}

func (db *Database) GetEffectHistoryCommand(userId int, partyId int) (history []typeeffects.EffectHistory, err error) {
	rows, err := dbaccess.QueryRows(getEffectHistoryQuery, userId, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		entry := typeeffects.EffectHistory{}
		var durationRaw sql.NullString
		err = rows.Scan(
			&entry.Id,
			&entry.UserId,
			&entry.PartyId,
			&entry.EffectId,
			&entry.Name,
			&entry.Description,
			&entry.UseCount,
			&durationRaw,
			&entry.Action,
			&entry.UsesLeft,
			&entry.SourceEventId,
			&entry.CreatedDate)

		if err == nil {
			entry.Duration, err = scanNullableInterval(durationRaw)
		}

		if err != nil {
			_ = rows.Close()
			return
		}

		history = append(history, entry)
	}

	dbaccess.LogDbResult(getEffectHistoryQuery, history, err)

	_ = rows.Close()
	return
}

var getUserEffectPointModifiersJsonbQuery = dbaccess.Query{Name: "GetUserEffectPointModifiersJsonbQuery", SQL: `SELECT get_user_effect_point_modifiers_jsonb($1::integer, $2::integer)`}

func (db *Database) GetUserEffectPointModifiersJsonbCommand(partyId int, userEffectId int) (modifiers []typeeffects.PointModifier, err error) {
	row := dbaccess.QueryRow(getUserEffectPointModifiersJsonbQuery, partyId, userEffectId)

	var raw []byte
	err = row.Scan(&raw)
	if err == nil {
		err = json.Unmarshal(raw, &modifiers)
	}

	dbaccess.LogDbResult(getUserEffectPointModifiersJsonbQuery, modifiers, err)

	return
}
