package dbeffectsmock

import (
	"FGG-Service/src/changes/types"
	"FGG-Service/src/effects/types"
	"time"

	"github.com/stretchr/testify/mock"
)

type DatabaseMock struct {
	mock.Mock
}

func (m *DatabaseMock) CreateEffectCommand(partyId int, name string, description string, useCount int, duration *time.Duration, change typechanges.Change) (effect typeeffects.EffectWithChange, err error) {
	args := m.Called(partyId, name, description, useCount, duration, change)
	effect = args.Get(0).(typeeffects.EffectWithChange)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetEffectCommand(partyId int, effectId int) (effect typeeffects.EffectWithChange, err error) {
	args := m.Called(partyId, effectId)
	effect = args.Get(0).(typeeffects.EffectWithChange)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetActualEffectsCommand(partyId int) (effects []typeeffects.Effect, err error) {
	args := m.Called(partyId)
	effects = args.Get(0).([]typeeffects.Effect)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetRemovedEffectsCommand(partyId int) (effects []typeeffects.Effect, err error) {
	args := m.Called(partyId)
	effects = args.Get(0).([]typeeffects.Effect)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) RemoveEffectCommand(partyId int, effectId int) error {
	args := m.Called(partyId, effectId)
	return args.Error(0)
}

func (m *DatabaseMock) CreateUserEffectCommand(userId int, partyId int, effectId int, sourceEventId int) (userEffect typeeffects.UserEffect, err error) {
	args := m.Called(userId, partyId, effectId, sourceEventId)
	userEffect = args.Get(0).(typeeffects.UserEffect)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetUserEffectCommand(userId int, partyId int, effectId int) (userEffect typeeffects.UserEffectDetail, err error) {
	args := m.Called(userId, partyId, effectId)
	userEffect = args.Get(0).(typeeffects.UserEffectDetail)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetUserEffectsCommand(userId int, partyId int) (userEffects []typeeffects.UserEffectDetail, err error) {
	args := m.Called(userId, partyId)
	userEffects = args.Get(0).([]typeeffects.UserEffectDetail)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) ChangeUserEffectUsesLeftCommand(userId int, partyId int, effectId int, usesLeft int, sourceEventId int) error {
	args := m.Called(userId, partyId, effectId, usesLeft, sourceEventId)
	return args.Error(0)
}

func (m *DatabaseMock) DeleteUserEffectCommand(userId int, partyId int, effectId int, sourceEventId int) error {
	args := m.Called(userId, partyId, effectId, sourceEventId)
	return args.Error(0)
}

func (m *DatabaseMock) DeleteEndedUserEffectsCommand() (deleted []typeeffects.EndedUserEffect, err error) {
	args := m.Called()
	deleted = args.Get(0).([]typeeffects.EndedUserEffect)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetEffectHistoryCommand(userId int, partyId int) (history []typeeffects.EffectHistory, err error) {
	args := m.Called(userId, partyId)
	history = args.Get(0).([]typeeffects.EffectHistory)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetUserEffectPointModifiersJsonbCommand(partyId int, userEffectId int) (modifiers []typeeffects.PointModifier, err error) {
	args := m.Called(partyId, userEffectId)
	modifiers = args.Get(0).([]typeeffects.PointModifier)
	err = args.Error(1)
	return
}
