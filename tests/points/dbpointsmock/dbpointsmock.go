package dbpointsmock

import (
	"FGG-Service/src/points/typepoints"
	"context"

	"github.com/stretchr/testify/mock"
)

type DatabaseMock struct {
	mock.Mock
}

func (m *DatabaseMock) CreatePointTypeCommand(_ context.Context, partyId int, name string, description string, startValue int, isPublic bool, isShared bool, minimum *int, maximum *int) (pointType typepoints.PointType, err error) {
	args := m.Called(partyId, name, description, startValue, isPublic, isShared, minimum, maximum)
	pointType = args.Get(0).(typepoints.PointType)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) ChangePointTypeCommand(_ context.Context, partyId int, pointTypeId int, name string, description string, startValue int, isPublic bool, isShared bool, minimum *int, maximum *int) error {
	args := m.Called(partyId, pointTypeId, name, description, startValue, isPublic, isShared, minimum, maximum)
	return args.Error(0)
}

func (m *DatabaseMock) RemovePointTypeCommand(_ context.Context, partyId int, pointTypeId int) error {
	args := m.Called(partyId, pointTypeId)
	return args.Error(0)
}

func (m *DatabaseMock) GetPointTypeByNameCommand(_ context.Context, partyId int, name string) (pointType typepoints.PointTypeInfo, err error) {
	args := m.Called(partyId, name)
	pointType = args.Get(0).(typepoints.PointTypeInfo)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetPointTypeCommand(_ context.Context, partyId int, pointTypeId int) (pointType typepoints.PointTypeInfo, err error) {
	args := m.Called(partyId, pointTypeId)
	pointType = args.Get(0).(typepoints.PointTypeInfo)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetPointTypesCommand(_ context.Context, partyId int) (pointTypes []typepoints.PointTypeInfo, err error) {
	args := m.Called(partyId)
	pointTypes = args.Get(0).([]typepoints.PointTypeInfo)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) CreateUserPointCommand(_ context.Context, userId int, partyId int, pointTypeId int, value int) (point typepoints.UserPoint, err error) {
	args := m.Called(userId, partyId, pointTypeId, value)
	point = args.Get(0).(typepoints.UserPoint)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) CreatePartyPointCommand(_ context.Context, partyId int, pointTypeId int, value int) (point typepoints.PartyPoint, err error) {
	args := m.Called(partyId, pointTypeId, value)
	point = args.Get(0).(typepoints.PartyPoint)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetUserPointCommand(_ context.Context, userId int, partyId int, pointTypeId int) (point typepoints.UserPoint, err error) {
	args := m.Called(userId, partyId, pointTypeId)
	point = args.Get(0).(typepoints.UserPoint)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetPartyPointCommand(_ context.Context, partyId int, pointTypeId int) (point typepoints.PartyPoint, err error) {
	args := m.Called(partyId, pointTypeId)
	point = args.Get(0).(typepoints.PartyPoint)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetUserPointsCommand(_ context.Context, userId int, partyId int) (points []typepoints.UserPoint, err error) {
	args := m.Called(userId, partyId)
	points = args.Get(0).([]typepoints.UserPoint)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetPartyPointsCommand(_ context.Context, partyId int) (points []typepoints.PartyPoint, err error) {
	args := m.Called(partyId)
	points = args.Get(0).([]typepoints.PartyPoint)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) ChangeUserPointValueCommand(_ context.Context, userId int, partyId int, pointTypeId int, changeValue int) error {
	args := m.Called(userId, partyId, pointTypeId, changeValue)
	return args.Error(0)
}

func (m *DatabaseMock) ChangePartyPointValueCommand(_ context.Context, partyId int, pointTypeId int, changeValue int) error {
	args := m.Called(partyId, pointTypeId, changeValue)
	return args.Error(0)
}

func (m *DatabaseMock) CreateUserPointHistoryCommand(_ context.Context, userId int, partyId int, pointTypeId int, actorUserId int, desiredChangeValue int, actualChangeValue int, finalValue int, sourceEventId int) (entry typepoints.UserPointHistoryEntry, err error) {
	args := m.Called(userId, partyId, pointTypeId, actorUserId, desiredChangeValue, actualChangeValue, finalValue, sourceEventId)
	entry = args.Get(0).(typepoints.UserPointHistoryEntry)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) CreatePartyPointHistoryCommand(_ context.Context, partyId int, pointTypeId int, actorUserId int, desiredChangeValue int, actualChangeValue int, finalValue int, sourceEventId int) (entry typepoints.PartyPointHistoryEntry, err error) {
	args := m.Called(partyId, pointTypeId, actorUserId, desiredChangeValue, actualChangeValue, finalValue, sourceEventId)
	entry = args.Get(0).(typepoints.PartyPointHistoryEntry)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetUserPointHistoryCommand(_ context.Context, userId int, partyId int) (history []typepoints.UserPointHistoryEntry, err error) {
	args := m.Called(userId, partyId)
	history = args.Get(0).([]typepoints.UserPointHistoryEntry)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetPartyPointHistoryCommand(_ context.Context, partyId int) (history []typepoints.PartyPointHistoryEntry, err error) {
	args := m.Called(partyId)
	history = args.Get(0).([]typepoints.PartyPointHistoryEntry)
	err = args.Error(1)
	return
}
