package engine

import (
	"fmt"
	"gobox/rules"
)

// inFight is whether a fight is on.
func (e *Engine) inFight() bool { return e.ruleset().Check(rules.InFight{}, nil).Yes }

// beginFlag is the game's begin flag (core.WordBeginFlag).
func (e *Engine) beginFlag() int { return int(e.worldWord(rules.WordBeginFlag)) }

// do resolves a rules action outside any situation; these never refuse.
func (e *Engine) do(a rules.Action) {
	if _, err := e.ruleset().Resolve(nil, a, nil); err != nil {
		panic(fmt.Sprintf("rules: %T: %v", a, err))
	}
}
