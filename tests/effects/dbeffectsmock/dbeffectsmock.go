package dbeffectsmock

import (
	"FGG-Service/src/changes/typechanges"
	"FGG-Service/src/effects/typeeffects"
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type DatabaseMock struct {
	mock.Mock
}

func (m *DatabaseMock) CreateEffectCommand(_ context.Context, partyId int, name string, description string, useCount *int, duration *time.Duration, change typechanges.Change, modifiers []typeeffects.PointModifier) (effect typeeffects.EffectWithChange, err error) {
	args := m.Called(partyId, name, description, useCount, duration, change, modifiers)
	effect = args.Get(0).(typeeffects.EffectWithChange)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetEffectCommand(_ context.Context, partyId int, effectId int) (effect typeeffects.EffectWithChange, err error) {
	args := m.Called(partyId, effectId)
	effect = args.Get(0).(typeeffects.EffectWithChange)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetActualEffectsCommand(_ context.Context, partyId int) (effects []typeeffects.Effect, err error) {
	args := m.Called(partyId)
	effects = args.Get(0).([]typeeffects.Effect)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetRemovedEffectsCommand(_ context.Context, partyId int) (effects []typeeffects.Effect, err error) {
	args := m.Called(partyId)
	effects = args.Get(0).([]typeeffects.Effect)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) RemoveEffectCommand(_ context.Context, partyId int, effectId int) error {
	args := m.Called(partyId, effectId)
	return args.Error(0)
}

func (m *DatabaseMock) CreateUserEffectCommand(_ context.Context, userId int, partyId int, effectId int, actorUserId int, sourceEventId int) (userEffect typeeffects.UserEffect, err error) {
	args := m.Called(userId, partyId, effectId, actorUserId, sourceEventId)
	userEffect = args.Get(0).(typeeffects.UserEffect)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetUserEffectCommand(_ context.Context, userId int, partyId int, effectId int) (userEffect typeeffects.UserEffectDetail, err error) {
	args := m.Called(userId, partyId, effectId)
	userEffect = args.Get(0).(typeeffects.UserEffectDetail)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetUserEffectsCommand(_ context.Context, userId int, partyId int) (userEffects []typeeffects.UserEffectDetail, err error) {
	args := m.Called(userId, partyId)
	userEffects = args.Get(0).([]typeeffects.UserEffectDetail)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) ChangeUserEffectUsesLeftCommand(_ context.Context, userId int, partyId int, effectId int, usesLeft *int, actorUserId int, sourceEventId *int) (historyEventId int, err error) {
	args := m.Called(userId, partyId, effectId, usesLeft, actorUserId, sourceEventId)
	historyEventId = args.Int(0)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) DeleteUserEffectCommand(_ context.Context, userId int, partyId int, effectId int, actorUserId int, sourceEventId *int) error {
	args := m.Called(userId, partyId, effectId, actorUserId, sourceEventId)
	return args.Error(0)
}

func (m *DatabaseMock) GetEndedUserEffectsCommand(_ context.Context) (ended []typeeffects.EndedUserEffect, err error) {
	args := m.Called()
	ended = args.Get(0).([]typeeffects.EndedUserEffect)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetEffectHistoryCommand(_ context.Context, userId int, partyId int) (history []typeeffects.EffectHistory, err error) {
	args := m.Called(userId, partyId)
	history = args.Get(0).([]typeeffects.EffectHistory)
	err = args.Error(1)
	return
}

func (m *DatabaseMock) GetUserEffectPointModifiersJsonbCommand(_ context.Context, partyId int, userEffectId int) (modifiers []typeeffects.PointModifier, err error) {
	args := m.Called(partyId, userEffectId)
	modifiers = args.Get(0).([]typeeffects.PointModifier)
	err = args.Error(1)
	return
}
