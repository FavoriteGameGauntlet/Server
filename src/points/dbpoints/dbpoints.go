package dbpoints

import (
	"FGG-Service/src/dbaccess"
	"FGG-Service/src/points/typepoints"
	"context"
)

type IDatabase interface {
	CreatePointTypeCommand(ctx context.Context, partyId int, name string, description string, startValue int, isPublic bool, isShared bool, minimum *int, maximum *int) (pointType typepoints.PointType, err error)
	ChangePointTypeCommand(ctx context.Context, partyId int, pointTypeId int, name string, description string, startValue int, isPublic bool, isShared bool, minimum *int, maximum *int) error
	RemovePointTypeCommand(ctx context.Context, partyId int, pointTypeId int) error
	GetPointTypeByNameCommand(ctx context.Context, partyId int, name string) (pointType typepoints.PointTypeInfo, err error)
	GetPointTypeCommand(ctx context.Context, partyId int, pointTypeId int) (pointType typepoints.PointTypeInfo, err error)
	GetPointTypesCommand(ctx context.Context, partyId int) (pointTypes []typepoints.PointTypeInfo, err error)
	CreateUserPointCommand(ctx context.Context, userId int, partyId int, pointTypeId int, value int) (point typepoints.UserPoint, err error)
	CreatePartyPointCommand(ctx context.Context, partyId int, pointTypeId int, value int) (point typepoints.PartyPoint, err error)
	GetUserPointCommand(ctx context.Context, userId int, partyId int, pointTypeId int) (point typepoints.UserPoint, err error)
	GetPartyPointCommand(ctx context.Context, partyId int, pointTypeId int) (point typepoints.PartyPoint, err error)
	GetUserPointsCommand(ctx context.Context, userId int, partyId int) (points []typepoints.UserPoint, err error)
	GetPartyPointsCommand(ctx context.Context, partyId int) (points []typepoints.PartyPoint, err error)
	ChangeUserPointValueCommand(ctx context.Context, userId int, partyId int, pointTypeId int, changeValue int) error
	ChangePartyPointValueCommand(ctx context.Context, partyId int, pointTypeId int, changeValue int) error
	CreateUserPointHistoryCommand(ctx context.Context, userId int, partyId int, pointTypeId int, actorUserId int, desiredChangeValue int, actualChangeValue int, finalValue int, sourceEventId int) (entry typepoints.UserPointHistoryEntry, err error)
	CreatePartyPointHistoryCommand(ctx context.Context, partyId int, pointTypeId int, actorUserId int, desiredChangeValue int, actualChangeValue int, finalValue int, sourceEventId int) (entry typepoints.PartyPointHistoryEntry, err error)
	GetUserPointHistoryCommand(ctx context.Context, userId int, partyId int) (history []typepoints.UserPointHistoryEntry, err error)
	GetPartyPointHistoryCommand(ctx context.Context, partyId int) (history []typepoints.PartyPointHistoryEntry, err error)
}

type Database struct {
}

var createPointTypeQuery = dbaccess.Query{Name: "CreatePointTypeQuery", SQL: `SELECT * FROM create_point_type($1::integer, $2::text, $3::text, $4::integer, $5::boolean, $6::boolean, $7::integer, $8::integer)`}

func (db *Database) CreatePointTypeCommand(ctx context.Context, partyId int, name string, description string, startValue int, isPublic bool, isShared bool, minimum *int, maximum *int) (pointType typepoints.PointType, err error) {
	row := dbaccess.QueryRow(ctx, createPointTypeQuery, partyId, name, description, startValue, isPublic, isShared, minimum, maximum)

	err = row.Scan(
		&pointType.Id,
		&pointType.PartyId,
		&pointType.Name,
		&pointType.Description,
		&pointType.StartValue,
		&pointType.IsPublic,
		&pointType.IsShared,
		&pointType.Minimum,
		&pointType.Maximum,
		&pointType.IsRemoved)

	dbaccess.LogDbResult(createPointTypeQuery, pointType, err)

	return
}

var changePointTypeQuery = dbaccess.Query{Name: "ChangePointTypeQuery", SQL: `SELECT change_point_type($1::integer, $2::integer, $3::text, $4::text, $5::integer, $6::boolean, $7::boolean, $8::integer, $9::integer)`}

func (db *Database) ChangePointTypeCommand(ctx context.Context, partyId int, pointTypeId int, name string, description string, startValue int, isPublic bool, isShared bool, minimum *int, maximum *int) error {
	_, err := dbaccess.Exec(ctx, changePointTypeQuery, partyId, pointTypeId, name, description, startValue, isPublic, isShared, minimum, maximum)

	dbaccess.LogDbResult(changePointTypeQuery, nil, err)

	return err
}

var removePointTypeQuery = dbaccess.Query{Name: "RemovePointTypeQuery", SQL: `SELECT remove_point_type($1::integer, $2::integer)`}

func (db *Database) RemovePointTypeCommand(ctx context.Context, partyId int, pointTypeId int) error {
	_, err := dbaccess.Exec(ctx, removePointTypeQuery, partyId, pointTypeId)

	dbaccess.LogDbResult(removePointTypeQuery, nil, err)

	return err
}

var getPointTypeByNameQuery = dbaccess.Query{Name: "GetPointTypeByNameQuery", SQL: `SELECT * FROM get_point_type_by_name($1::integer, $2::text)`}

func (db *Database) GetPointTypeByNameCommand(ctx context.Context, partyId int, name string) (pointType typepoints.PointTypeInfo, err error) {
	row := dbaccess.QueryRow(ctx, getPointTypeByNameQuery, partyId, name)

	err = row.Scan(
		&pointType.Id,
		&pointType.PartyId,
		&pointType.Name,
		&pointType.Description,
		&pointType.StartValue,
		&pointType.IsPublic,
		&pointType.IsShared,
		&pointType.Minimum,
		&pointType.Maximum)

	dbaccess.LogDbResult(getPointTypeByNameQuery, pointType, err)

	return
}

var getPointTypeQuery = dbaccess.Query{Name: "GetPointTypeQuery", SQL: `SELECT * FROM get_point_type($1::integer, $2::integer)`}

func (db *Database) GetPointTypeCommand(ctx context.Context, partyId int, pointTypeId int) (pointType typepoints.PointTypeInfo, err error) {
	row := dbaccess.QueryRow(ctx, getPointTypeQuery, partyId, pointTypeId)

	err = row.Scan(
		&pointType.Id,
		&pointType.PartyId,
		&pointType.Name,
		&pointType.Description,
		&pointType.StartValue,
		&pointType.IsPublic,
		&pointType.IsShared,
		&pointType.Minimum,
		&pointType.Maximum)

	dbaccess.LogDbResult(getPointTypeQuery, pointType, err)

	return
}

var getPointTypesQuery = dbaccess.Query{Name: "GetPointTypesQuery", SQL: `SELECT * FROM get_point_types($1::integer)`}

func (db *Database) GetPointTypesCommand(ctx context.Context, partyId int) (pointTypes []typepoints.PointTypeInfo, err error) {
	rows, err := dbaccess.QueryRows(ctx, getPointTypesQuery, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		pointType := typepoints.PointTypeInfo{}
		err = rows.Scan(
			&pointType.Id,
			&pointType.PartyId,
			&pointType.Name,
			&pointType.Description,
			&pointType.StartValue,
			&pointType.IsPublic,
			&pointType.IsShared,
			&pointType.Minimum,
			&pointType.Maximum)

		if err != nil {
			_ = rows.Close()
			return
		}

		pointTypes = append(pointTypes, pointType)
	}

	err = rows.Err()

	dbaccess.LogDbResult(getPointTypesQuery, pointTypes, err)

	_ = rows.Close()
	return
}

var createUserPointQuery = dbaccess.Query{Name: "CreateUserPointQuery", SQL: `SELECT * FROM create_user_point($1::integer, $2::integer, $3::integer, $4::integer)`}

func (db *Database) CreateUserPointCommand(ctx context.Context, userId int, partyId int, pointTypeId int, value int) (point typepoints.UserPoint, err error) {
	row := dbaccess.QueryRow(ctx, createUserPointQuery, userId, partyId, pointTypeId, value)

	err = row.Scan(&point.Id, &point.UserId, &point.PartyId, &point.PointTypeId, &point.Value)

	dbaccess.LogDbResult(createUserPointQuery, point, err)

	return
}

var createPartyPointQuery = dbaccess.Query{Name: "CreatePartyPointQuery", SQL: `SELECT * FROM create_party_point($1::integer, $2::integer, $3::integer)`}

func (db *Database) CreatePartyPointCommand(ctx context.Context, partyId int, pointTypeId int, value int) (point typepoints.PartyPoint, err error) {
	row := dbaccess.QueryRow(ctx, createPartyPointQuery, partyId, pointTypeId, value)

	err = row.Scan(&point.Id, &point.PartyId, &point.PointTypeId, &point.Value)

	dbaccess.LogDbResult(createPartyPointQuery, point, err)

	return
}

var getUserPointQuery = dbaccess.Query{Name: "GetUserPointQuery", SQL: `SELECT * FROM get_user_point($1::integer, $2::integer, $3::integer)`}

func (db *Database) GetUserPointCommand(ctx context.Context, userId int, partyId int, pointTypeId int) (point typepoints.UserPoint, err error) {
	row := dbaccess.QueryRow(ctx, getUserPointQuery, userId, partyId, pointTypeId)

	err = row.Scan(&point.Id, &point.UserId, &point.PartyId, &point.PointTypeId, &point.Value)

	dbaccess.LogDbResult(getUserPointQuery, point, err)

	return
}

var getPartyPointQuery = dbaccess.Query{Name: "GetPartyPointQuery", SQL: `SELECT * FROM get_party_point($1::integer, $2::integer)`}

func (db *Database) GetPartyPointCommand(ctx context.Context, partyId int, pointTypeId int) (point typepoints.PartyPoint, err error) {
	row := dbaccess.QueryRow(ctx, getPartyPointQuery, partyId, pointTypeId)

	err = row.Scan(&point.Id, &point.PartyId, &point.PointTypeId, &point.Value)

	dbaccess.LogDbResult(getPartyPointQuery, point, err)

	return
}

var getUserPointsQuery = dbaccess.Query{Name: "GetUserPointsQuery", SQL: `SELECT * FROM get_user_points($1::integer, $2::integer)`}

func (db *Database) GetUserPointsCommand(ctx context.Context, userId int, partyId int) (points []typepoints.UserPoint, err error) {
	rows, err := dbaccess.QueryRows(ctx, getUserPointsQuery, userId, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		point := typepoints.UserPoint{}
		err = rows.Scan(&point.Id, &point.UserId, &point.PartyId, &point.PointTypeId, &point.Value)

		if err != nil {
			_ = rows.Close()
			return
		}

		points = append(points, point)
	}

	err = rows.Err()

	dbaccess.LogDbResult(getUserPointsQuery, points, err)

	_ = rows.Close()
	return
}

var getPartyPointsQuery = dbaccess.Query{Name: "GetPartyPointsQuery", SQL: `SELECT * FROM get_party_points($1::integer)`}

func (db *Database) GetPartyPointsCommand(ctx context.Context, partyId int) (points []typepoints.PartyPoint, err error) {
	rows, err := dbaccess.QueryRows(ctx, getPartyPointsQuery, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		point := typepoints.PartyPoint{}
		err = rows.Scan(&point.Id, &point.PartyId, &point.PointTypeId, &point.Value)

		if err != nil {
			_ = rows.Close()
			return
		}

		points = append(points, point)
	}

	err = rows.Err()

	dbaccess.LogDbResult(getPartyPointsQuery, points, err)

	_ = rows.Close()
	return
}

var changeUserPointValueQuery = dbaccess.Query{Name: "ChangeUserPointValueQuery", SQL: `SELECT change_user_point_value($1::integer, $2::integer, $3::integer, $4::integer)`}

func (db *Database) ChangeUserPointValueCommand(ctx context.Context, userId int, partyId int, pointTypeId int, changeValue int) error {
	_, err := dbaccess.Exec(ctx, changeUserPointValueQuery, userId, partyId, pointTypeId, changeValue)

	dbaccess.LogDbResult(changeUserPointValueQuery, nil, err)

	return err
}

var changePartyPointValueQuery = dbaccess.Query{Name: "ChangePartyPointValueQuery", SQL: `SELECT change_party_point_value($1::integer, $2::integer, $3::integer)`}

func (db *Database) ChangePartyPointValueCommand(ctx context.Context, partyId int, pointTypeId int, changeValue int) error {
	_, err := dbaccess.Exec(ctx, changePartyPointValueQuery, partyId, pointTypeId, changeValue)

	dbaccess.LogDbResult(changePartyPointValueQuery, nil, err)

	return err
}

var createUserPointHistoryQuery = dbaccess.Query{Name: "CreateUserPointHistoryQuery", SQL: `SELECT * FROM create_user_point_history($1::integer, $2::integer, $3::integer, $4::integer, $5::integer, $6::integer, $7::integer, $8::integer)`}

func (db *Database) CreateUserPointHistoryCommand(ctx context.Context, userId int, partyId int, pointTypeId int, actorUserId int, desiredChangeValue int, actualChangeValue int, finalValue int, sourceEventId int) (entry typepoints.UserPointHistoryEntry, err error) {
	row := dbaccess.QueryRow(ctx, createUserPointHistoryQuery, userId, partyId, pointTypeId, actorUserId, desiredChangeValue, actualChangeValue, finalValue, sourceEventId)

	err = row.Scan(
		&entry.Id,
		&entry.UserId,
		&entry.PartyId,
		&entry.PointTypeId,
		&entry.ActorUserId,
		&entry.DesiredChangeValue,
		&entry.ActualChangeValue,
		&entry.FinalValue,
		&entry.SourceEventId,
		&entry.ChangedDate)

	dbaccess.LogDbResult(createUserPointHistoryQuery, entry, err)

	return
}

var createPartyPointHistoryQuery = dbaccess.Query{Name: "CreatePartyPointHistoryQuery", SQL: `SELECT * FROM create_party_point_history($1::integer, $2::integer, $3::integer, $4::integer, $5::integer, $6::integer, $7::integer)`}

func (db *Database) CreatePartyPointHistoryCommand(ctx context.Context, partyId int, pointTypeId int, actorUserId int, desiredChangeValue int, actualChangeValue int, finalValue int, sourceEventId int) (entry typepoints.PartyPointHistoryEntry, err error) {
	row := dbaccess.QueryRow(ctx, createPartyPointHistoryQuery, partyId, pointTypeId, actorUserId, desiredChangeValue, actualChangeValue, finalValue, sourceEventId)

	err = row.Scan(
		&entry.Id,
		&entry.PartyId,
		&entry.PointTypeId,
		&entry.ActorUserId,
		&entry.DesiredChangeValue,
		&entry.ActualChangeValue,
		&entry.FinalValue,
		&entry.SourceEventId,
		&entry.ChangedDate)

	dbaccess.LogDbResult(createPartyPointHistoryQuery, entry, err)

	return
}

var getUserPointHistoryQuery = dbaccess.Query{Name: "GetUserPointHistoryQuery", SQL: `SELECT * FROM get_user_point_history($1::integer, $2::integer)`}

func (db *Database) GetUserPointHistoryCommand(ctx context.Context, userId int, partyId int) (history []typepoints.UserPointHistoryEntry, err error) {
	rows, err := dbaccess.QueryRows(ctx, getUserPointHistoryQuery, userId, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		entry := typepoints.UserPointHistoryEntry{}
		err = rows.Scan(
			&entry.Id,
			&entry.UserId,
			&entry.PartyId,
			&entry.PointTypeId,
			&entry.ActorUserId,
			&entry.DesiredChangeValue,
			&entry.ActualChangeValue,
			&entry.FinalValue,
			&entry.SourceEventId,
			&entry.ChangedDate)

		if err != nil {
			_ = rows.Close()
			return
		}

		history = append(history, entry)
	}

	err = rows.Err()

	dbaccess.LogDbResult(getUserPointHistoryQuery, history, err)

	_ = rows.Close()
	return
}

var getPartyPointHistoryQuery = dbaccess.Query{Name: "GetPartyPointHistoryQuery", SQL: `SELECT * FROM get_party_point_history($1::integer)`}

func (db *Database) GetPartyPointHistoryCommand(ctx context.Context, partyId int) (history []typepoints.PartyPointHistoryEntry, err error) {
	rows, err := dbaccess.QueryRows(ctx, getPartyPointHistoryQuery, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		entry := typepoints.PartyPointHistoryEntry{}
		err = rows.Scan(
			&entry.Id,
			&entry.PartyId,
			&entry.PointTypeId,
			&entry.ActorUserId,
			&entry.DesiredChangeValue,
			&entry.ActualChangeValue,
			&entry.FinalValue,
			&entry.SourceEventId,
			&entry.ChangedDate)

		if err != nil {
			_ = rows.Close()
			return
		}

		history = append(history, entry)
	}

	err = rows.Err()

	dbaccess.LogDbResult(getPartyPointHistoryQuery, history, err)

	_ = rows.Close()
	return
}
