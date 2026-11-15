package rules1e

import (
	"fmt"
	"gobox/rules"
	"path/filepath"
)

// Rules1e builds the first engine's ruleset, rules1e, on the Engine's source
// (rules.Builder): its monster records, base item table and treasure blocks
// are read from src; host is the Engine's side of the creation. The one
// builder of rules1e for the Engine: the host and tests hand it to
// engine.New. The adventure's rules.json ("": it has none, as POR) holds
// the tables the game plays on, laid out by the layout.json
// (rules1e.LayoutFile) beside it, which it must then carry too (a missing one
// panics: run the game's import); else on rules1e.POR.
func Rules1e(src rules.Source, host rules.CreationHost) (rules.Rules, rules.CreationRules) {
	r := &Rules{S: &RuleState{}}
	t := newTables(src)
	if rulesFile := src.RulesFile(); rulesFile != "" {
		l, err := ReadLayout(filepath.Join(filepath.Dir(rulesFile), LayoutFile))
		if err != nil {
			panic(fmt.Sprintf("rules1e: %s needs its layout: %v", rulesFile, err))
		}
		tabs, err := POR.LoadRules(rulesFile, l)
		if err != nil {
			panic(fmt.Sprintf("rules1e: %v", err))
		}
		r.S.Tables = tabs
	}
	r.Monsters = t.monster
	r.FirstMonster = t.firstMonster
	r.ItemTable = t.itemTable
	r.TreasureBlock = t.itemBlock
	r.ItemRecords = func(items []Item) rules.ItemRecords { return WriteItems(items) }
	r.Host = host
	return r, r.Creation()
}
