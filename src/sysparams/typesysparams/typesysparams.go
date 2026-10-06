package typesysparams

type SystemParameter struct {
	Id          int
	Code        string
	Name        string
	Description string
	Value       string
	IsDefault   bool
}

type DefaultSystemParameter struct {
	Id           int
	Code         string
	DefaultValue string
	Name         string
	Description  string
}

const (
	ParamMinimumNumberOfWishlistGames        = "MinimumNumberOfWishlistGames"
	ParamTimerDurationInS                    = "TimerDurationInS"
	ParamMinimumAvailableWheelEffectsForRoll = "MinimumAvailableWheelEffectsForRoll"
)
