package engine

import (
	"fmt"
	"gobox/rules"
)

// takeKey is the key pressed for the next computer turn (nil: none), taken.
func (e *Engine) takeKey() rules.Action {
	k := e.aiKey
	e.aiKey = nil
	return k
}

// maxComputerSteps bounds one computer turn's Decide/Resolve rounds: a ruleset that never ends a turn is a bug.
const maxComputerSteps = 1000

// computerTurn plays c's turn as the ruleset decides it. The key pressed for it is a typed action (rules.TakeBack,
// rules.QuickMagic), resolved once before the turn's first action for every ruleset (CheckKeystroke): XXVc takes the
// members back there; por's rules remember it for the polls inside its whole turn. A ruleset that answers AutoTurn
// plays the turn in one Resolve; one that decides action by action (XXVc's CombatAiControl) is asked again until an
// action uses the turn up (Outcome.TurnOver); a member handed back ends the turn.
func (e *Engine) computerTurn(cID rules.CharID) {
	key := e.takeKey()
	for range maxComputerSteps {
		f := e.fightBy(cID)
		if key != nil {
			pressed := key
			key = nil
			back, err := e.ruleset().Resolve(f, pressed, nil)
			if err != nil {
				panic(fmt.Sprintf("computer turn: key %T: %v", pressed, err))
			}
			e.present(back.Effects...)
			if back.TurnOver {
				return
			}
		}
		a := e.ruleset().Decide(f, f.Actor, nil)
		_, auto := a.(rules.AutoTurn)
		out, err := e.ruleset().Resolve(f, a, nil)
		if err != nil {
			panic(fmt.Sprintf("computer turn: %v", err))
		}
		e.present(out.Effects...)
		if auto || out.TurnOver {
			return
		}
	}
	panic(fmt.Sprintf("computer turn of %d: no action used the turn up in %d steps", cID, maxComputerSteps))
}
