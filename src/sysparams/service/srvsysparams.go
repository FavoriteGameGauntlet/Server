package srvsysparams

import (
	"FGG-Service/src/common"
	"FGG-Service/src/sysparams/database"
	"FGG-Service/src/sysparams/types"
	"database/sql"
	"errors"
	"strconv"
	"strings"
)

type IService interface {
	GetAll() ([]typesysparams.SystemParameter, error)
	GetAllApp() ([]typesysparams.SystemParameter, error)
	GetParameter(name string) (typesysparams.SystemParameter, error)
	GetAppParameter(name string) (typesysparams.SystemParameter, error)
	GetString(name string) (string, error)
	GetInt(name string) (int, error)
	GetBool(name string) (bool, error)
	GetIntSlice(name string) ([]int, error)
	ChangeValue(name string, value string) error
}

// defaultPartyId is a stopgap until real party-context resolution exists (see project plan).
const defaultPartyId = 1

type Service struct {
	Database dbsysparams.IDatabase
}

func NewService() *Service {
	db := new(dbsysparams.Database)

	return &Service{
		Database: db,
	}
}

func (s *Service) GetAll() (parameters []typesysparams.SystemParameter, err error) {
	parameters, err = s.Database.GetAllSystemParametersCommand(defaultPartyId)

	return
}

// GetAllApp used to filter to parameters flagged ShouldShowToApp; that flag no longer exists in the
// schema, so this currently returns every parameter. TODO: reintroduce app-visibility filtering once
// the schema has a way to express it.
func (s *Service) GetAllApp() (parameters []typesysparams.SystemParameter, err error) {
	return s.Database.GetAllSystemParametersCommand(defaultPartyId)
}

func (s *Service) GetParameter(name string) (parameter typesysparams.SystemParameter, err error) {
	parameter, err = s.Database.GetSystemParameterCommand(defaultPartyId, name)

	if errors.Is(err, sql.ErrNoRows) {
		err = common.NewSystemParameterNotFoundError(name)
		return
	}

	return
}

// GetAppParameter used to also require ShouldShowToApp; that flag no longer exists in the schema, so
// this is currently equivalent to GetParameter. TODO: reintroduce app-visibility filtering once the
// schema has a way to express it.
func (s *Service) GetAppParameter(name string) (parameter typesysparams.SystemParameter, err error) {
	return s.GetParameter(name)
}

func (s *Service) GetString(name string) (value string, err error) {
	parameter, err := s.GetParameter(name)

	if err != nil {
		return
	}

	value = parameter.Value

	return
}

func (s *Service) GetInt(name string) (value int, err error) {
	stringValue, err := s.GetString(name)

	if err != nil {
		return
	}

	value, err = strconv.Atoi(stringValue)

	return
}

func (s *Service) GetBool(name string) (value bool, err error) {
	str, err := s.GetString(name)

	if err != nil {
		return
	}

	if str == "" {
		return
	}

	value, err = strconv.ParseBool(str)

	return
}

func (s *Service) GetIntSlice(name string) (values []int, err error) {
	str, err := s.GetString(name)

	if err != nil {
		return
	}

	for _, part := range strings.Split(strings.Trim(str, "[]"), ",") {
		var v int
		v, err = strconv.Atoi(strings.TrimSpace(part))

		if err != nil {
			return
		}

		values = append(values, v)
	}

	return
}

func (s *Service) ChangeValue(name string, value string) (err error) {
	parameter, err := s.GetParameter(name)

	if err != nil {
		return
	}

	err = s.Database.ChangeSystemParameterValueCommand(defaultPartyId, parameter.Id, value)

	return
}
