package char

import (
	"fmt"
	"gobox/core"
	"gobox/core/goldbox"
	"path/filepath"
	"rules1e"
)

// Rules1e builds the first engine's ruleset, rules1e, from the Engine's
// sources (pool.Ruleset): monster, items and treasure are the game data's
// monster records, base item table and treasure blocks; host is the
// Engine's side of the creation. The one builder of rules1e for the Engine:
// games and tests hand it to pool.New. rules is the adventure's rules.json ("": it
// has none, as POR): the game plays on those tables laid out by the layout.json
// (rules1e.LayoutFile) beside it, which it must then carry too (a missing one
// panics: run the game's import); else on rules1e.POR.
func Rules1e(
	rules string,
	monster func(core.MonsterID) (Record285, []AffectRecord, []Item, bool),
	items func() []ItemBase,
	treasure func(block int) []Item,
	host goldbox.CreationHost,
) (core.Rules, core.CreationRules) {
	r := &rules1e.Rules{S: &rules1e.RuleState{}}
	if rules != "" {
		l, err := rules1e.ReadLayout(filepath.Join(filepath.Dir(rules), rules1e.LayoutFile))
		if err != nil {
			panic(fmt.Sprintf("rules1e: %s needs its layout: %v", rules, err))
		}
		t, err := rules1e.POR.LoadRules(rules, l)
		if err != nil {
			panic(fmt.Sprintf("rules1e: %v", err))
		}
		r.S.Tables = t
	}
	r.Monsters = func(id core.MonsterID) (rules1e.Monster, bool) {
		rec, affects, its, ok := monster(id)
		return Monster285(rec, affects, its), ok
	}
	r.ItemTable = items
	r.TreasureBlock = treasure
	r.ItemRecords = func(items []Item) core.ItemRecords { return WriteItems(items) }
	r.Host = host
	return r, r.Creation()
}
