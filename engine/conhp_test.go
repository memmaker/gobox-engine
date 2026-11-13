package engine

import (
	"gobox/rules1e"
	"testing"
)

func TestConAdjClamp(t *testing.T) {
	c := &rules1e.Character{}
	c.Class = rules1e.ClsFighter
	if rules1e.ConAdj(c, 2) != -2 || rules1e.ConAdj(c, 0) != -2 || rules1e.ConAdj(c, 25) != rules1e.ConAdj(c, 20) {
		t.Errorf("constitution out of the table: %d %d %d", rules1e.ConAdj(c, 2), rules1e.ConAdj(c, 0), rules1e.ConAdj(c, 25))
	}
}

// Genheeris, a level 7 magic-user past the slot table, keeps his record's
// slots when the joining recomputes him.
func TestGenheerisSlots(t *testing.T) {
	e := &Engine{}
	e.initVM()
	c := &rules1e.Character{}
	c.Levels[rules1e.ClsMage], c.SpellSlots = 7, rules1e.SpellSlots{rules1e.SpellClassMage: {4, 2, 2}}
	e.state().RecomputeCharacter(c)
	if c.SpellSlots != (rules1e.SpellSlots{rules1e.SpellClassMage: {4, 2, 2}}) {
		t.Errorf("slots %v", c.SpellSlots)
	}
}

// A magic-user drained from level 2 to 1 memorizes one spell again.
func TestDrainedSlots(t *testing.T) {
	e := &Engine{}
	e.initVM()
	c := &rules1e.Character{}
	c.Levels[rules1e.ClsMage], c.SpellSlots = 1, rules1e.SpellSlots{rules1e.SpellClassMage: {2, 0, 0}}
	e.state().RecomputeCharacter(c)
	if c.SpellSlots != (rules1e.SpellSlots{rules1e.SpellClassMage: {1, 0, 0}}) {
		t.Errorf("slots %v", c.SpellSlots)
	}
}
