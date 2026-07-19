package typechanges

type ChangeEntry struct {
	EntryId     *int `json:"entry_id,omitempty"`
	Amount      int  `json:"amount"`
	PointTypeId *int `json:"point_type_id"`
	ItemId      *int `json:"item_id"`
	PerkId      *int `json:"perk_id"`
	EffectId    *int `json:"effect_id"`
	UserId      *int `json:"user_id,omitempty"`
}

type Change struct {
	ChangeId         *int          `json:"change_id,omitempty"`
	ShouldApplyToAll bool          `json:"should_apply_to_all"`
	IsManualChange   bool          `json:"is_manual_change"`
	Entries          []ChangeEntry `json:"entries"`
}

type UserChange struct {
	ChangeId *int          `json:"change_id,omitempty"`
	Entries  []ChangeEntry `json:"entries"`
}
