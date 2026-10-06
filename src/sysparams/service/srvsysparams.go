package srvsysparams

import (
	"FGG-Service/src/common"
	"FGG-Service/src/sysparams/database"
	"FGG-Service/src/sysparams/types"
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
)

type IService interface {
	GetAll(ctx context.Context, partyId int) ([]typesysparams.SystemParameter, error)
	GetParameter(ctx context.Context, partyId int, name string) (typesysparams.SystemParameter, error)
	ResetParameter(ctx context.Context, partyId int, name string) error
	GetString(ctx context.Context, partyId int, name string) (string, error)
	GetInt(ctx context.Context, partyId int, name string) (int, error)
	GetBool(ctx context.Context, partyId int, name string) (bool, error)
	GetIntSlice(ctx context.Context, partyId int, name string) ([]int, error)
	ChangeValue(ctx context.Context, partyId int, name string, value string) (bool, error)
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

func (s *Service) GetAll(ctx context.Context, partyId int) (parameters []typesysparams.SystemParameter, err error) {
	parameters, err = s.Database.GetAllSystemParametersCommand(ctx, partyId)

	return
}

func (s *Service) GetParameter(ctx context.Context, partyId int, name string) (parameter typesysparams.SystemParameter, err error) {
	parameter, err = s.Database.GetSystemParameterCommand(ctx, partyId, name)

	if errors.Is(err, sql.ErrNoRows) {
		err = common.NewSystemParameterNotFoundError(name)
		return
	}

	return
}

func (s *Service) GetString(ctx context.Context, partyId int, name string) (value string, err error) {
	parameter, err := s.GetParameter(ctx, partyId, name)

	if err != nil {
		return
	}

	value = parameter.Value

	return
}

func (s *Service) GetInt(ctx context.Context, partyId int, name string) (value int, err error) {
	stringValue, err := s.GetString(ctx, partyId, name)

	if err != nil {
		return
	}

	value, err = strconv.Atoi(stringValue)

	return
}

func (s *Service) GetBool(ctx context.Context, partyId int, name string) (value bool, err error) {
	str, err := s.GetString(ctx, partyId, name)

	if err != nil {
		return
	}

	if str == "" {
		return
	}

	value, err = strconv.ParseBool(str)

	return
}

func (s *Service) GetIntSlice(ctx context.Context, partyId int, name string) (values []int, err error) {
	str, err := s.GetString(ctx, partyId, name)

	if err != nil {
		return
	}

	for part := range strings.SplitSeq(strings.Trim(str, "[]"), ",") {
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
func (s *Service) ChangeValue(ctx context.Context, partyId int, name string, value string) (created bool, err error) {
	parameter, err := s.GetParameter(ctx, partyId, name)

	if err != nil {
		return
	}

	return s.Database.ChangeSystemParameterValueCommand(ctx, partyId, parameter.Id, value)
}

// ResetParameter drops the party's override of a parameter so it falls back to its default value.
func (s *Service) ResetParameter(ctx context.Context, partyId int, name string) (err error) {
	parameter, err := s.GetParameter(ctx, partyId, name)

	if err != nil {
		return
	}

	return s.Database.DeleteSystemParameterOverrideCommand(ctx, partyId, parameter.Id)
}
