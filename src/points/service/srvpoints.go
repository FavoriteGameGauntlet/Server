package srvpoints

import (
	"FGG-Service/src/changes/database"
	"FGG-Service/src/changes/types"
	"FGG-Service/src/common"
	"FGG-Service/src/grants/database"
	"FGG-Service/src/parties/database"
	"FGG-Service/src/points/database"
	"FGG-Service/src/points/type"
	"FGG-Service/src/validator"
	"database/sql"
	"errors"
)

type Service struct {
	Database        dbpoints.IDatabase
	PartiesDatabase dbparties.IDatabase
	GrantsDatabase  dbgrants.IDatabase
	ChangesDatabase dbchanges.IDatabase
}

func NewService() *Service {
	return &Service{
		Database:        new(dbpoints.Database),
		PartiesDatabase: new(dbparties.Database),
		GrantsDatabase:  new(dbgrants.Database),
		ChangesDatabase: new(dbchanges.Database),
	}
}

func (s *Service) GetPointTypes(partyId int) (pointTypes []typepoints.PointTypeInfo, err error) {
	return s.Database.GetPointTypesCommand(partyId)
}

// GetVisiblePointTypes lists the point types the user may see. A type that is not public is left out
// unless the user is an admin of the party.
func (s *Service) GetVisiblePointTypes(actorUserId int, partyId int) (pointTypes []typepoints.PointTypeInfo, err error) {
	pointTypes, err = s.Database.GetPointTypesCommand(partyId)

	if err != nil {
		return
	}

	isAdmin, err := s.isAdmin(actorUserId, partyId)

	if err != nil || isAdmin {
		return
	}

	visible := make([]typepoints.PointTypeInfo, 0, len(pointTypes))

	for _, pointType := range pointTypes {
		if pointType.IsPublic {
			visible = append(visible, pointType)
		}
	}

	return visible, nil
}

// isAdmin reports whether the user is an admin of the party. A user who is not a member of the party
// is simply not an admin.
func (s *Service) isAdmin(actorUserId int, partyId int) (isAdmin bool, err error) {
	member, err := s.PartiesDatabase.GetMemberCommand(actorUserId, partyId)

	if errors.Is(err, sql.ErrNoRows) {
		err = nil
		return
	}

	if err != nil {
		return
	}

	return member.IsAdmin, nil
}

// CreatePointType adds a point type to the party. Names identify point types across the API, so a
// duplicate is rejected rather than silently shadowing the existing one.
func (s *Service) CreatePointType(partyId int, pointType typepoints.PointType) (created typepoints.PointType, err error) {
	err = validator.ValidatePointTypeBounds(pointType.StartValue, pointType.Minimum, pointType.Maximum)

	if err != nil {
		return
	}

	_, err = s.Database.GetPointTypeByNameCommand(partyId, pointType.Name)

	if err == nil {
		err = common.NewPointTypeAlreadyExistsConflictError(pointType.Name)
		return
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return
	}

	return s.Database.CreatePointTypeCommand(
		partyId,
		pointType.Name,
		pointType.Description,
		pointType.StartValue,
		pointType.IsPublic,
		pointType.IsShared,
		pointType.Minimum,
		pointType.Maximum)
}

// ChangePointType updates the named point type, including its name and start value. A new name must
// not be taken by another point type, since names identify point types across the API. A new start
// value only affects values not held yet: a point already held keeps its value.
func (s *Service) ChangePointType(partyId int, name string, pointType typepoints.PointType) (err error) {
	current, err := s.GetPointTypeByName(partyId, name)

	if err != nil {
		return
	}

	err = validator.ValidatePointTypeBounds(pointType.StartValue, pointType.Minimum, pointType.Maximum)

	if err != nil {
		return
	}

	if pointType.Name != current.Name {
		_, err = s.Database.GetPointTypeByNameCommand(partyId, pointType.Name)

		if err == nil {
			err = common.NewPointTypeAlreadyExistsConflictError(pointType.Name)
			return
		}

		if !errors.Is(err, sql.ErrNoRows) {
			return
		}
	}

	return s.Database.ChangePointTypeCommand(
		partyId,
		current.Id,
		pointType.Name,
		pointType.Description,
		pointType.StartValue,
		pointType.IsPublic,
		pointType.IsShared,
		pointType.Minimum,
		pointType.Maximum)
}

func (s *Service) RemovePointType(partyId int, name string) (err error) {
	pointType, err := s.GetPointTypeByName(partyId, name)

	if err != nil {
		return
	}

	return s.Database.RemovePointTypeCommand(partyId, pointType.Id)
}

// GetPointTypeByName resolves a point type from the name the API addresses it by.
func (s *Service) GetPointTypeByName(partyId int, name string) (pointType typepoints.PointTypeInfo, err error) {
	pointType, err = s.Database.GetPointTypeByNameCommand(partyId, name)

	if errors.Is(err, sql.ErrNoRows) {
		err = common.NewPointTypeNotFoundError(name)
	}

	return
}

// getVisiblePointTypeByName resolves a point type for a user. A type that is not public reads as not
// found unless the user is an admin of the party.
func (s *Service) getVisiblePointTypeByName(actorUserId int, partyId int, name string) (pointType typepoints.PointTypeInfo, err error) {
	pointType, err = s.GetPointTypeByName(partyId, name)

	if err != nil || pointType.IsPublic {
		return
	}

	isAdmin, err := s.isAdmin(actorUserId, partyId)

	if err == nil && !isAdmin {
		err = common.NewPointTypeNotFoundError(name)
	}

	return
}

// getUserPointTypeByName resolves a point type that users hold values of, rejecting a shared one.
func (s *Service) getUserPointTypeByName(actorUserId int, partyId int, name string) (pointType typepoints.PointTypeInfo, err error) {
	pointType, err = s.getVisiblePointTypeByName(actorUserId, partyId, name)

	if err == nil && pointType.IsShared {
		err = common.NewSharedPointTypeConflictError(name)
	}

	return
}

// getSharedPointTypeByName resolves a point type the party holds the value of, rejecting any other.
func (s *Service) getSharedPointTypeByName(actorUserId int, partyId int, name string) (pointType typepoints.PointTypeInfo, err error) {
	pointType, err = s.getVisiblePointTypeByName(actorUserId, partyId, name)

	if err == nil && !pointType.IsShared {
		err = common.NewNotSharedPointTypeConflictError(name)
	}

	return
}
// SeedUserPoints gives a user a starting value for every point type of the party. The schema no
// longer seeds points on signup, so this runs when a user joins a party.
func (s *Service) SeedUserPoints(affectedUserId int, partyId int) (err error) {
	pointTypes, err := s.Database.GetPointTypesCommand(partyId)

	if err != nil {
		return
	}

	for _, pointType := range pointTypes {
		if pointType.IsShared {
			continue
		}

		_, err = s.Database.CreateUserPointCommand(affectedUserId, partyId, pointType.Id, pointType.StartValue)

		if err != nil {
			return
		}
	}

	return
}

// GetUserPoints lists every point type of the party with the value the affected user holds for it.
// Shared point types are left out, since their value belongs to the party rather than to the user (see
// GetAllPartyPoints). A type the affected user has no row for yet reads as its starting value. A type
// that is not public is left out unless the user asking is an admin.
func (s *Service) GetUserPoints(actorUserId int, affectedUserId int, partyId int) (values []typepoints.PointValue, err error) {
	pointTypes, err := s.GetVisiblePointTypes(actorUserId, partyId)

	if err != nil {
		return
	}

	return s.userPointValues(affectedUserId, partyId, pointTypes)
}

// userPointValues reads the user's value for each of the point types that is not shared.
func (s *Service) userPointValues(affectedUserId int, partyId int, pointTypes []typepoints.PointTypeInfo) (values []typepoints.PointValue, err error) {
	values = make([]typepoints.PointValue, 0, len(pointTypes))

	for _, pointType := range pointTypes {
		if pointType.IsShared {
			continue
		}

		var value int
		value, err = s.pointValue(affectedUserId, partyId, pointType)

		if err != nil {
			return
		}

		values = append(values, typepoints.PointValue{PointType: pointType, Value: value})
	}

	return
}

// GetAllUserPoints lists the points of every current member of the party. Shared point types are left
// out, since their value belongs to the party rather than to any member (see GetAllPartyPoints). A
// type that is not public is left out unless the user asking is an admin.
func (s *Service) GetAllUserPoints(actorUserId int, partyId int) (byLogin []typepoints.UserPointValuesByLogin, err error) {
	members, err := s.PartiesDatabase.GetMembersCommand(partyId)

	if err != nil {
		return
	}

	pointTypes, err := s.GetVisiblePointTypes(actorUserId, partyId)

	if err != nil {
		return
	}

	byLogin = make([]typepoints.UserPointValuesByLogin, 0, len(members))

	for _, member := range members {
		if member.LeftDate != nil {
			continue
		}

		var values []typepoints.PointValue
		values, err = s.userPointValues(member.UserId, partyId, pointTypes)

		if err != nil {
			return
		}

		byLogin = append(byLogin, typepoints.UserPointValuesByLogin{Login: member.Login, Points: values})
	}

	return
}

// GetAllPartyPoints lists every shared point type of the party with the value the party holds for it.
// A type that is not public is left out unless the user asking is an admin.
func (s *Service) GetAllPartyPoints(actorUserId int, partyId int) (values []typepoints.PointValue, err error) {
	pointTypes, err := s.GetVisiblePointTypes(actorUserId, partyId)

	if err != nil {
		return
	}

	values = make([]typepoints.PointValue, 0, len(pointTypes))

	for _, pointType := range pointTypes {
		if !pointType.IsShared {
			continue
		}

		var value int
		value, err = s.pointValue(0, partyId, pointType)

		if err != nil {
			return
		}

		values = append(values, typepoints.PointValue{PointType: pointType, Value: value})
	}

	return
}

// pointValue reads the value held for one point type, from the party pool when the type is shared.
func (s *Service) pointValue(affectedUserId int, partyId int, pointType typepoints.PointTypeInfo) (value int, err error) {
	if pointType.IsShared {
		var partyPoint typepoints.PartyPoint
		partyPoint, err = s.Database.GetPartyPointCommand(partyId, pointType.Id)

		if errors.Is(err, sql.ErrNoRows) {
			return pointType.StartValue, nil
		}

		return partyPoint.Value, err
	}

	point, err := s.Database.GetUserPointCommand(affectedUserId, partyId, pointType.Id)

	if errors.Is(err, sql.ErrNoRows) {
		return pointType.StartValue, nil
	}

	return point.Value, err
}

func (s *Service) GetPartyPointValueByTypeName(actorUserId int, partyId int, name string) (value int, err error) {
	pointType, err := s.getSharedPointTypeByName(actorUserId, partyId, name)

	if err != nil {
		return
	}

	partyPoint, err := s.Database.GetPartyPointCommand(partyId, pointType.Id)

	if errors.Is(err, sql.ErrNoRows) {
		return pointType.StartValue, nil
	}

	return partyPoint.Value, err
}
// GetUserPointHistoryByTypeName returns the recorded changes to a user's value for one point type.
func (s *Service) GetUserPointHistoryByTypeName(actorUserId int, affectedUserId int, partyId int, name string) (history []typepoints.PointHistoryEntry, err error) {
	pointType, err := s.getUserPointTypeByName(actorUserId, partyId, name)

	if err != nil {
		return
	}

	entries, err := s.Database.GetUserPointHistoryCommand(affectedUserId, partyId)

	if err != nil {
		return
	}

	history = make([]typepoints.PointHistoryEntry, 0, len(entries))

	for _, entry := range entries {
		if entry.PointTypeId != pointType.Id {
			continue
		}

		history = append(history, typepoints.PointHistoryEntry{
			DesiredChangeValue: entry.DesiredChangeValue,
			ActualChangeValue:  entry.ActualChangeValue,
			FinalValue:         entry.FinalValue,
			ActorUserId:       entry.ActorUserId,
			ChangedDate:        entry.ChangedDate,
		})
	}

	return
}

// GetPartyPointHistoryByTypeName returns the recorded changes to the party pool for one point type.
func (s *Service) GetPartyPointHistoryByTypeName(actorUserId int, partyId int, name string) (history []typepoints.PointHistoryEntry, err error) {
	pointType, err := s.getSharedPointTypeByName(actorUserId, partyId, name)

	if err != nil {
		return
	}

	entries, err := s.Database.GetPartyPointHistoryCommand(partyId)

	if err != nil {
		return
	}

	history = make([]typepoints.PointHistoryEntry, 0, len(entries))

	for _, entry := range entries {
		if entry.PointTypeId != pointType.Id {
			continue
		}

		history = append(history, typepoints.PointHistoryEntry{
			DesiredChangeValue: entry.DesiredChangeValue,
			ActualChangeValue:  entry.ActualChangeValue,
			FinalValue:         entry.FinalValue,
			ActorUserId:       entry.ActorUserId,
			ChangedDate:        entry.ChangedDate,
		})
	}

	return
}

// createManualSourceEvent records a direct point change made by a user as a manual history
// entry, whose id becomes the source event of the point history the change produces. Every recorded
// point change names the event that caused it, and a direct change has no other event behind it.
func (s *Service) createManualSourceEvent(actorUserId int, affectedUserId int, partyId int, pointTypeId int, changeValue int) (sourceEventId int, err error) {
	entries := []typechanges.ChangeEntry{
		{Amount: changeValue, PointTypeId: &pointTypeId, UserId: &affectedUserId},
	}

	userChange, err := s.ChangesDatabase.CreateUserChangeFromJsonbCommand(partyId, entries)

	if err != nil {
		return
	}

	history, err := s.GrantsDatabase.CreateManualHistoryCommand(affectedUserId, partyId, *userChange.ChangeId, actorUserId, nil)

	if err != nil {
		return
	}

	return history.Id, nil
}
// GetPointTypeById reads a point type by the id a change entry stores.
func (s *Service) GetPointTypeById(partyId int, pointTypeId int) (pointType typepoints.PointTypeInfo, err error) {
	return s.Database.GetPointTypeCommand(partyId, pointTypeId)
}