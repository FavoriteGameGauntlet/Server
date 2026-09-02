package typesysparams

type SystemParameter struct {
	Id          int
	Code        string
	Name        string
	Description string
	Value       string
}

type DefaultSystemParameter struct {
	Id           int
	Code         string
	DefaultValue string
	Name         string
	Description  string
}

const (
	ParamMinimumNumberOfWishlistGames = "MinimumNumberOfWishlistGames"

	ParamTimerDurationInS                  = "TimerDurationInS"
	ParamTimerFinisherSchedulerIntervalInS = "TimerFinisherSchedulerIntervalInS"
	ParamEffectFinisherSchedulerIntervalInS = "EffectFinisherSchedulerIntervalInS"
	ParamMaximumAvailableRollCountForTimer = "MaximumAvailableRollCountForTimer"

	ParamAvailableRollChangeByTimer       = "AvailableRollChangeByTimer"
	ParamAvailableRollChangeByRoll        = "AvailableRollChangeByRoll"
	ParamTerritoryHourChangeByTimer       = "TerritoryHourChangeByTimer"
	ParamExperiencePointChangeByTimer     = "ExperiencePointChangeByTimer"
	ParamExperiencePointByLevelUp         = "ExperiencePointChangeByLevelUp"
	ParamTerritoryHourChangeBySeizeSlice  = "TerritoryHourChangeBySeizeSlice"
	ParamTerritoryPointChangeBySeizeSlice = "TerritoryPointChangeBySeizeSlice"
	ParamFreePointChangeBySandstorm       = "FreePointChangeBySandstorm"
	ParamFreePointChangeByBaseTeleport    = "FreePointChangeByBaseTeleport"
	ParamFreePointsMinimum                = "FreePointsMinimum"
	ParamShouldLimitFreePoints            = "ShouldLimitFreePoints"
	ParamSeizePenaltyPoints               = "SeizePenaltyPoints"

	ParamMinimumAvailableRollCountForRoll    = "MinimumAvailableRollCountForRoll"
	ParamMinimumAvailableWheelEffectsForRoll = "MinimumAvailableWheelEffectsForRoll"
)
