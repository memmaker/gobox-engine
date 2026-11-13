// Package porgame makes POR's Engine: the engine over POR's ruleset
// (rules1e, built by char.Rules1e) and POR's tables (the portrait tables
// its import writes).
package porgame

import (
	"gobox/engine"
	"gobox/rules1e"
)

// New is POR's Engine on the adventure folder adv and the host data directory data (pool.New).
func New(adv, data string) (*engine.Engine, error) {
	tables, err := engine.ReadGameTables(adv)
	if err != nil {
		return nil, err
	}
	return engine.New(rules1e.Rules1e, tables, adv, data)
}

// OpenSave is pool.OpenSave with POR's ruleset.
func OpenSave(dir, adv string) (*engine.SaveEditor, error) {
	return engine.OpenSave(rules1e.Rules1e, dir, adv)
}
