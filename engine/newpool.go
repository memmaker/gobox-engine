package engine

import (
	"gobox/adventure"
	"gobox/rules1e"
)

// NewPool is the Engine of a rules1e game (POR, COTAB, SotSB, POD) on the
// adventure folder adv, with the tables its import wrote, and the host data
// directory data.
func NewPool(adv, data string) (*Engine, error) {
	tables, err := adventure.ReadGameTables(adv)
	if err != nil {
		return nil, err
	}
	return New(rules1e.Rules1e, tables, adv, data)
}
