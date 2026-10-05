package srvsysparamsmock

import (
	"FGG-Service/src/sysparams/types"

	"github.com/stretchr/testify/mock"
)

type ServiceMock struct {
	mock.Mock
}

func (m *ServiceMock) GetAll(partyId int) ([]typesysparams.SystemParameter, error) {
	args := m.Called(partyId)
	return args.Get(0).([]typesysparams.SystemParameter), args.Error(1)
}

func (m *ServiceMock) GetParameter(partyId int, name string) (typesysparams.SystemParameter, error) {
	args := m.Called(partyId, name)
	return args.Get(0).(typesysparams.SystemParameter), args.Error(1)
}

func (m *ServiceMock) GetString(partyId int, name string) (string, error) {
	args := m.Called(partyId, name)
	return args.String(0), args.Error(1)
}

func (m *ServiceMock) GetInt(partyId int, name string) (int, error) {
	args := m.Called(partyId, name)
	return args.Int(0), args.Error(1)
}

func (m *ServiceMock) GetBool(partyId int, name string) (bool, error) {
	args := m.Called(partyId, name)
	return args.Bool(0), args.Error(1)
}

func (m *ServiceMock) GetIntSlice(partyId int, name string) ([]int, error) {
	args := m.Called(partyId, name)
	return args.Get(0).([]int), args.Error(1)
}

func (m *ServiceMock) ChangeValue(partyId int, name string, value string) (bool, error) {
	args := m.Called(partyId, name, value)
	return args.Bool(0), args.Error(1)
}

func (m *ServiceMock) ResetParameter(partyId int, name string) error {
	args := m.Called(partyId, name)
	return args.Error(0)
}
