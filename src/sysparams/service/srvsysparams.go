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
	GetAll(partyId int) ([]typesysparams.SystemParameter, error)
	GetParameter(partyId int, name string) (typesysparams.SystemParameter, error)
	ResetParameter(partyId int, name string) error
	GetString(partyId int, name string) (string, error)
	GetInt(partyId int, name string) (int, error)
	GetBool(partyId int, name string) (bool, error)
	GetIntSlice(partyId int, name string) ([]int, error)
	ChangeValue(partyId int, name string, value string) (bool, error)
}

type Service struct {
	Database dbsysparams.IDatabase
}

func NewService() *Service {
	db := new(dbsysparams.Database)

	return &Service{
		Database: db,
	}
}

func (s *Service) GetAll(partyId int) (parameters []typesysparams.SystemParameter, err error) {
	parameters, err = s.Database.GetAllSystemParametersCommand(partyId)

	return
}

func (s *Service) GetParameter(partyId int, name string) (parameter typesysparams.SystemParameter, err error) {
	parameter, err = s.Database.GetSystemParameterCommand(partyId, name)

	if errors.Is(err, sql.ErrNoRows) {
		err = common.NewSystemParameterNotFoundError(name)
		return
	}

	return
}

func (s *Service) GetString(partyId int, name string) (value string, err error) {
	parameter, err := s.GetParameter(partyId, name)

	if err != nil {
		return
	}

	value = parameter.Value

	return
}

func (s *Service) GetInt(partyId int, name string) (value int, err error) {
	stringValue, err := s.GetString(partyId, name)

	if err != nil {
		return
	}

	value, err = strconv.Atoi(stringValue)

	return
}

func (s *Service) GetBool(partyId int, name string) (value bool, err error) {
	str, err := s.GetString(partyId, name)

	if err != nil {
		return
	}

	if str == "" {
		return
	}

	value, err = strconv.ParseBool(str)

	return
}

func (s *Service) GetIntSlice(partyId int, name string) (values []int, err error) {
	str, err := s.GetString(partyId, name)

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
func (s *Service) ChangeValue(partyId int, name string, value string) (created bool, err error) {
	parameter, err := s.GetParameter(partyId, name)

	if err != nil {
		return
	}

	return s.Database.ChangeSystemParameterValueCommand(partyId, parameter.Id, value)
}

// ResetParameter drops the party's override of a parameter so it falls back to its default value.
func (s *Service) ResetParameter(partyId int, name string) (err error) {
	parameter, err := s.GetParameter(partyId, name)

	if err != nil {
		return
	}

	return s.Database.DeleteSystemParameterOverrideCommand(partyId, parameter.Id)
}