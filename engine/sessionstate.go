package engine

import (
	"fmt"

	"gobox/rules"
)

// The session state the rules keep and the engine asks by typed kinds (the clock, the party's count),
// and the phase the engine keeps and hands to the rules.

// setPhase is the one place the game's phase changes: the engine's own and the rules' copy.
func (e *Engine) setPhase(p rules.Phase) {
	e.phase = p
	e.do(rules.EnterPhase{Phase: p})
}

// clockPart is the clock's part (a bug of the engine when the rules do not keep it).
func (e *Engine) clockPart(part rules.ClockPart) int {
	a := e.ruleset().Check(rules.ClockIs{Part: part}, nil)
	if !a.Yes {
		panic(fmt.Sprintf("rules keep no clock part %d", part))
	}
	return int(a.Value)
}

// setClockPart stores v in the clock's part.
func (e *Engine) setClockPart(part rules.ClockPart, v int) {
	e.do(rules.SetClock{Part: part, Value: int32(v)})
}

// memberCount is the party's count in the game's record (not partySize: the list's length).
func (e *Engine) memberCount() int {
	return int(e.ruleset().Check(rules.MemberCount{}, nil).Value)
}

// setMemberCount stores the party's count in the game's record.
func (e *Engine) setMemberCount(n int) { e.do(rules.SetMemberCount{N: int32(n)}) }
