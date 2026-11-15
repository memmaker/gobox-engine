package engine

import (
	"fmt"
	"gobox/adventure"
	"gobox/rules"
	"os"
	"path/filepath"
	"slices"
)

// monsterSet is the set the monsters' MON and CPIC archives are read from.
func (e *Engine) monsterSet() rules.Set {
	if s := e.game().MonsterSet; s != 0 {
		return s
	}
	return e.set
}

// artNames is the game's ArtNames.
func (e *Engine) artNames() adventure.ArtNames {
	if a := e.game().Art; a != nil {
		return *a
	}
	return adventure.ArtNames{}
}

// bareRuleset and bareData are what an Engine made without New (a test's
// &Engine{}) builds from; the tests set them (r1e_test.go).
var (
	bareRuleset rules.Builder
	bareData    adventure.GameTables
)

// ruleset is the Engine's ruleset, owner of the rules' state (fold step E-R6), made
// at first use by the game's Ruleset: the Engine reaches the rules through it.
func (e *Engine) ruleset() rules.Rules {
	if e.rules == nil {
		build := e.newRules
		if build == nil {
			build = bareRuleset
		}
		e.rules, e.create = build(source{e}, creationHost{e})
	}
	return e.rules
}

// game is the game's GameTables (bareData for an Engine made without New).
func (e *Engine) game() adventure.GameTables {
	if e.given == nil {
		return bareData
	}
	return *e.given
}

// selFact is the selected member's Fact f as a number; a Fact the ruleset
// does not hold is a bug of the script table.
func (e *Engine) selFact(f rules.Fact) int64 {
	return int64(e.selValue(f).(rules.Number))
}

// setSelFact writes the selected member's Fact f.
func (e *Engine) setSelFact(f rules.Fact, v int64) {
	e.setSelValue(f, rules.Number(v))
}

// selValue is the selected member's Fact f.
func (e *Engine) selValue(f rules.Fact) rules.Value {
	v, err := e.ruleset().Fact(e.selected(), f)
	if err != nil {
		panic(fmt.Sprintf("script: fact %v: %v", f, err))
	}
	return v
}

// setSelValue writes the selected member's Fact f.
func (e *Engine) setSelValue(f rules.Fact, v rules.Value) {
	if err := e.ruleset().SetFact(e.selected(), f, v); err != nil {
		panic(fmt.Sprintf("script: fact %v: %v", f, err))
	}
}

// selIndex is the selected member's typed Fact f as the script's number: its
// place in the record's table of f's values. A record byte outside the table
// comes back as a core.Number and is read as it stands (SpecialReadZ1).
func selIndex[V comparable](e *Engine, f rules.Fact, table []V) int {
	switch v := e.selValue(f).(type) {
	case rules.Number:
		return int(v)
	case V:
		return slices.Index(table, v)
	}
	panic(fmt.Sprintf("script: fact %v: unexpected value", f))
}

// worldWord is the world word w's value (core.WordIs); a word the rules do not
// keep is a bug of the Engine.
func (e *Engine) worldWord(w rules.WorldWord) int {
	a := e.ruleset().Check(rules.WordIs{Word: w}, nil)
	if !a.Yes {
		panic(fmt.Sprintf("rules keep no world word %d", w))
	}
	return int(a.Value)
}

// setWorldWord stores v in the world word w (core.SetWord).
func (e *Engine) setWorldWord(w rules.WorldWord, v int) {
	e.do(rules.SetWord{Word: w, Value: int32(v)})
}

// rollDice is n dice of sides from the rules' generator (core.RollDice).
func (e *Engine) rollDice(n, sides int) int {
	return int(e.ruleset().Check(rules.RollDice{N: int32(n), Sides: int32(sides)}, nil).Value)
}

// rollRandom is the rules' Random(n) (core.RollRandom).
func (e *Engine) rollRandom(n int) int {
	return int(e.ruleset().Check(rules.RollRandom{N: int32(n)}, nil).Value)
}

// rulesFile is the adventure's rules.json (tablespec.RulesFile in its data folder), "" when it has none.
func (e *Engine) rulesFile() string {
	p := filepath.Join(e.adv, adventure.DataDir, adventure.RulesFile)
	if _, err := os.Stat(p); err != nil {
		return ""
	}
	return p
}
