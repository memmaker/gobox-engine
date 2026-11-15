package rules1e

import (
	"errors"

	"gobox/rules"
)

// The state of the 1e rules' own session that the engine asks and writes by typed
// kinds: the clock, the party's count, the member the rules act for, the affect
// table, the phase the engine handed over and the hall opened for everyone.

var errClockReadOnly = errors.New("rules1e: the minute of the day is read only")

// clockWord is the area clock's word of part (nil: not a stored word).
func (r *Rules) clockWord(part rules.ClockPart) *uint16 {
	switch part {
	case rules.ClockTenMinutes:
		return &r.S.Area.Clock[2]
	case rules.ClockHours:
		return &r.S.Area.Clock[3]
	}
	return nil
}

// sessionCheck answers ClockIs, MemberCount, ActingFor and PartyEffect.
func (r *Rules) sessionCheck(k rules.Check) rules.Answer {
	switch k := k.(type) {
	case rules.ClockIs:
		if k.Part == rules.ClockMinuteOfDay {
			return rules.Answer{Yes: true, Value: int32(r.S.ClockMinutes())}
		}
		if p := r.clockWord(k.Part); p != nil {
			return rules.Answer{Yes: true, Value: int32(*p)}
		}
	case rules.MemberCount:
		return rules.Answer{Yes: true, Value: int32(r.S.PartyCount())}
	case rules.ActingFor:
		return rules.Answer{Yes: true, Value: int32(r.S.Who)}
	case rules.PartyEffect:
		v, ok := r.S.EffectTable(int(k.Index))
		return rules.Answer{Yes: ok, Value: int32(v)}
	}
	return rules.Answer{}
}

// sessionResolve resolves SetClock, SetMemberCount, ActFor, EnterPhase and
// OpenHallForAll; a clock word keeps the value's low bits.
func (r *Rules) sessionResolve(a rules.Action) error {
	switch a := a.(type) {
	case rules.SetClock:
		p := r.clockWord(a.Part)
		if p == nil {
			return errClockReadOnly
		}
		*p = uint16(a.Value)
	case rules.SetMemberCount:
		r.S.Game.PartyCount = uint16(a.N)
	case rules.ActFor:
		r.S.Who = int(a.Member)
	case rules.EnterPhase:
		r.S.Phase = a.Phase
	case rules.OpenHallForAll:
		r.S.Training = true
	}
	return nil
}
