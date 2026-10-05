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
	GetUserParties(userId int) ([]typeparties.Party, error)
	RequireMember(userId int, partyId int) error
	RequireAdmin(userId int, partyId int) error
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

// GetUserParties lists the parties the user is a member of.
func (s *Service) GetUserParties(userId int) (parties []typeparties.Party, err error) {
	return s.Database.GetUserPartiesCommand(userId)
}

// RequireMember returns an error unless the user is a current member of the party. Anyone else gets
// a not-found error that cannot be told apart from the one for a party that doesn't exist, so the
// answer doesn't reveal which parties exist.
func (s *Service) RequireMember(userId int, partyId int) (err error) {
	_, err = s.requireCurrentMember(userId, partyId)

	return
}

// RequireAdmin returns an error unless the user is a current member of the party and an admin of it.
func (s *Service) RequireAdmin(userId int, partyId int) (err error) {
	member, err := s.requireCurrentMember(userId, partyId)

	if err != nil {
		return
	}

	if !member.IsAdmin {
		err = common.NewNotAdminUnauthorizedError()
	}

	return
}

func (s *Service) requireCurrentMember(userId int, partyId int) (member typeparties.MemberWithLogin, err error) {
	member, err = s.Database.GetMemberCommand(userId, partyId)

	if errors.Is(err, sql.ErrNoRows) || (err == nil && member.LeftDate != nil) {
		err = common.NewPartyNotFoundError(partyId)
	}

	return
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

	member, err = s.Database.CreateMemberCommand(userId, partyId, nilIfEmpty(displayName), isAdmin)

	if err != nil {
		return
	}

	err = s.PointsService.SeedUserPoints(userId, partyId)

	return
}

// nilIfEmpty turns an empty display name into "no display name". The empty string is how a request
// says "remove it", and storing it would also clash with the unique display name within a party.
func nilIfEmpty(displayName *string) *string {
	if displayName != nil && *displayName == "" {
		return nil
	}

	return displayName
}

// ChangeMember updates whichever of a member's display name and admin flag were given. An empty
// display name removes it.
func (s *Service) ChangeMember(userId int, partyId int, displayName *string, isAdmin *bool) (err error) {
	_, err = s.requireMember(userId, partyId)

	if err != nil {
		return
	}

	if displayName != nil {
		err = s.Database.ChangeMemberDisplayNameCommand(userId, partyId, nilIfEmpty(displayName))

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