package engine

import (
	"fmt"
	"gobox/rules"
)

// takeKey is the key pressed for the next computer turn, taken.
func (e *Engine) takeKey() rules.AutoKey {
	k := e.aiKey
	e.aiKey = rules.AutoKeyNone
	return k
}

// maxComputerSteps bounds one computer turn's Decide/Resolve rounds: a ruleset that never ends a turn is a bug.
const maxComputerSteps = 1000

// computerTurn plays c's turn as the ruleset decides it, the key pressed for it carried in the action. A ruleset
// that answers AutoTurn plays the whole turn in one Resolve (por); one that decides action by action (XXVc's
// CombatAiControl) is asked again until an action uses the turn up (Outcome.TurnOver); there the key pressed is
// the typed rules.TakeBack, resolved once before the turn's first action (CheckKeystroke), and a member handed
// back ends the turn.
func (e *Engine) computerTurn(cID rules.CharID) {
	key := e.takeKey()
	for range maxComputerSteps {
		f := e.fightBy(cID)
		a := e.ruleset().Decide(f, f.Actor, nil)
		t, auto := a.(rules.AutoTurn)
		if auto {
			t.Key = key
			a = t
		} else if key == rules.AutoKeyTakeOver {
			key = rules.AutoKeyNone
			back, err := e.ruleset().Resolve(f, rules.TakeBack{}, nil)
			if err != nil {
				panic(fmt.Sprintf("computer turn: take back: %v", err))
			}
			e.present(back.Effects...)
			if back.TurnOver {
				return
			}
		}
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
