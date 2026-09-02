package srvpoints

import (
	"FGG-Service/src/changes/types"
	"FGG-Service/src/common"
	"FGG-Service/src/history/database"
	"FGG-Service/src/parties/database"
	"FGG-Service/src/points/database"
	"FGG-Service/src/points/type"
	"database/sql"
	"errors"
)

type Service struct {
	Database        dbpoints.IDatabase
	PartiesDatabase dbparties.IDatabase
	HistoryDatabase dbhistory.IDatabase
}

func NewService() *Service {
	return &Service{
		Database:        new(dbpoints.Database),
		PartiesDatabase: new(dbparties.Database),
		HistoryDatabase: new(dbhistory.Database),
	}
}

func (s *Service) GetPointTypes(partyId int) (pointTypes []typepoints.PointTypeInfo, err error) {
	return s.Database.GetPointTypesCommand(partyId)
}

// CreatePointType adds a point type to the party. Names identify point types across the API, so a
// duplicate is rejected rather than silently shadowing the existing one.
func (s *Service) CreatePointType(partyId int, pointType typepoints.PointType) (created typepoints.PointType, err error) {
	doesExist, err := s.Database.DoesPointTypeExistCommand(partyId, pointType.Name)

	if err != nil {
		return
	}

	if doesExist {
		err = common.NewPointTypeAlreadyExistsConflictError(pointType.Name)
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

// ChangePointType updates the named point type. The name itself is the API identity of the type and
// is not editable.
func (s *Service) ChangePointType(partyId int, name string, pointType typepoints.PointType) (err error) {
	current, err := s.GetPointTypeByName(partyId, name)

	if err != nil {
		return
	}

	return s.Database.ChangePointTypeCommand(
		partyId,
		current.Id,
		name,
		pointType.Description,
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
	pointTypes, err := s.Database.GetPointTypesCommand(partyId)

	if err != nil {
		return
	}

	for _, candidate := range pointTypes {
		if candidate.Name == name {
			return candidate, nil
		}
	}

	err = common.NewPointTypeNotFoundError(name)

	return
}
// SeedUserPoints gives a user a starting value for every point type of the party. The schema no
// longer seeds points on signup, so this runs when a user joins a party.
func (s *Service) SeedUserPoints(userId int, partyId int) (err error) {
	pointTypes, err := s.Database.GetPointTypesCommand(partyId)

	if err != nil {
		return
	}

	for _, pointType := range pointTypes {
		if pointType.IsShared {
			continue
		}

		_, err = s.Database.CreateUserPointCommand(userId, partyId, pointType.Id, pointType.StartValue)

		if err != nil {
			return
		}
	}

	return
}

// GetUserPoints lists every point type of the party with the value this user holds for it. A shared
// point type carries the party-wide value, and a type the user has no row for yet reads as its
// starting value.
func (s *Service) GetUserPoints(userId int, partyId int) (values []typepoints.UserPointValue, err error) {
	pointTypes, err := s.Database.GetPointTypesCommand(partyId)

	if err != nil {
		return
	}

	values = make([]typepoints.UserPointValue, 0, len(pointTypes))

	for _, pointType := range pointTypes {
		var value int
		value, err = s.pointValue(userId, partyId, pointType)

		if err != nil {
			return
		}

		values = append(values, typepoints.UserPointValue{PointType: pointType, Value: value})
	}

	return
}

// GetAllUserPoints lists the points of every current member of the party.
func (s *Service) GetAllUserPoints(partyId int) (byLogin []typepoints.UserPointValuesByLogin, err error) {
	members, err := s.PartiesDatabase.GetMembersCommand(partyId)

	if err != nil {
		return
	}

	byLogin = make([]typepoints.UserPointValuesByLogin, 0, len(members))

	for _, member := range members {
		if member.LeftDate != nil {
			continue
		}

		var values []typepoints.UserPointValue
		values, err = s.GetUserPoints(member.UserId, partyId)

		if err != nil {
			return
		}

		byLogin = append(byLogin, typepoints.UserPointValuesByLogin{Login: member.Login, Points: values})
	}

	return
}

// pointValue reads the value held for one point type, from the party pool when the type is shared.
func (s *Service) pointValue(userId int, partyId int, pointType typepoints.PointTypeInfo) (value int, err error) {
	if pointType.IsShared {
		var partyPoint typepoints.PartyPoint
		partyPoint, err = s.Database.GetPartyPointCommand(partyId, pointType.Id)

		if errors.Is(err, sql.ErrNoRows) {
			return pointType.StartValue, nil
		}

		return partyPoint.Value, err
	}

	point, err := s.Database.GetUserPointCommand(userId, partyId, pointType.Id)

	if errors.Is(err, sql.ErrNoRows) {
		return pointType.StartValue, nil
	}

	return point.Value, err
}

func (s *Service) GetPartyPointValueByTypeName(partyId int, name string) (value int, err error) {
	pointType, err := s.GetPointTypeByName(partyId, name)

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
func (s *Service) GetUserPointHistoryByTypeName(userId int, partyId int, name string) (history []typepoints.PointHistoryEntry, err error) {
	pointType, err := s.GetPointTypeByName(partyId, name)

	if err != nil {
		return
	}

	entries, err := s.Database.GetUserPointHistoryCommand(userId, partyId)

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
			SourceUserId:       entry.SourceUserId,
			ChangedDate:        entry.ChangedDate,
		})
	}

	return
}

// GetPartyPointHistoryByTypeName returns the recorded changes to the party pool for one point type.
func (s *Service) GetPartyPointHistoryByTypeName(partyId int, name string) (history []typepoints.PointHistoryEntry, err error) {
	pointType, err := s.GetPointTypeByName(partyId, name)

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
			SourceUserId:       entry.SourceUserId,
			ChangedDate:        entry.ChangedDate,
		})
	}

	return
}

// createManualSourceEvent records a direct point change by an administrator as a manual history
// entry, whose id becomes the source event of the point history the change produces. Every recorded
// point change names the event that caused it, and a direct change has no other event behind it.
func (s *Service) createManualSourceEvent(actorUserId int, userId int, partyId int, pointTypeId int, changeValue int) (sourceEventId int, err error) {
	entries := []typechanges.ChangeEntry{
		{Amount: changeValue, PointTypeId: &pointTypeId, UserId: &userId},
	}

	created, err := s.HistoryDatabase.CreateManualHistoryCommand(partyId, actorUserId, entries, nil)

	if err != nil {
		return
	}

	if len(created) == 0 {
		err = errors.New("manual history recorded no entries")
		return
	}

	return created[0].Id, nil
}
// GetPointTypeById reads a point type by the id a change entry stores.
func (s *Service) GetPointTypeById(partyId int, pointTypeId int) (pointType typepoints.PointTypeInfo, err error) {
	return s.Database.GetPointTypeCommand(partyId, pointTypeId)
}