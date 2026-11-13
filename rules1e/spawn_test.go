package rules1e

import (
	"gobox/core"
	"testing"
)

// Spawn copies the group's first monster (lists reversed, no new dice);
// NewGroup makes the next one roll again.
func TestSpawnGroup(t *testing.T) {
	s := &RuleState{RandSeed: 7, Items: make([]ItemBase, 1)}
	r := &Rules{S: s, Monsters: func(id core.MonsterID) (Monster, bool) {
		return Monster{Items: []Item{{}, {}}, Affects: []AffectRecord{{Type: 1}, {Type: 2}}}, id == 3
	}}
	if _, err := r.Spawn(4, nil); err == nil {
		t.Fatal("unknown monster spawned")
	}
	r.Spawn(3, nil)
	seed := s.RandSeed
	r.Spawn(3, nil)
	if s.RandSeed != seed {
		t.Error("a copy rolled dice")
	}
	a, b := s.Party[0], s.Party[1]
	if b.Affects[0].Type != a.Affects[1].Type || b.IconColour != a.IconColour {
		t.Error("copy is not the reversed template")
	}
	r.Resolve(nil, core.NewGroup{}, nil)
	r.Spawn(3, nil)
	if s.RandSeed == seed || len(s.Party) != 3 {
		t.Error("NewGroup did not roll a fresh monster")
	}
}
