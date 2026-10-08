package common

import "time"

const (
	SessionCookieName    = "session_id"
	ConstraintZeroOrLess = "zero or less"
	ConstraintZeroOrMore = "zero or more"
)

// SchedulerInterval is how often the background schedulers (finishing timers, clearing ended
// effects) look for work.
const SchedulerInterval = time.Second
