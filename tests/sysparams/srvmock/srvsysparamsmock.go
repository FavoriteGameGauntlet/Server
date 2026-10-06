package srvsysparamsmock

import (
	"FGG-Service/src/sysparams/types"
	"context"

	"github.com/stretchr/testify/mock"
)

type ServiceMock struct {
	mock.Mock
}

func (m *ServiceMock) GetAll(_ context.Context, partyId int) ([]typesysparams.SystemParameter, error) {
	args := m.Called(partyId)
	return args.Get(0).([]typesysparams.SystemParameter), args.Error(1)
}

func (m *ServiceMock) GetParameter(_ context.Context, partyId int, name string) (typesysparams.SystemParameter, error) {
	args := m.Called(partyId, name)
	return args.Get(0).(typesysparams.SystemParameter), args.Error(1)
}

func (m *ServiceMock) GetString(_ context.Context, partyId int, name string) (string, error) {
	args := m.Called(partyId, name)
	return args.String(0), args.Error(1)
}

func (m *ServiceMock) GetInt(_ context.Context, partyId int, name string) (int, error) {
	args := m.Called(partyId, name)
	return args.Int(0), args.Error(1)
}

func (m *ServiceMock) GetBool(_ context.Context, partyId int, name string) (bool, error) {
	args := m.Called(partyId, name)
	return args.Bool(0), args.Error(1)
}

func (m *ServiceMock) GetIntSlice(_ context.Context, partyId int, name string) ([]int, error) {
	args := m.Called(partyId, name)
	return args.Get(0).([]int), args.Error(1)
}

func (m *ServiceMock) ChangeValue(_ context.Context, partyId int, name string, value string) (bool, error) {
	args := m.Called(partyId, name, value)
	return args.Bool(0), args.Error(1)
}

func (m *ServiceMock) ResetParameter(_ context.Context, partyId int, name string) error {
	args := m.Called(partyId, name)
	return args.Error(0)
}
