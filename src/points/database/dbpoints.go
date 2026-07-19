package dbpoints

import (
	"FGG-Service/src/dbaccess"
	"FGG-Service/src/points/type"
)

type IDatabase interface {
	ChangeAvailableRollsCommand(userId int, changeValue int) error
	GetExperiencePointsCommand(userId int) (points int, err error)
	ChangeExperiencePointsCommand(userId int, changeValue int) error
	GetFreePointsCommand(userId int) (points int, err error)
	ChangeFreePointsCommand(userId int, changeValue int) error
	AddFreePointHistoryCommand(
		userId int,
		sourceUserId int,
		changeSource string,
		changeValue int,
		actualChangeValue int,
		finalValue int,
		wheelEffectId *int) error
	GetFreePointHistoryCommand(userId int) (history typepoints.FreePointChangeHistories, err error)
	GetTerritoryHoursCommand(userId int) (points int, err error)
	ChangeTerritoryHoursCommand(userId int, changeValue int) error
	GetTerritoryPointsCommand(userId int) (points int, err error)
	ChangeTerritoryPointsCommand(userId int, changeValue int) error
	AddTerritoryPointHistoryCommand(
		userId int,
		sourceUserId int,
		changeSource string,
		changeValue int,
		actualChangeValue int,
		finalValue int) error
	GetTerritoryPointHistoryCommand(userId int) (history typepoints.TerritoryPointChangeHistories, err error)
	GetPointInfoCommand(userId int) (info typepoints.PointInfo, err error)
	GetAllPointInfoCommand() (infos typepoints.PointInfoByLogins, err error)

	CreatePointTypeCommand(partyId int, name string, description string, startValue int, isPublic bool, isShared bool, minimum int, maximum int) (pointType typepoints.PointType, err error)
	ChangePointTypeCommand(partyId int, pointTypeId int, name string, description string, isPublic bool, isShared bool, minimum int, maximum int) error
	RemovePointTypeCommand(partyId int, pointTypeId int) error
	DoesPointTypeExistCommand(partyId int, name string) (doesExist bool, err error)
	GetPointTypeCommand(partyId int, pointTypeId int) (pointType typepoints.PointTypeInfo, err error)
	GetPointTypesCommand(partyId int) (pointTypes []typepoints.PointTypeInfo, err error)
	CreateUserPointCommand(userId int, partyId int, pointTypeId int, value int) (point typepoints.UserPoint, err error)
	CreatePartyPointCommand(partyId int, pointTypeId int, value int) (point typepoints.PartyPoint, err error)
	GetUserPointCommand(userId int, partyId int, pointTypeId int) (point typepoints.UserPoint, err error)
	GetPartyPointCommand(partyId int, pointTypeId int) (point typepoints.PartyPoint, err error)
	GetUserPointsCommand(userId int, partyId int) (points []typepoints.UserPoint, err error)
	GetPartyPointsCommand(partyId int) (points []typepoints.PartyPoint, err error)
	ChangeUserPointValueCommand(userId int, partyId int, pointTypeId int, changeValue int) error
	ChangePartyPointValueCommand(partyId int, pointTypeId int, changeValue int) error
	CreateUserPointHistoryCommand(userId int, partyId int, pointTypeId int, sourceUserId int, desiredChangeValue int, actualChangeValue int, finalValue int, sourceEventId int) (entry typepoints.UserPointHistoryEntry, err error)
	CreatePartyPointHistoryCommand(partyId int, pointTypeId int, sourceUserId int, desiredChangeValue int, actualChangeValue int, finalValue int, sourceEventId int) (entry typepoints.PartyPointHistoryEntry, err error)
	GetUserPointHistoryCommand(userId int, partyId int) (history []typepoints.UserPointHistoryEntry, err error)
	GetPartyPointHistoryCommand(partyId int) (history []typepoints.PartyPointHistoryEntry, err error)
}

type Database struct {
}

var changeAvailableRollsQuery = dbaccess.Query{Name: "ChangeAvailableRollsQuery", SQL: `SELECT change_available_rolls($1::integer, $2::integer)`}

func (db *Database) ChangeAvailableRollsCommand(userId int, changeValue int) error {
	_, err := dbaccess.Exec(changeAvailableRollsQuery, userId, changeValue)

	dbaccess.LogDbResult(changeAvailableRollsQuery, nil, err)

	return err
}

var changeExperiencePointsQuery = dbaccess.Query{Name: "ChangeExperiencePointsQuery", SQL: `SELECT change_experience_points($1::integer, $2::integer)`}

func (db *Database) ChangeExperiencePointsCommand(userId int, changeValue int) error {
	_, err := dbaccess.Exec(changeExperiencePointsQuery, userId, changeValue)

	dbaccess.LogDbResult(changeExperiencePointsQuery, nil, err)

	return err
}

var getExperiencePointsQuery = dbaccess.Query{Name: "ExperiencePointsQuery", SQL: `SELECT * FROM get_experience_points($1::integer)`}

func (db *Database) GetExperiencePointsCommand(userId int) (points int, err error) {
	row := dbaccess.QueryRow(getExperiencePointsQuery, userId)

	err = row.Scan(&points)

	dbaccess.LogDbResult(getExperiencePointsQuery, points, err)

	return
}

var getTerritoryHoursQuery = dbaccess.Query{Name: "GetTerritoryHoursQuery", SQL: `SELECT * FROM get_territory_hours($1::integer)`}

func (db *Database) GetTerritoryHoursCommand(userId int) (points int, err error) {
	row := dbaccess.QueryRow(getTerritoryHoursQuery, userId)

	err = row.Scan(&points)

	dbaccess.LogDbResult(getTerritoryHoursQuery, points, err)

	return
}

var changeTerritoryHoursQuery = dbaccess.Query{Name: "ChangeTerritoryHoursQuery", SQL: `SELECT change_territory_hours($1::integer, $2::integer)`}

func (db *Database) ChangeTerritoryHoursCommand(userId int, changeValue int) error {
	_, err := dbaccess.Exec(changeTerritoryHoursQuery, userId, changeValue)

	dbaccess.LogDbResult(changeTerritoryHoursQuery, nil, err)

	return err
}

var getTerritoryPointsQuery = dbaccess.Query{Name: "GetTerritoryPointsQuery", SQL: `SELECT * FROM get_territory_points($1::integer)`}

func (db *Database) GetTerritoryPointsCommand(userId int) (points int, err error) {
	row := dbaccess.QueryRow(getTerritoryPointsQuery, userId)

	err = row.Scan(&points)

	dbaccess.LogDbResult(getTerritoryPointsQuery, points, err)

	return
}

var changeTerritoryPointsQuery = dbaccess.Query{Name: "ChangeTerritoryPointsQuery", SQL: `SELECT change_territory_points($1::integer, $2::integer)`}

func (db *Database) ChangeTerritoryPointsCommand(userId int, changeValue int) error {
	_, err := dbaccess.Exec(changeTerritoryPointsQuery, userId, changeValue)

	dbaccess.LogDbResult(changeTerritoryPointsQuery, nil, err)

	return err
}

var addTerritoryPointHistoryQuery = dbaccess.Query{Name: "AddTerritoryPointHistoryQuery", SQL: `SELECT add_territory_point_history($1::integer, $2::integer, $3::text, $4::integer, $5::integer, $6::integer)`}

func (db *Database) AddTerritoryPointHistoryCommand(
	userId int,
	sourceUserId int,
	changeSource string,
	changeValue int,
	actualChangeValue int,
	finalValue int) error {

	_, err := dbaccess.Exec(
		addTerritoryPointHistoryQuery,
		userId,
		sourceUserId,
		changeSource,
		changeValue,
		actualChangeValue,
		finalValue)

	dbaccess.LogDbResult(addTerritoryPointHistoryQuery, nil, err)

	return err
}

var getTerritoryPointHistoryQuery = dbaccess.Query{Name: "GetTerritoryPointHistoryQuery", SQL: `SELECT * FROM get_territory_point_history($1::integer)`}

func (db *Database) GetTerritoryPointHistoryCommand(userId int) (
	history typepoints.TerritoryPointChangeHistories, err error) {

	rows, err := dbaccess.QueryRows(getTerritoryPointHistoryQuery, userId)

	if err != nil {
		return
	}

	for rows.Next() {
		entry := typepoints.TerritoryPointChangeHistory{}
		err = rows.Scan(
			&entry.ActualChangeValue,
			&entry.ChangeDate,
			&entry.ChangeSource,
			&entry.DesiredChangeValue,
			&entry.FinalValue,
			&entry.SourceLogin)

		if err != nil {
			_ = rows.Close()
			return
		}

		history = append(history, entry)
	}

	dbaccess.LogDbResult(getTerritoryPointHistoryQuery, history, err)

	_ = rows.Close()
	return
}

var getPointInfoQuery = dbaccess.Query{Name: "GetPointInfoQuery", SQL: `SELECT * FROM get_point_info($1::integer)`}

func (db *Database) GetPointInfoCommand(userId int) (info typepoints.PointInfo, err error) {
	row := dbaccess.QueryRow(getPointInfoQuery, userId)

	err = row.Scan(
		&info.TerritoryPoints,
		&info.FreePoints)

	dbaccess.LogDbResult(getPointInfoQuery, info, err)

	return
}

var getAllPointInfoQuery = dbaccess.Query{Name: "GetAllPointInfoQuery", SQL: `SELECT * FROM get_all_point_info()`}

func (db *Database) GetAllPointInfoCommand() (infos typepoints.PointInfoByLogins, err error) {
	rows, err := dbaccess.QueryRows(getAllPointInfoQuery)

	if err != nil {
		return
	}

	for rows.Next() {
		info := typepoints.PointInfoByLogin{}
		err = rows.Scan(
			&info.Login,
			&info.PointInfo.TerritoryPoints,
			&info.PointInfo.FreePoints)

		if err != nil {
			_ = rows.Close()
			return
		}

		infos = append(infos, info)
	}

	dbaccess.LogDbResult(getAllPointInfoQuery, infos, err)

	_ = rows.Close()
	return
}

var getFreePointsQuery = dbaccess.Query{Name: "GetFreePointsQuery", SQL: `SELECT * FROM get_free_points($1::integer)`}

func (db *Database) GetFreePointsCommand(userId int) (points int, err error) {
	row := dbaccess.QueryRow(getFreePointsQuery, userId)

	err = row.Scan(&points)

	dbaccess.LogDbResult(getFreePointsQuery, points, err)

	return
}

var changeFreePointsQuery = dbaccess.Query{Name: "ChangeFreePointsQuery", SQL: `SELECT change_free_points($1::integer, $2::integer)`}

func (db *Database) ChangeFreePointsCommand(userId int, changeValue int) error {
	_, err := dbaccess.Exec(changeFreePointsQuery, userId, changeValue)

	dbaccess.LogDbResult(changeFreePointsQuery, nil, err)

	return err
}

var addFreePointHistoryQuery = dbaccess.Query{Name: "AddFreePointHistoryQuery", SQL: `SELECT add_free_point_history($1::integer, $2::integer, $3::text, $4::integer, $5::integer, $6::integer, $7::integer)`}

func (db *Database) AddFreePointHistoryCommand(
	userId int,
	sourceUserId int,
	changeSource string,
	changeValue int,
	actualChangeValue int,
	finalValue int,
	wheelEffectId *int) error {

	_, err := dbaccess.Exec(
		addFreePointHistoryQuery,
		userId,
		sourceUserId,
		changeSource,
		changeValue,
		actualChangeValue,
		finalValue,
		wheelEffectId)

	dbaccess.LogDbResult(addFreePointHistoryQuery, nil, err)

	return err
}

var getFreePointHistoryQuery = dbaccess.Query{Name: "GetFreePointHistoryQuery", SQL: `SELECT * FROM get_free_point_history($1::integer)`}

func (db *Database) GetFreePointHistoryCommand(userId int) (history typepoints.FreePointChangeHistories, err error) {
	rows, err := dbaccess.QueryRows(getFreePointHistoryQuery, userId)

	if err != nil {
		return
	}

	for rows.Next() {
		entry := typepoints.FreePointChangeHistory{}
		err = rows.Scan(
			&entry.ActualChangeValue,
			&entry.ChangeDate,
			&entry.ChangeSource,
			&entry.DesiredChangeValue,
			&entry.FinalValue,
			&entry.SourceLogin,
			&entry.WheelEffectName)

		if err != nil {
			_ = rows.Close()
			return
		}

		history = append(history, entry)
	}

	dbaccess.LogDbResult(getFreePointHistoryQuery, history, err)

	_ = rows.Close()
	return
}

var createPointTypeQuery = dbaccess.Query{Name: "CreatePointTypeQuery", SQL: `SELECT * FROM create_point_type($1::integer, $2::text, $3::text, $4::integer, $5::boolean, $6::boolean, $7::integer, $8::integer)`}

func (db *Database) CreatePointTypeCommand(partyId int, name string, description string, startValue int, isPublic bool, isShared bool, minimum int, maximum int) (pointType typepoints.PointType, err error) {
	row := dbaccess.QueryRow(createPointTypeQuery, partyId, name, description, startValue, isPublic, isShared, minimum, maximum)

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

var changePointTypeQuery = dbaccess.Query{Name: "ChangePointTypeQuery", SQL: `SELECT change_point_type($1::integer, $2::integer, $3::text, $4::text, $5::boolean, $6::boolean, $7::integer, $8::integer)`}

func (db *Database) ChangePointTypeCommand(partyId int, pointTypeId int, name string, description string, isPublic bool, isShared bool, minimum int, maximum int) error {
	_, err := dbaccess.Exec(changePointTypeQuery, partyId, pointTypeId, name, description, isPublic, isShared, minimum, maximum)

	dbaccess.LogDbResult(changePointTypeQuery, nil, err)

	return err
}

var removePointTypeQuery = dbaccess.Query{Name: "RemovePointTypeQuery", SQL: `SELECT remove_point_type($1::integer, $2::integer)`}

func (db *Database) RemovePointTypeCommand(partyId int, pointTypeId int) error {
	_, err := dbaccess.Exec(removePointTypeQuery, partyId, pointTypeId)

	dbaccess.LogDbResult(removePointTypeQuery, nil, err)

	return err
}

var doesPointTypeExistQuery = dbaccess.Query{Name: "DoesPointTypeExistQuery", SQL: `SELECT does_point_type_exist($1::integer, $2::text)`}

func (db *Database) DoesPointTypeExistCommand(partyId int, name string) (doesExist bool, err error) {
	row := dbaccess.QueryRow(doesPointTypeExistQuery, partyId, name)

	err = row.Scan(&doesExist)

	dbaccess.LogDbResult(doesPointTypeExistQuery, doesExist, err)

	return
}

var getPointTypeQuery = dbaccess.Query{Name: "GetPointTypeQuery", SQL: `SELECT * FROM get_point_type($1::integer, $2::integer)`}

func (db *Database) GetPointTypeCommand(partyId int, pointTypeId int) (pointType typepoints.PointTypeInfo, err error) {
	row := dbaccess.QueryRow(getPointTypeQuery, partyId, pointTypeId)

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

func (db *Database) GetPointTypesCommand(partyId int) (pointTypes []typepoints.PointTypeInfo, err error) {
	rows, err := dbaccess.QueryRows(getPointTypesQuery, partyId)

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

	dbaccess.LogDbResult(getPointTypesQuery, pointTypes, err)

	_ = rows.Close()
	return
}

var createUserPointQuery = dbaccess.Query{Name: "CreateUserPointQuery", SQL: `SELECT * FROM create_user_point($1::integer, $2::integer, $3::integer, $4::integer)`}

func (db *Database) CreateUserPointCommand(userId int, partyId int, pointTypeId int, value int) (point typepoints.UserPoint, err error) {
	row := dbaccess.QueryRow(createUserPointQuery, userId, partyId, pointTypeId, value)

	err = row.Scan(&point.Id, &point.UserId, &point.PartyId, &point.PointTypeId, &point.Value)

	dbaccess.LogDbResult(createUserPointQuery, point, err)

	return
}

var createPartyPointQuery = dbaccess.Query{Name: "CreatePartyPointQuery", SQL: `SELECT * FROM create_party_point($1::integer, $2::integer, $3::integer)`}

func (db *Database) CreatePartyPointCommand(partyId int, pointTypeId int, value int) (point typepoints.PartyPoint, err error) {
	row := dbaccess.QueryRow(createPartyPointQuery, partyId, pointTypeId, value)

	err = row.Scan(&point.Id, &point.PartyId, &point.PointTypeId, &point.Value)

	dbaccess.LogDbResult(createPartyPointQuery, point, err)

	return
}

var getUserPointQuery = dbaccess.Query{Name: "GetUserPointQuery", SQL: `SELECT * FROM get_user_point($1::integer, $2::integer, $3::integer)`}

func (db *Database) GetUserPointCommand(userId int, partyId int, pointTypeId int) (point typepoints.UserPoint, err error) {
	row := dbaccess.QueryRow(getUserPointQuery, userId, partyId, pointTypeId)

	err = row.Scan(&point.Id, &point.UserId, &point.PartyId, &point.PointTypeId, &point.Value)

	dbaccess.LogDbResult(getUserPointQuery, point, err)

	return
}

var getPartyPointQuery = dbaccess.Query{Name: "GetPartyPointQuery", SQL: `SELECT * FROM get_party_point($1::integer, $2::integer)`}

func (db *Database) GetPartyPointCommand(partyId int, pointTypeId int) (point typepoints.PartyPoint, err error) {
	row := dbaccess.QueryRow(getPartyPointQuery, partyId, pointTypeId)

	err = row.Scan(&point.Id, &point.PartyId, &point.PointTypeId, &point.Value)

	dbaccess.LogDbResult(getPartyPointQuery, point, err)

	return
}

var getUserPointsQuery = dbaccess.Query{Name: "GetUserPointsQuery", SQL: `SELECT * FROM get_user_points($1::integer, $2::integer)`}

func (db *Database) GetUserPointsCommand(userId int, partyId int) (points []typepoints.UserPoint, err error) {
	rows, err := dbaccess.QueryRows(getUserPointsQuery, userId, partyId)

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

	dbaccess.LogDbResult(getUserPointsQuery, points, err)

	_ = rows.Close()
	return
}

var getPartyPointsQuery = dbaccess.Query{Name: "GetPartyPointsQuery", SQL: `SELECT * FROM get_party_points($1::integer)`}

func (db *Database) GetPartyPointsCommand(partyId int) (points []typepoints.PartyPoint, err error) {
	rows, err := dbaccess.QueryRows(getPartyPointsQuery, partyId)

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

	dbaccess.LogDbResult(getPartyPointsQuery, points, err)

	_ = rows.Close()
	return
}

var changeUserPointValueQuery = dbaccess.Query{Name: "ChangeUserPointValueQuery", SQL: `SELECT change_user_point_value($1::integer, $2::integer, $3::integer, $4::integer)`}

func (db *Database) ChangeUserPointValueCommand(userId int, partyId int, pointTypeId int, changeValue int) error {
	_, err := dbaccess.Exec(changeUserPointValueQuery, userId, partyId, pointTypeId, changeValue)

	dbaccess.LogDbResult(changeUserPointValueQuery, nil, err)

	return err
}

var changePartyPointValueQuery = dbaccess.Query{Name: "ChangePartyPointValueQuery", SQL: `SELECT change_party_point_value($1::integer, $2::integer, $3::integer)`}

func (db *Database) ChangePartyPointValueCommand(partyId int, pointTypeId int, changeValue int) error {
	_, err := dbaccess.Exec(changePartyPointValueQuery, partyId, pointTypeId, changeValue)

	dbaccess.LogDbResult(changePartyPointValueQuery, nil, err)

	return err
}

var createUserPointHistoryQuery = dbaccess.Query{Name: "CreateUserPointHistoryQuery", SQL: `SELECT * FROM create_user_point_history($1::integer, $2::integer, $3::integer, $4::integer, $5::integer, $6::integer, $7::integer, $8::integer)`}

func (db *Database) CreateUserPointHistoryCommand(userId int, partyId int, pointTypeId int, sourceUserId int, desiredChangeValue int, actualChangeValue int, finalValue int, sourceEventId int) (entry typepoints.UserPointHistoryEntry, err error) {
	row := dbaccess.QueryRow(createUserPointHistoryQuery, userId, partyId, pointTypeId, sourceUserId, desiredChangeValue, actualChangeValue, finalValue, sourceEventId)

	err = row.Scan(
		&entry.Id,
		&entry.UserId,
		&entry.PartyId,
		&entry.PointTypeId,
		&entry.SourceUserId,
		&entry.DesiredChangeValue,
		&entry.ActualChangeValue,
		&entry.FinalValue,
		&entry.SourceEventId,
		&entry.ChangedDate)

	dbaccess.LogDbResult(createUserPointHistoryQuery, entry, err)

	return
}

var createPartyPointHistoryQuery = dbaccess.Query{Name: "CreatePartyPointHistoryQuery", SQL: `SELECT * FROM create_party_point_history($1::integer, $2::integer, $3::integer, $4::integer, $5::integer, $6::integer, $7::integer)`}

func (db *Database) CreatePartyPointHistoryCommand(partyId int, pointTypeId int, sourceUserId int, desiredChangeValue int, actualChangeValue int, finalValue int, sourceEventId int) (entry typepoints.PartyPointHistoryEntry, err error) {
	row := dbaccess.QueryRow(createPartyPointHistoryQuery, partyId, pointTypeId, sourceUserId, desiredChangeValue, actualChangeValue, finalValue, sourceEventId)

	err = row.Scan(
		&entry.Id,
		&entry.PartyId,
		&entry.PointTypeId,
		&entry.SourceUserId,
		&entry.DesiredChangeValue,
		&entry.ActualChangeValue,
		&entry.FinalValue,
		&entry.SourceEventId,
		&entry.ChangedDate)

	dbaccess.LogDbResult(createPartyPointHistoryQuery, entry, err)

	return
}

var getUserPointHistoryQuery = dbaccess.Query{Name: "GetUserPointHistoryQuery", SQL: `SELECT * FROM get_user_point_history($1::integer, $2::integer)`}

func (db *Database) GetUserPointHistoryCommand(userId int, partyId int) (history []typepoints.UserPointHistoryEntry, err error) {
	rows, err := dbaccess.QueryRows(getUserPointHistoryQuery, userId, partyId)

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
			&entry.SourceUserId,
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

	dbaccess.LogDbResult(getUserPointHistoryQuery, history, err)

	_ = rows.Close()
	return
}

var getPartyPointHistoryQuery = dbaccess.Query{Name: "GetPartyPointHistoryQuery", SQL: `SELECT * FROM get_party_point_history($1::integer)`}

func (db *Database) GetPartyPointHistoryCommand(partyId int) (history []typepoints.PartyPointHistoryEntry, err error) {
	rows, err := dbaccess.QueryRows(getPartyPointHistoryQuery, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		entry := typepoints.PartyPointHistoryEntry{}
		err = rows.Scan(
			&entry.Id,
			&entry.PartyId,
			&entry.PointTypeId,
			&entry.SourceUserId,
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

	dbaccess.LogDbResult(getPartyPointHistoryQuery, history, err)

	_ = rows.Close()
	return
}
