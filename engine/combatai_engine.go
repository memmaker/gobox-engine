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

// computerTurn plays c's turn as core.Rules decides it, the key pressed
// for it carried in the action.
func (e *Engine) computerTurn(cID rules.CharID) {
	f := rules.Fight{Actor: cID}
	a := e.ruleset().Decide(f, f.Actor, nil)
	if t, ok := a.(rules.AutoTurn); ok {
		t.Key = e.takeKey()
		a = t
	}
	out, err := e.ruleset().Resolve(f, a, nil)
	if err != nil {
		panic(fmt.Sprintf("computer turn: %v", err))
	}
	e.present(out.Effects...)
}
