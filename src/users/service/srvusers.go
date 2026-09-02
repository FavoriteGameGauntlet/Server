package srvusers

import (
	"FGG-Service/src/common"
	"FGG-Service/src/parties/database"
	"FGG-Service/src/users/types"
	"database/sql"
	"errors"
)

// defaultPartyId is a stopgap until real party context exists (see project plan) — display names are
// read from membership of this one hardcoded party.
const defaultPartyId = 1

type Service struct {
	PartiesDatabase dbparties.IDatabase
}

func NewService() *Service {
	return &Service{
		PartiesDatabase: new(dbparties.Database),
	}
}

// GetAllUserNames lists the members of the party by login and display name. A display name belongs
// to a membership rather than to the user, since the same user can go by a different name in each
// party they are in.
func (s *Service) GetAllUserNames() (users typeusers.Users, err error) {
	members, err := s.PartiesDatabase.GetMembersCommand(defaultPartyId)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return
	}

	for _, member := range members {
		if member.LeftDate != nil {
			continue
		}

		displayName := member.DisplayName

		users = append(users, typeusers.User{Login: member.Login, DisplayName: &displayName})
	}

	return
}

func (s *Service) ChangeDisplayName(userId int, displayName string) error {
	return s.PartiesDatabase.ChangeMemberDisplayNameCommand(userId, defaultPartyId, displayName)
}

func (s *Service) GetDisplayName(userId int) (displayName *string, err error) {
	member, err := s.PartiesDatabase.GetMemberCommand(userId, defaultPartyId)

	if errors.Is(err, sql.ErrNoRows) || err == nil && member.DisplayName == "" {
		return nil, common.NewDisplayNameNotFoundError()
	}

	if err != nil {
		return
	}

	name := member.DisplayName

	return &name, nil
}