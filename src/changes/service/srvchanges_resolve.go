package srvchanges

import (
	"FGG-Service/src/changes/types"
	"FGG-Service/src/common"
)

// ResolveChangeEntries turns the names an API caller uses into the ids a change entry stores. It
// lives here because it is the one place that already reaches every domain an entry can target.
func (s *Service) ResolveChangeEntries(partyId int, inputs []typechanges.ChangeEntryInput) (entries []typechanges.ChangeEntry, err error) {
	entries = make([]typechanges.ChangeEntry, 0, len(inputs))

	for _, input := range inputs {
		var entry typechanges.ChangeEntry
		entry, err = s.resolveChangeEntry(partyId, input)

		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	return
}

func (s *Service) resolveChangeEntry(partyId int, input typechanges.ChangeEntryInput) (entry typechanges.ChangeEntry, err error) {
	entry.Amount = input.Amount

	switch {
	case input.PointTypeName != nil:
		var pointTypeId int
		pointTypeId, err = s.pointTypeIdByName(partyId, *input.PointTypeName)
		entry.PointTypeId = &pointTypeId
	case input.ItemName != nil:
		var itemId int
		itemId, err = s.itemIdByName(partyId, *input.ItemName)
		entry.ItemId = &itemId
	case input.PerkName != nil:
		var perkId int
		perkId, err = s.perkIdByName(partyId, *input.PerkName)
		entry.PerkId = &perkId
	case input.EffectName != nil:
		var effectId int
		effectId, err = s.effectIdByName(partyId, *input.EffectName)
		entry.EffectId = &effectId
	default:
		err = common.NewChangeEntryUnprocessableError()
	}

	return
}

func (s *Service) pointTypeIdByName(partyId int, name string) (id int, err error) {
	pointType, err := s.PointsService.GetPointTypeByName(partyId, name)

	if err != nil {
		return
	}

	return pointType.Id, nil
}

func (s *Service) itemIdByName(partyId int, name string) (id int, err error) {
	items, err := s.ItemsDatabase.GetActualItemsCommand(partyId)

	if err != nil {
		return
	}

	for _, item := range items {
		if item.Name == name {
			return item.Id, nil
		}
	}

	return 0, common.NewItemNotFoundError(name)
}

func (s *Service) perkIdByName(partyId int, name string) (id int, err error) {
	perks, err := s.PerksDatabase.GetActualPerksCommand(partyId)

	if err != nil {
		return
	}

	for _, perk := range perks {
		if perk.Name == name {
			return perk.Id, nil
		}
	}

	return 0, common.NewPerkNotFoundError(name)
}

func (s *Service) effectIdByName(partyId int, name string) (id int, err error) {
	effects, err := s.EffectsDatabase.GetActualEffectsCommand(partyId)

	if err != nil {
		return
	}

	for _, effect := range effects {
		if effect.Name == name {
			return effect.Id, nil
		}
	}

	return 0, common.NewEffectNotFoundError(name)
}