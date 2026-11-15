package engine

import (
	"fmt"
	"strings"
	"testing"

	"gobox/rules"
)

// TestBR1Fight: a script's fight on the shared engine with the XXVc rules (buckrogers-go TestFight). Block 16
// runs ClearMonsters, LoadMonster(41, 1, 32), Combat. So far: the monster is loaded through the ruleset
// (ClearLoot, NewGroup, Spawn, SetOrder, BeastKindFact); the ruleset holds the board (rulesxxvc board seam) and
// lays a wilderness ground and places the units (BuildGround, PlaceFight). What stops it: the script's area is a
// 3D one, so Combat asks BuildGround{Dungeon}, and BuildDungeonArena needs the area kind (zone0 areaKind: the floor
// 7 or 45, bits $40 and $08), the script id (its random seed) and the barrier plane, none of which the engine
// hands the rules (FightTerms.Area is filled by nobody; BuildGround has no mode). The test asserts that stop.
func TestBR1Fight(t *testing.T) {
	e := br1Begun(t)
	members := e.partySize()
	func() {
		defer func() {
			got := fmt.Sprint(recover())
			if !strings.Contains(got, "BuildGround dungeon: rulesxxvc: not ported yet") {
				t.Fatalf("Combat stopped elsewhere than at the dungeon ground: %s", got)
			}
		}()
		br1OwnBlock(t, e, 201, "    ClearMonsters()\n    LoadMonster(41, count=2, icon=32)\n    Combat()")
		t.Fatal("Combat went through: TestBR1Fight must now play the fight")
	}()
	ids := e.partyIDs()
	if len(ids) != members+2 {
		t.Fatalf("party of %d, want %d members and 2 monsters", len(ids), members)
	}
	for _, id := range ids[members:] {
		if e.ruleset().Sheet(id).Name == "" || e.fact(id, rules.FactSide) != rules.SideMonsters {
			t.Errorf("monster %d: %+v", id, e.ruleset().Sheet(id))
		}
	}
}
