package dbsysparams

import (
	"FGG-Service/src/dbaccess"
	"FGG-Service/src/sysparams/types"
	"context"
)

type IDatabase interface {
	GetAllSystemParametersCommand(ctx context.Context, partyId int) (parameters []typesysparams.SystemParameter, err error)
	GetSystemParameterCommand(ctx context.Context, partyId int, code string) (parameter typesysparams.SystemParameter, err error)
	ChangeSystemParameterValueCommand(ctx context.Context, partyId int, systemParameterId int, value string) (created bool, err error)
	GetDefaultSystemParametersCommand(ctx context.Context) (parameters []typesysparams.DefaultSystemParameter, err error)
	DeleteSystemParameterOverrideCommand(ctx context.Context, partyId int, systemParameterId int) error
}

type Database struct {
}

var getAllSystemParametersQuery = dbaccess.Query{Name: "GetAllSystemParametersQuery", SQL: `SELECT * FROM get_system_parameters($1::integer)`}

func (db *Database) GetAllSystemParametersCommand(ctx context.Context, partyId int) (parameters []typesysparams.SystemParameter, err error) {
	rows, err := dbaccess.QueryRows(ctx, getAllSystemParametersQuery, partyId)

	if err != nil {
		return
	}

	for rows.Next() {
		parameter := typesysparams.SystemParameter{}
		err = rows.Scan(&parameter.Id, &parameter.Code, &parameter.Name, &parameter.Description, &parameter.Value, &parameter.IsDefault)

		if err != nil {
			_ = rows.Close()
			return
		}

		parameters = append(parameters, parameter)
	}

	err = rows.Err()

	dbaccess.LogDbResult(getAllSystemParametersQuery, parameters, err)

	_ = rows.Close()
	return
}

var getSystemParameterQuery = dbaccess.Query{Name: "GetSystemParameterQuery", SQL: `SELECT * FROM get_system_parameter($1::integer, $2::text)`, IsSilent: true}

func (db *Database) GetSystemParameterCommand(ctx context.Context, partyId int, code string) (parameter typesysparams.SystemParameter, err error) {
	row := dbaccess.QueryRow(ctx, getSystemParameterQuery, partyId, code)

	err = row.Scan(&parameter.Id, &parameter.Code, &parameter.Name, &parameter.Description, &parameter.Value, &parameter.IsDefault)

	dbaccess.LogDbResult(getSystemParameterQuery, parameter, err)

	return
}

var changeSystemParameterValueQuery = dbaccess.Query{Name: "ChangeSystemParameterValueQuery", SQL: `SELECT change_system_parameter_value($1::integer, $2::integer, $3::text)`}

// ChangeSystemParameterValueCommand upserts the party's override of a parameter and reports whether
// the override was created rather than changed.
func (db *Database) ChangeSystemParameterValueCommand(ctx context.Context, partyId int, systemParameterId int, value string) (created bool, err error) {
	row := dbaccess.QueryRow(ctx, changeSystemParameterValueQuery, partyId, systemParameterId, value)

	err = row.Scan(&created)

	dbaccess.LogDbResult(changeSystemParameterValueQuery, created, err)

	return
}

var getDefaultSystemParametersQuery = dbaccess.Query{Name: "GetDefaultSystemParametersQuery", SQL: `SELECT * FROM get_default_system_parameters()`}

func (db *Database) GetDefaultSystemParametersCommand(ctx context.Context) (parameters []typesysparams.DefaultSystemParameter, err error) {
	rows, err := dbaccess.QueryRows(ctx, getDefaultSystemParametersQuery)

	if err != nil {
		return
	}

	for rows.Next() {
		parameter := typesysparams.DefaultSystemParameter{}
		err = rows.Scan(&parameter.Id, &parameter.Code, &parameter.DefaultValue, &parameter.Name, &parameter.Description)

		if err != nil {
			_ = rows.Close()
			return
		}

		parameters = append(parameters, parameter)
	}

	err = rows.Err()

	dbaccess.LogDbResult(getDefaultSystemParametersQuery, parameters, err)

	_ = rows.Close()
	return
}

var deleteSystemParameterOverrideQuery = dbaccess.Query{Name: "DeleteSystemParameterOverrideQuery", SQL: `SELECT delete_system_parameter_override($1::integer, $2::integer)`}

func (db *Database) DeleteSystemParameterOverrideCommand(ctx context.Context, partyId int, systemParameterId int) error {
	_, err := dbaccess.Exec(ctx, deleteSystemParameterOverrideQuery, partyId, systemParameterId)

	dbaccess.LogDbResult(deleteSystemParameterOverrideQuery, nil, err)

	return err
}
