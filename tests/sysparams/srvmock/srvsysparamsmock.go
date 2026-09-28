package srvsysparamsmock

import (
	"FGG-Service/src/sysparams/types"

	"github.com/stretchr/testify/mock"
)

type ServiceMock struct {
	mock.Mock
}

func (m *ServiceMock) GetAll() ([]typesysparams.SystemParameter, error) {
	args := m.Called()
	return args.Get(0).([]typesysparams.SystemParameter), args.Error(1)
}

func (m *ServiceMock) GetParameter(name string) (typesysparams.SystemParameter, error) {
	args := m.Called(name)
	return args.Get(0).(typesysparams.SystemParameter), args.Error(1)
}

func (m *ServiceMock) GetString(name string) (string, error) {
	args := m.Called(name)
	return args.String(0), args.Error(1)
}

func (m *ServiceMock) GetInt(name string) (int, error) {
	args := m.Called(name)
	return args.Int(0), args.Error(1)
}

func (m *ServiceMock) GetBool(name string) (bool, error) {
	args := m.Called(name)
	return args.Bool(0), args.Error(1)
}

func (m *ServiceMock) GetIntSlice(name string) ([]int, error) {
	args := m.Called(name)
	return args.Get(0).([]int), args.Error(1)
}

func (m *ServiceMock) ChangeValue(name string, value string) (bool, error) {
	args := m.Called(name, value)
	return args.Bool(0), args.Error(1)
}

func (m *ServiceMock) ResetParameter(name string) error {
	args := m.Called(name)
	return args.Error(0)
}