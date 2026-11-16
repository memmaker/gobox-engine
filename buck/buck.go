// Package buck is the host of the Buck Rogers games (BR1, BR2) on the shared
// engine: the one place that names their ruleset, rulesxxvc. Their player
// has no rule options to hand in (as pool.Settings does for rules1e).
package buck

import (
	"gobox/adventure"
	"gobox/engine"
	"gobox/rulesxxvc"
)

// New is the Engine of a rulesxxvc game on the adventure folder adv, with
// the tables its import wrote, and the host data directory data.
func New(adv, data string) (*engine.Engine, error) {
	tables, err := adventure.ReadGameTables(adv)
	if err != nil {
		return nil, err
	}
	return engine.New(rulesxxvc.Builder, tables, adv, data)
}
