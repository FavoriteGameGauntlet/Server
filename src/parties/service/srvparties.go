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
	CreateParty(creatorUserId int, name string) (typeparties.Party, error)
	GetParties() ([]typeparties.Party, error)
	GetParty(partyId int) (typeparties.Party, error)
	ChangePartyName(partyId int, name string) error
	GetMembers(partyId int) ([]typeparties.MemberWithLogin, error)
	AddMember(userId int, partyId int, displayName *string, isAdmin bool) (typeparties.Member, error)
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
func (s *Service) CreateParty(creatorUserId int, name string) (party typeparties.Party, err error) {
	party, err = s.Database.CreatePartyCommand(name)

	if err != nil {
		return
	}

	_, err = s.AddMember(creatorUserId, party.Id, nil, true)

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

func (s *Service) GetMembers(partyId int) (members []typeparties.MemberWithLogin, err error) {
	return s.Database.GetMembersCommand(partyId)
}

// AddMember joins a user to the party and gives them a starting value for each of its point types.
// Nothing seeds a member's points otherwise, since the schema dropped create_user_stats.
func (s *Service) AddMember(userId int, partyId int, displayName *string, isAdmin bool) (member typeparties.Member, err error) {
	_, err = s.Database.GetMemberCommand(userId, partyId)

	if err == nil {
		err = common.NewMemberAlreadyExistsConflictError()
		return
	}

	if !errors.Is(err, sql.ErrNoRows) {
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
	_, err = s.requireMember(userId, partyId)

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

// RemoveMember takes a user out of the party. If that leaves the party without an admin, the
// remaining member who joined earliest is made one, so the party never has nobody to administer it.
func (s *Service) RemoveMember(userId int, partyId int) (err error) {
	leaving, err := s.requireMember(userId, partyId)

	if err != nil {
		return
	}

	if leaving.IsAdmin {
		var successor *typeparties.MemberWithLogin

		successor, err = s.findAdminSuccessor(userId, partyId)

		if err != nil {
			return
		}

		if successor != nil {
			err = s.Database.ChangeMemberAdminStatusCommand(successor.UserId, partyId, true)

			if err != nil {
				return
			}
		}
	}

	return s.Database.RemoveMemberCommand(userId, partyId)
}

// findAdminSuccessor returns the member to promote when the admin leavingUserId leaves the party,
// or nil if another admin remains or nobody remains.
func (s *Service) findAdminSuccessor(leavingUserId int, partyId int) (successor *typeparties.MemberWithLogin, err error) {
	members, err := s.Database.GetMembersCommand(partyId)

	if err != nil {
		return
	}

	for i := range members {
		member := &members[i]

		if member.LeftDate != nil || member.UserId == leavingUserId {
			continue
		}

		if member.IsAdmin {
			return nil, nil
		}

		if successor == nil || member.JoinedDate.Before(successor.JoinedDate) {
			successor = member
		}
	}

	return
}

func (s *Service) requireMember(userId int, partyId int) (member typeparties.MemberWithLogin, err error) {
	member, err = s.Database.GetMemberCommand(userId, partyId)

	if errors.Is(err, sql.ErrNoRows) {
		err = common.NewMemberNotFoundError()
	}

	return
}