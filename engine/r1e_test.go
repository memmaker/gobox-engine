package engine

import (
	"gobox/adventure"
	"gobox/rules"
	"gobox/rules1e"
)

// The tests cannot import por (the engine's module does not require the
// game): they build the rules with char.Rules1e and read the portrait tables
// the game's import wrote (ReadGameTables). Without the game files there are
// none, and the tests that need them skip (testEngine).
var porTables = func() adventure.GameTables {
	t, _ := adventure.ReadGameTables("../data/por/sources/assets")
	return t
}()

// A test's bare &Engine{} runs on POR's ruleset and tables.
func init() { bareRuleset, bareData = rules1e.Rules1e, porTables }

// The tests' window into the rules' state: the Engine itself reaches the
// rules only through core.Rules (fold E-R6).

// r1e is the ruleset as rules1e, for what core.Rules does not carry yet (a
// character by id, the save's view): tests only.
func (e *Engine) r1e() *rules1e.Rules { return e.ruleset().(*rules1e.Rules) }

// charID is c's id in the ruleset, until the Engine holds ids.
func (e *Engine) charID(c *rules1e.Character) rules.CharID { return e.r1e().ID(c) }

// state is the rules' state, owned by the ruleset (the Pascal's gdvars,
// rulestate.go). The fold's next steps replace these reads by core.Rules calls.
func (e *Engine) state() *rules1e.RuleState { return e.r1e().S }

// member is the character of id (goes when the Engine holds ids).
func (e *Engine) member(id rules.CharID) *rules1e.Character { return e.r1e().Char(id) }
