package rules

// ClockPart names a part of the area clock a script or the engine reads or
// writes (the 1e rules keep it: Area.Clock).
type ClockPart int

const (
	ClockTenMinutes  ClockPart = iota + 1 // the tens of minutes (por area.Clock[2])
	ClockHours                            // the hours (por area.Clock[3])
	ClockMinuteOfDay                      // the minutes since midnight (read only)
)

// ClockIs asks the clock's Part (Answer.Value).
type ClockIs struct{ Part ClockPart }

// SetClock stores Value in the clock's Part (not ClockMinuteOfDay: read only).
type SetClock struct {
	Part  ClockPart
	Value int32
}

func (ClockIs) check()   {}
func (SetClock) action() {}
