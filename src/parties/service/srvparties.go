package srvparties

import (
	"FGG-Service/src/common"
	"FGG-Service/src/parties/database"
	"FGG-Service/src/parties/types"
	srvpoints "FGG-Service/src/points/service"
	"database/sql"
	"errors"
)

type IService interface {
	CreateParty(creatorUserId int, name string, displayName string) (typeparties.Party, error)
	GetParties() ([]typeparties.Party, error)
	GetParty(partyId int) (typeparties.Party, error)
	ChangePartyName(partyId int, name string) error
	DeleteParty(partyId int) error
	GetMembers(partyId int) ([]typeparties.MemberWithLogin, error)
	AddMember(userId int, partyId int, displayName string, isAdmin bool) (typeparties.Member, error)
	ChangeMember(userId int, partyId int, displayName *string, isAdmin *bool) error
	RemoveMember(userId int, partyId int) error
}

type Service struct {
	Database      dbparties.IDatabase
	PointsService *srvpoints.Service
}

func NewService() *Service {
	return &Service{
		Database:      new(dbparties.Database),
		PointsService: srvpoints.NewService(),
	}
}

// CreateParty makes a party and puts the user who asked for it in as its first admin. Admin rights
// are membership data, so without this a new party would have nobody able to administer it.
func (s *Service) CreateParty(creatorUserId int, name string, displayName string) (party typeparties.Party, err error) {
	party, err = s.Database.CreatePartyCommand(name)

	if err != nil {
		return
	}

	_, err = s.AddMember(creatorUserId, party.Id, displayName, true)

	return
}

func (s *Service) GetParties() (parties []typeparties.Party, err error) {
	return s.Database.GetPartiesCommand()
}

func (s *Service) GetParty(partyId int) (party typeparties.Party, err error) {
	party, err = s.Database.GetPartyCommand(partyId)

	if errors.Is(err, sql.ErrNoRows) {
		err = common.NewPartyNotFoundError(partyId)
	}

	return
}

func (s *Service) ChangePartyName(partyId int, name string) error {
	return s.Database.ChangePartyNameCommand(partyId, name)
}

func (s *Service) DeleteParty(partyId int) error {
	return s.Database.DeletePartyCommand(partyId)
}

func (s *Service) GetMembers(partyId int) (members []typeparties.MemberWithLogin, err error) {
	return s.Database.GetMembersCommand(partyId)
}

// AddMember joins a user to the party and gives them a starting value for each of its point types.
// Nothing seeds a member's points otherwise, since the schema dropped create_user_stats.
func (s *Service) AddMember(userId int, partyId int, displayName string, isAdmin bool) (member typeparties.Member, err error) {
	doesExist, err := s.Database.DoesMemberExistCommand(userId, partyId)

	if err != nil {
		return
	}

	if doesExist {
		err = common.NewMemberAlreadyExistsConflictError()
		return
	}

	member, err = s.Database.CreateMemberCommand(userId, partyId, displayName, isAdmin)

	if err != nil {
		return
	}

	err = s.PointsService.SeedUserPoints(userId, partyId)

	return
}

// ChangeMember updates whichever of a member's display name and admin flag were given.
func (s *Service) ChangeMember(userId int, partyId int, displayName *string, isAdmin *bool) (err error) {
	err = s.requireMember(userId, partyId)

	if err != nil {
		return
	}

	if displayName != nil {
		err = s.Database.ChangeMemberDisplayNameCommand(userId, partyId, *displayName)

		if err != nil {
			return
		}
	}

	if isAdmin != nil {
		err = s.Database.ChangeMemberAdminStatusCommand(userId, partyId, *isAdmin)
	}

	return
}

func (s *Service) RemoveMember(userId int, partyId int) (err error) {
	err = s.requireMember(userId, partyId)

	if err != nil {
		return
	}

	return s.Database.RemoveMemberCommand(userId, partyId)
}

func (s *Service) requireMember(userId int, partyId int) (err error) {
	doesExist, err := s.Database.DoesMemberExistCommand(userId, partyId)

	if err != nil {
		return
	}

	if !doesExist {
		err = common.NewMemberNotFoundError()
	}

	return
}