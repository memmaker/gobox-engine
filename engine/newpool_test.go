package engine

import (
	"gobox/adventure"
	"gobox/rules1e"
)

// NewPool is the Engine of a rules1e game on the adventure folder adv and the
// host data directory data (pool.New, which the engine's tests cannot import).
func NewPool(adv, data string) (*Engine, error) {
	tables, err := adventure.ReadGameTables(adv)
	if err != nil {
		return nil, err
	}
	return New(rules1e.Rules1e, tables, adv, data)
}
