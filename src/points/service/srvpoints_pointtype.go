package srvpoints

import (
	"FGG-Service/src/common"
	"FGG-Service/src/points/type"
	"FGG-Service/src/validator"
	"context"
	"math"
)

// GetPointValueByTypeName reads a user's current value for the named point type.
func (s *Service) GetPointValueByTypeName(ctx context.Context, affectedUserId int, partyId int, pointTypeName string) (value int, err error) {
	pointType, err := s.GetPointTypeByName(ctx, partyId, pointTypeName)

	if err != nil {
		return
	}

	return s.pointValue(ctx, affectedUserId, partyId, pointType)
}

// GetUserPointValueByTypeName reads a user's own value for the named point type. A shared point type
// has no value of the user's own, so it is rejected rather than read from the party pool. A type that
// is not public reads as not found unless the user asking is an admin.
func (s *Service) GetUserPointValueByTypeName(ctx context.Context, actorUserId int, affectedUserId int, partyId int, pointTypeName string) (value int, err error) {
	pointType, err := s.getUserPointTypeByName(ctx, actorUserId, partyId, pointTypeName)

	if err != nil {
		return
	}

	return s.pointValue(ctx, affectedUserId, partyId, pointType)
}

// clampedPointChange computes the value actually applied once changeValue is clamped to the point
// type's Minimum/Maximum, and the value the point ends up at. A nil bound falls back to the range of
// the INTEGER column the value is stored in, so an unbounded point clamps instead of overflowing.
func (s *Service) clampedPointChange(ctx context.Context, affectedUserId int, partyId int, pointType typepoints.PointTypeInfo, changeValue int) (result typepoints.PointChangeResult, err error) {
	currentValue, err := s.pointValue(ctx, affectedUserId, partyId, pointType)

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
	ctx context.Context,
	affectedUserId int,
	partyId int,
	pointType typepoints.PointTypeInfo,
	actorUserId int,
	changeValue int,
	sourceEventId *int) (result typepoints.PointChangeResult, err error) {
	result, err = s.clampedPointChange(ctx, affectedUserId, partyId, pointType, changeValue)

	if err != nil {
		return
	}

	if pointType.IsShared {
		err = s.Database.ChangePartyPointValueCommand(ctx, partyId, pointType.Id, result.ActualChangeValue)
	} else {
		err = s.Database.ChangeUserPointValueCommand(ctx, affectedUserId, partyId, pointType.Id, result.ActualChangeValue)
	}

	if err != nil || sourceEventId == nil {
		return
	}

	if pointType.IsShared {
		_, err = s.Database.CreatePartyPointHistoryCommand(
			ctx,
			partyId,
			pointType.Id,
			actorUserId,
			result.DesiredChangeValue,
			result.ActualChangeValue,
			result.FinalValue,
			*sourceEventId)

		return
	}

	_, err = s.Database.CreateUserPointHistoryCommand(
		ctx,
		affectedUserId,
		partyId,
		pointType.Id,
		actorUserId,
		result.DesiredChangeValue,
		result.ActualChangeValue,
		result.FinalValue,
		*sourceEventId)

	return
}

// ChangePointValueByTypeName changes a user's value for the named point type, clamped to the type's
// bounds, and records it against the originating history event.
func (s *Service) ChangePointValueByTypeName(ctx context.Context, affectedUserId int, partyId int, pointTypeName string, changeValue int, sourceEventId int) (err error) {
	pointType, err := s.GetPointTypeByName(ctx, partyId, pointTypeName)

	if err != nil {
		return
	}

	_, err = s.changePointValue(ctx, affectedUserId, partyId, pointType, affectedUserId, changeValue, &sourceEventId)

	return
}

// ChangePointValueByTypeNameNoHistory is ChangePointValueByTypeName without a history entry — for
// changes with no event to attach to, such as spending a roll to spin the wheel.
func (s *Service) ChangePointValueByTypeNameNoHistory(ctx context.Context, affectedUserId int, partyId int, pointTypeName string, changeValue int) (err error) {
	pointType, err := s.GetPointTypeByName(ctx, partyId, pointTypeName)

	if err != nil {
		return
	}

	_, err = s.changePointValue(ctx, affectedUserId, partyId, pointType, affectedUserId, changeValue, nil)

	return
}

// ChangeUserPointValueClamped changes a value by point type id, clamped and recorded. Change entries
// are applied through this (see srvchanges), which is why it takes an id rather than a name.
func (s *Service) ChangeUserPointValueClamped(ctx context.Context, affectedUserId int, partyId int, pointTypeId int, changeValue int, actorUserId int, sourceEventId int) (err error) {
	pointType, err := s.Database.GetPointTypeCommand(ctx, partyId, pointTypeId)

	if err != nil {
		return
	}

	_, err = s.changePointValue(ctx, affectedUserId, partyId, pointType, actorUserId, changeValue, &sourceEventId)

	return
}

// ChangeUserPointByTypeName applies a direct change made by the acting user to one user's points, recorded
// against a manual history entry as its source event.
func (s *Service) ChangeUserPointByTypeName(
	ctx context.Context,
	actorUserId int,
	affectedUserId int,
	partyId int,
	pointTypeName string,
	changeValue int) (result typepoints.PointChangeResult, err error) {
	err = validator.ValidateChangeAmount(changeValue)

	if err != nil {
		return
	}

	pointType, err := s.GetPointTypeByName(ctx, partyId, pointTypeName)

	if err != nil {
		return
	}

	if pointType.IsShared {
		err = common.NewSharedPointTypeConflictError(pointTypeName)
		return
	}

	sourceEventId, err := s.createManualSourceEvent(ctx, actorUserId, affectedUserId, partyId, pointType.Id, changeValue)

	if err != nil {
		return
	}

	return s.changePointValue(ctx, affectedUserId, partyId, pointType, actorUserId, changeValue, &sourceEventId)
}

// ChangePartyPointByTypeName applies a direct change made by the acting user to the party pool of a shared
// point type, recorded against a manual history entry as its source event.
func (s *Service) ChangePartyPointByTypeName(
	ctx context.Context,
	actorUserId int,
	partyId int,
	pointTypeName string,
	changeValue int) (result typepoints.PointChangeResult, err error) {
	err = validator.ValidateChangeAmount(changeValue)

	if err != nil {
		return
	}

	pointType, err := s.GetPointTypeByName(ctx, partyId, pointTypeName)

	if err != nil {
		return
	}

	if !pointType.IsShared {
		err = common.NewNotSharedPointTypeConflictError(pointTypeName)
		return
	}

	sourceEventId, err := s.createManualSourceEvent(ctx, actorUserId, actorUserId, partyId, pointType.Id, changeValue)

	if err != nil {
		return
	}

	return s.changePointValue(ctx, actorUserId, partyId, pointType, actorUserId, changeValue, &sourceEventId)
}
