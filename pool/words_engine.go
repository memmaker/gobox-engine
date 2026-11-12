package pool

import (
	"fmt"
	"gobox/core"
	"gobox/ecl/script"
)

// namedWord is the loaded script's layout field called name; a game whose data names a
// field its layout lacks is a bug of the game's data.
func (e *Engine) namedWord(name string) script.Ref {
	l := e.vm.Script.Layout
	if l == nil {
		panic(fmt.Sprintf("gametables.json Words names %s: the script has no layout", name))
	}
	r, err := l.Named(name)
	if err != nil {
		panic(fmt.Sprintf("gametables.json Words: %v", err))
	}
	return r
}

// bindWords points DIVIDE's remainder at the game's Remainder field (GameTables.Words),
// for the script just loaded.
func (e *Engine) bindWords() {
	e.vm.Remainder = script.Ref{}
	if n := e.game().Words.Remainder; n != "" {
		e.vm.Remainder = e.namedWord(n)
	}
}

// setLastBlock keeps block, the one NewECL leaves, to come back to: in the game's LastBlock
// field, else in the ruleset's core.WordLastBlock.
func (e *Engine) setLastBlock(block int) {
	if n := e.game().Words.LastBlock; n != "" {
		e.host.Set(e.namedWord(n), int64(block))
		return
	}
	e.setWorldWord(core.WordLastBlock, block)
}
