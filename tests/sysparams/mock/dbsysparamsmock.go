package dbsysparamsmock

import (
	"FGG-Service/src/sysparams/types"

	"github.com/stretchr/testify/mock"
)

type DatabaseMock struct {
	mock.Mock
}

func (m *DatabaseMock) GetAllSystemParametersCommand(partyId int) (parameters []typesysparams.SystemParameter, err error) {
	args := m.Called(partyId)
	parameters = args.Get(0).([]typesysparams.SystemParameter)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetSystemParameterCommand(partyId int, code string) (parameter typesysparams.SystemParameter, err error) {
	args := m.Called(partyId, code)
	parameter = args.Get(0).(typesysparams.SystemParameter)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) ChangeSystemParameterValueCommand(partyId int, systemParameterId int, value string) (bool, error) {
	args := m.Called(partyId, systemParameterId, value)
	return args.Bool(0), args.Error(1)
}

func (m *DatabaseMock) GetDefaultSystemParametersCommand() (parameters []typesysparams.DefaultSystemParameter, err error) {
	args := m.Called()
	parameters = args.Get(0).([]typesysparams.DefaultSystemParameter)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) DeleteSystemParameterOverrideCommand(partyId int, systemParameterId int) error {
	args := m.Called(partyId, systemParameterId)
	return args.Error(0)
}
