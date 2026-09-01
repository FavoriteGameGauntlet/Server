package srvpoints

import (
	"database/sql"
	"errors"
	"fmt"
)

// getPointTypeIdByName resolves a party.PointTypes row's Id from its Name. There is no by-name lookup
// in the dbaccess layer, so this scans GetPointTypesCommand's result.
func (s *Service) getPointTypeIdByName(partyId int, name string) (id int, err error) {
	pointTypes, err := s.Database.GetPointTypesCommand(partyId)

	if err != nil {
		return
	}

	for _, pointType := range pointTypes {
		if pointType.Name == name {
			return pointType.Id, nil
		}
	}

	err = fmt.Errorf("point type %q not found for party %d", name, partyId)

	return
}

// GetPointValueByTypeName reads a user's current value for the named PointType (e.g.
// typepoints.PointTypeAvailableRolls). Returns 0 if the user has no row for it yet.
func (s *Service) GetPointValueByTypeName(userId int, partyId int, pointTypeName string) (value int, err error) {
	pointTypeId, err := s.getPointTypeIdByName(partyId, pointTypeName)

	if err != nil {
		return
	}

	point, err := s.Database.GetUserPointCommand(userId, partyId, pointTypeId)

	if errors.Is(err, sql.ErrNoRows) {
		err = nil
		return
	}

	if err != nil {
		return
	}

	value = point.Value

	return
}

// clampedPointChange computes the value actually applied when changeValue is clamped to the
// PointType's Minimum/Maximum, and the resulting final value.
func (s *Service) clampedPointChange(userId int, partyId int, pointTypeId int, changeValue int) (actualChangeValue int, finalValue int, err error) {
	pointType, err := s.Database.GetPointTypeCommand(partyId, pointTypeId)

	if err != nil {
		return
	}

	currentPoint, err := s.Database.GetUserPointCommand(userId, partyId, pointTypeId)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return
	}

	err = nil

	actualChangeValue = changeValue
	finalValue = currentPoint.Value + changeValue

	if finalValue < pointType.Minimum {
		finalValue = pointType.Minimum
		actualChangeValue = finalValue - currentPoint.Value
	}

	if finalValue > pointType.Maximum {
		finalValue = pointType.Maximum
		actualChangeValue = finalValue - currentPoint.Value
	}

	return
}

// ChangePointValueByTypeName changes a user's value for the named PointType, clamped to the type's
// Minimum/Maximum, and records the change in its history using sourceEventId as the originating
// HistoryEvents row.
func (s *Service) ChangePointValueByTypeName(userId int, partyId int, pointTypeName string, changeValue int, sourceEventId int) (err error) {
	pointTypeId, err := s.getPointTypeIdByName(partyId, pointTypeName)

	if err != nil {
		return
	}

	return s.ChangeUserPointValueClamped(userId, partyId, pointTypeId, changeValue, sourceEventId)
}

// ChangePointValueByTypeNameNoHistory is like ChangePointValueByTypeName but does not record a
// history entry — for changes with no HistoryEvents row to attach to (e.g. deducting a roll when a
// wheel roll is initiated, which itself produces no history event).
func (s *Service) ChangePointValueByTypeNameNoHistory(userId int, partyId int, pointTypeName string, changeValue int) (err error) {
	pointTypeId, err := s.getPointTypeIdByName(partyId, pointTypeName)

	if err != nil {
		return
	}

	actualChangeValue, _, err := s.clampedPointChange(userId, partyId, pointTypeId, changeValue)

	if err != nil {
		return
	}

	return s.Database.ChangeUserPointValueCommand(userId, partyId, pointTypeId, actualChangeValue)
}

// ChangeUserPointValueClamped changes a user's value for the given PointType, clamped to the type's
// Minimum/Maximum, and records the change in its history using sourceEventId as the originating
// HistoryEvents row. Used to apply Change entries generically (see srvchanges).
func (s *Service) ChangeUserPointValueClamped(userId int, partyId int, pointTypeId int, changeValue int, sourceEventId int) (err error) {
	actualChangeValue, finalValue, err := s.clampedPointChange(userId, partyId, pointTypeId, changeValue)

	if err != nil {
		return
	}

	err = s.Database.ChangeUserPointValueCommand(userId, partyId, pointTypeId, actualChangeValue)

	if err != nil {
		return
	}

	_, err = s.Database.CreateUserPointHistoryCommand(
		userId,
		partyId,
		pointTypeId,
		userId,
		changeValue,
		actualChangeValue,
		finalValue,
		sourceEventId)

	return
}
