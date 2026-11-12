package pool

import (
	"gobox/char"
	"gobox/core"
	"gobox/core/goldbox"
	"rules1e"
)

// The tests cannot import por (the engine's module does not require the
// game): porRuleset and porTables are copies of porrules.Ruleset and of
// presentdata's portrait tables, test-only. Keep them equal to those.
var testPortraitHead = [42]goldbox.PortraitPic{{}, {Set: 3, Id: 0x00}, {Set: 3, Id: 0x08}, {Set: 3, Id: 0x09}, {Set: 3, Id: 0x0D}, {Set: 3, Id: 0x10}, {Set: 3, Id: 0x12}, {Set: 3, Id: 0x16}, {Set: 3, Id: 0x22}, {Set: 3, Id: 0x2D}, {Set: 3, Id: 0x33}, {Set: 3, Id: 0x35}, {Set: 3, Id: 0x39}, {Set: 3, Id: 0x43}, {Set: 3, Id: 0x44}, {Set: 3, Id: 0x18}, {Set: 3, Id: 0x20}, {Set: 3, Id: 0x2A}, {Set: 3, Id: 0x2C}, {Set: 3, Id: 0x37}, {Set: 3, Id: 0x3F}, {Set: 3, Id: 0x47}, {Set: 1, Id: 0x06}, {Set: 1, Id: 0x30}, {Set: 2, Id: 0x05}, {Set: 2, Id: 0x0B}, {Set: 2, Id: 0x19}, {Set: 2, Id: 0x34}, {Set: 5, Id: 0x0C}, {Set: 5, Id: 0x0E}, {Set: 5, Id: 0x36}, {Set: 5, Id: 0x3B}, {Set: 5, Id: 0x46}, {Set: 5, Id: 0x48}, {Set: 6, Id: 0x01}, {Set: 6, Id: 0x03}, {Set: 6, Id: 0x0A}, {Set: 7, Id: 0x38}, {Set: 7, Id: 0x3D}, {Set: 7, Id: 0x3E}, {Set: 7, Id: 0x4A}, {Set: 8, Id: 0x45}}
var testPortraitBody = [22]goldbox.PortraitPic{{}, {Set: 3, Id: 0x01}, {Set: 3, Id: 0x02}, {Set: 3, Id: 0x03}, {Set: 3, Id: 0x04}, {Set: 3, Id: 0x07}, {Set: 3, Id: 0x08}, {Set: 3, Id: 0x12}, {Set: 3, Id: 0x18}, {Set: 3, Id: 0x1A}, {Set: 3, Id: 0x21}, {Set: 3, Id: 0x23}, {Set: 3, Id: 0x25}, {Set: 3, Id: 0x09}, {Set: 3, Id: 0x1F}, {Set: 3, Id: 0x20}, {Set: 3, Id: 0x22}, {Set: 1, Id: 0x19}, {Set: 2, Id: 0x05}, {Set: 2, Id: 0x0D}, {Set: 5, Id: 0x1E}, {Set: 6, Id: 0x06}}

var porTables = GameTables{PortraitHead: testPortraitHead[:], PortraitBody: testPortraitBody[:]}

func porRuleset(
	monster func(core.MonsterID) (char.Record285, []char.AffectRecord, []char.Item, bool),
	items func() []char.ItemBase,
	treasure func(block int) []char.Item,
	host goldbox.CreationHost,
) (core.Rules, core.CreationRules) {
	r := &rules1e.Rules{S: &rules1e.RuleState{}}
	r.Monsters = func(id core.MonsterID) (rules1e.Monster, bool) {
		rec, affects, its, ok := monster(id)
		return char.Monster285(rec, affects, its), ok
	}
	r.ItemTable = items
	r.TreasureBlock = treasure
	r.Host = host
	return r, r.Creation()
}

// A test's bare &Engine{} runs on POR's ruleset and tables.
func init() { bareRuleset, bareData = porRuleset, porTables }

// The tests' window into the rules' state: the Engine itself reaches the
// rules only through core.Rules (fold E-R6).

// r1e is the ruleset as rules1e, for what core.Rules does not carry yet (a
// character by id, the save's view): tests only.
func (e *Engine) r1e() *rules1e.Rules { return e.ruleset().(*rules1e.Rules) }

// charID is c's id in the ruleset, until the Engine holds ids.
func (e *Engine) charID(c *rules1e.Character) core.CharID { return e.r1e().ID(c) }

// state is the rules' state, owned by the ruleset (the Pascal's gdvars,
// rulestate.go). The fold's next steps replace these reads by core.Rules calls.
func (e *Engine) state() *rules1e.RuleState { return e.r1e().S }

// member is the character of id (goes when the Engine holds ids).
func (e *Engine) member(id core.CharID) *rules1e.Character { return e.r1e().Char(id) }
