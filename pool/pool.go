// Package pool is the host of the pool games (POR, COTAB, SotSB, POD): the
// one place that names their ruleset, rules1e, and the combat rules the
// player may choose, rules4e.
package pool

import (
	"gobox/adventure"
	"gobox/engine"
	"gobox/rules"
	"gobox/rules1e"
	"gobox/rules4e"
)

// New is the Engine of a rules1e game (POR, COTAB, SotSB, POD) on the
// adventure folder adv, with the tables its import wrote, and the host data
// directory data.
func New(adv, data string) (*engine.Engine, error) {
	tables, err := adventure.ReadGameTables(adv)
	if err != nil {
		return nil, err
	}
	return engine.New(rules1e.Rules1e, tables, adv, data)
}

// Settings are the player's rule options the frontend hands in (Engine.HandIn).
type Settings struct {
	PathSearch bool // Gameplay "Path search": the computer walks by path search
	Combat4e   bool // Gameplay "4e Combat": the rules4e combat rules
}

// Rules are s as the ruleset's settings.
func (s Settings) Rules() rules.Settings {
	var combat rules1e.CombatRules // nil: the original's
	if s.Combat4e {
		combat = rules4e.New()
	}
	return rules1e.Settings{Combat: combat, PathSearch: s.PathSearch}
}

// ReadRoster is the characters another game of record layout l saved in dir,
// as canonical files (engine.Config.Foreign).
func ReadRoster(l rules.RecordLayout, dir string) ([]rules.CharacterFile, error) {
	return rules1e.ReadRoster(l, dir)
}
