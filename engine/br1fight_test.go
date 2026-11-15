package engine

import (
	"testing"

	"gobox/rules"
)

// TestBR1Fight: a script's fight on the shared engine with the XXVc rules (buckrogers-go TestFight). Block 16
// runs ClearMonsters, LoadMonster(41, 1, 32), Combat. So far: the monster is loaded through the ruleset
// (ClearLoot, NewGroup, Spawn, SetOrder, BeastKindFact). What stops it: Combat's set-up asks
// rules.BuildGround, and rulesxxvc holds no board (it takes Fight.Board per call, the shared engine passes
// none and expects the ruleset to lay ground, place the units and answer UnitsFact/GroundFact): see the
// hand-back of branch fold/f1.
func TestBR1Fight(t *testing.T) {
	e := br1Begun(t)
	members := e.partySize()
	br1OwnBlock(t, e, 201, "    ClearMonsters()\n    LoadMonster(41, count=2, icon=32)")
	ids := e.partyIDs()
	if len(ids) != members+2 {
		t.Fatalf("party of %d, want %d members and 2 monsters", len(ids), members)
	}
	for _, id := range ids[members:] {
		if e.ruleset().Sheet(id).Name == "" || e.fact(id, rules.FactSide) != rules.SideMonsters {
			t.Errorf("monster %d: %+v", id, e.ruleset().Sheet(id))
		}
	}
	t.Skip("Combat: rulesxxvc has no BuildGround, PlaceFight or board state (REFACTOR.md B1f fight)")
}
