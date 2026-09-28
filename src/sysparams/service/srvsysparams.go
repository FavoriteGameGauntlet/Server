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
	GetParameter(name string) (typesysparams.SystemParameter, error)
	ResetParameter(name string) error
	GetString(name string) (string, error)
	GetInt(name string) (int, error)
	GetBool(name string) (bool, error)
	GetIntSlice(name string) ([]int, error)
	ChangeValue(name string, value string) (bool, error)
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

func (s *Service) GetParameter(name string) (parameter typesysparams.SystemParameter, err error) {
	parameter, err = s.Database.GetSystemParameterCommand(defaultPartyId, name)

	if errors.Is(err, sql.ErrNoRows) {
		err = common.NewSystemParameterNotFoundError(name)
		return
	}

	return
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

// ChangeValue sets the party's override of a parameter; created tells whether the override is new.
func (s *Service) ChangeValue(name string, value string) (created bool, err error) {
	parameter, err := s.GetParameter(name)

	if err != nil {
		return
	}

	return s.Database.ChangeSystemParameterValueCommand(defaultPartyId, parameter.Id, value)
}

// ResetParameter drops the party's override of a parameter so it falls back to its default value.
func (s *Service) ResetParameter(name string) (err error) {
	parameter, err := s.GetParameter(name)

	if err != nil {
		return
	}

	return s.Database.DeleteSystemParameterOverrideCommand(defaultPartyId, parameter.Id)
}