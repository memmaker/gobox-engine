package char

import (
	"fmt"
	"gobox/rules"
	"gobox/rules1e"
	"path/filepath"
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
	rulesFile string,
	monster func(rules.MonsterID) (Sheet, []AffectRecord, []Item, bool),
	items func() []ItemBase,
	treasure func(block int) []Item,
	host rules.CreationHost,
) (rules.Rules, rules.CreationRules) {
	r := &rules1e.Rules{S: &rules1e.RuleState{}}
	if rulesFile != "" {
		l, err := rules1e.ReadLayout(filepath.Join(filepath.Dir(rulesFile), rules1e.LayoutFile))
		if err != nil {
			panic(fmt.Sprintf("rules1e: %s needs its layout: %v", rulesFile, err))
		}
		t, err := rules1e.POR.LoadRules(rulesFile, l)
		if err != nil {
			panic(fmt.Sprintf("rules1e: %v", err))
		}
		r.S.Tables = t
	}
	r.Monsters = func(id rules.MonsterID) (rules1e.Monster, bool) {
		rec, affects, its, ok := monster(id)
		return rules1e.Monster{Sheet: rec, Affects: affects, Items: its}, ok
	}
	r.ItemTable = items
	r.TreasureBlock = treasure
	r.ItemRecords = func(items []Item) rules.ItemRecords { return WriteItems(items) }
	r.Host = host
	return r, r.Creation()
}
