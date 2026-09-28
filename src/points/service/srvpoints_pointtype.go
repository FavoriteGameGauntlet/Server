package srvpoints

import (
	"FGG-Service/src/common"
	"FGG-Service/src/points/type"
	"FGG-Service/src/validator"
	"math"
)

// GetPointValueByTypeName reads a user's current value for the named point type.
func (s *Service) GetPointValueByTypeName(userId int, partyId int, pointTypeName string) (value int, err error) {
	pointType, err := s.GetPointTypeByName(partyId, pointTypeName)

	if err != nil {
		return
	}

	return s.pointValue(userId, partyId, pointType)
}

// clampedPointChange computes the value actually applied once changeValue is clamped to the point
// type's Minimum/Maximum, and the value the point ends up at. A nil bound falls back to the range of
// the INTEGER column the value is stored in, so an unbounded point clamps instead of overflowing.
func (s *Service) clampedPointChange(userId int, partyId int, pointType typepoints.PointTypeInfo, changeValue int) (result typepoints.PointChangeResult, err error) {
	currentValue, err := s.pointValue(userId, partyId, pointType)

	if err != nil {
		return
	}

	result = typepoints.PointChangeResult{
		DesiredChangeValue: changeValue,
		ActualChangeValue:  changeValue,
		FinalValue:         currentValue + changeValue,
	}

	minimum, maximum := math.MinInt32, math.MaxInt32

	if pointType.Minimum != nil {
		minimum = *pointType.Minimum
	}

	if pointType.Maximum != nil {
		maximum = *pointType.Maximum
	}

	if result.FinalValue < minimum {
		result.FinalValue = minimum
		result.ActualChangeValue = result.FinalValue - currentValue
	}

	if result.FinalValue > maximum {
		result.FinalValue = maximum
		result.ActualChangeValue = result.FinalValue - currentValue
	}

	return
}

// changePointValue applies a clamped change and records it. A shared point type changes the party
// pool rather than the user's own value. sourceEventId is the history event that caused the change;
// when it is nil the change is applied without being recorded, for changes with no event behind them.
func (s *Service) changePointValue(
	userId int,
	partyId int,
	pointType typepoints.PointTypeInfo,
	sourceUserId int,
	changeValue int,
	sourceEventId *int) (result typepoints.PointChangeResult, err error) {
	result, err = s.clampedPointChange(userId, partyId, pointType, changeValue)

	if err != nil {
		return
	}

	if pointType.IsShared {
		err = s.Database.ChangePartyPointValueCommand(partyId, pointType.Id, result.ActualChangeValue)
	} else {
		err = s.Database.ChangeUserPointValueCommand(userId, partyId, pointType.Id, result.ActualChangeValue)
	}

	if err != nil || sourceEventId == nil {
		return
	}

	if pointType.IsShared {
		_, err = s.Database.CreatePartyPointHistoryCommand(
			partyId,
			pointType.Id,
			sourceUserId,
			result.DesiredChangeValue,
			result.ActualChangeValue,
			result.FinalValue,
			*sourceEventId)

		return
	}

	_, err = s.Database.CreateUserPointHistoryCommand(
		userId,
		partyId,
		pointType.Id,
		sourceUserId,
		result.DesiredChangeValue,
		result.ActualChangeValue,
		result.FinalValue,
		*sourceEventId)

	return
}

// ChangePointValueByTypeName changes a user's value for the named point type, clamped to the type's
// bounds, and records it against the originating history event.
func (s *Service) ChangePointValueByTypeName(userId int, partyId int, pointTypeName string, changeValue int, sourceEventId int) (err error) {
	pointType, err := s.GetPointTypeByName(partyId, pointTypeName)

	if err != nil {
		return
	}

	_, err = s.changePointValue(userId, partyId, pointType, userId, changeValue, &sourceEventId)

	return
}

// ChangePointValueByTypeNameNoHistory is ChangePointValueByTypeName without a history entry — for
// changes with no event to attach to, such as spending a roll to spin the wheel.
func (s *Service) ChangePointValueByTypeNameNoHistory(userId int, partyId int, pointTypeName string, changeValue int) (err error) {
	pointType, err := s.GetPointTypeByName(partyId, pointTypeName)

	if err != nil {
		return
	}

	_, err = s.changePointValue(userId, partyId, pointType, userId, changeValue, nil)

	return
}

// ChangeUserPointValueClamped changes a value by point type id, clamped and recorded. Change entries
// are applied through this (see srvchanges), which is why it takes an id rather than a name.
func (s *Service) ChangeUserPointValueClamped(userId int, partyId int, pointTypeId int, changeValue int, actorUserId int, sourceEventId int) (err error) {
	pointType, err := s.Database.GetPointTypeCommand(partyId, pointTypeId)

	if err != nil {
		return
	}

	_, err = s.changePointValue(userId, partyId, pointType, actorUserId, changeValue, &sourceEventId)

	return
}

// ChangeUserPointByTypeName applies an administrator's direct change to one user's points, recorded
// against a manual history entry as its source event.
func (s *Service) ChangeUserPointByTypeName(
	actorUserId int,
	userId int,
	partyId int,
	pointTypeName string,
	changeValue int) (result typepoints.PointChangeResult, err error) {
	err = validator.ValidateChangeAmount(changeValue)

	if err != nil {
		return
	}

	pointType, err := s.GetPointTypeByName(partyId, pointTypeName)

	if err != nil {
		return
	}

	if pointType.IsShared {
		err = common.NewSharedPointTypeConflictError(pointTypeName)
		return
	}

	sourceEventId, err := s.createManualSourceEvent(actorUserId, userId, partyId, pointType.Id, changeValue)

	if err != nil {
		return
	}

	return s.changePointValue(userId, partyId, pointType, actorUserId, changeValue, &sourceEventId)
}

// ChangePartyPointByTypeName applies an administrator's direct change to the party pool of a shared
// point type, recorded against a manual history entry as its source event.
func (s *Service) ChangePartyPointByTypeName(
	actorUserId int,
	partyId int,
	pointTypeName string,
	changeValue int) (result typepoints.PointChangeResult, err error) {
	err = validator.ValidateChangeAmount(changeValue)

	if err != nil {
		return
	}

	pointType, err := s.GetPointTypeByName(partyId, pointTypeName)

	if err != nil {
		return
	}

	if !pointType.IsShared {
		err = common.NewNotSharedPointTypeConflictError(pointTypeName)
		return
	}

	sourceEventId, err := s.createManualSourceEvent(actorUserId, actorUserId, partyId, pointType.Id, changeValue)

	if err != nil {
		return
	}

	return s.changePointValue(actorUserId, partyId, pointType, actorUserId, changeValue, &sourceEventId)
}