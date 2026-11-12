package char

import (
	"gobox/core"
	"gobox/core/goldbox"
	"rules1e"
)

// Rules1e builds the first engine's ruleset, rules1e, from the Engine's
// sources (pool.Ruleset): monster, items and treasure are the game data's
// monster records, base item table and treasure blocks; host is the
// Engine's side of the creation. The one builder of rules1e for the Engine:
// games and tests hand it to pool.New.
func Rules1e(
	monster func(core.MonsterID) (Record285, []AffectRecord, []Item, bool),
	items func() []ItemBase,
	treasure func(block int) []Item,
	host goldbox.CreationHost,
) (core.Rules, core.CreationRules) {
	r := &rules1e.Rules{S: &rules1e.RuleState{}}
	r.Monsters = func(id core.MonsterID) (rules1e.Monster, bool) {
		rec, affects, its, ok := monster(id)
		return Monster285(rec, affects, its), ok
	}
	r.ItemTable = items
	r.TreasureBlock = treasure
	r.Host = host
	return r, r.Creation()
}
