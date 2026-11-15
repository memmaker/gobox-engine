package engine

import (
	"fmt"
	"strings"
	"testing"

	"gobox/adventure"
	"gobox/game"

	"gobox/rules"
)

// TestBR1Fight: a script's fight on the shared engine with the XXVc rules (buckrogers-go TestFight). Block 201
// runs ClearMonsters, LoadMonster(41, 2, 32), Combat. The monsters load through the ruleset; Combat asks BuildGround
// (the dungeon arena from the block, the place's traits and the walls) and PlaceFight (the units, sized by their icon
// dimensions, the monsters from the sides SetupMonster named).
func TestBR1Fight(t *testing.T) {
	e := br1Begun(t)
	members := e.partySize()
	br1OwnBlock(t, e, 201, "    ClearMonsters()\n    LoadMonster(41, count=2, icon=32)\n    Combat()")
	ids := e.partyIDs()
	if len(ids) != members+2 {
		t.Fatalf("party of %d, want %d members and 2 monsters", len(ids), members)
	}
	for _, id := range ids[members:] {
		if e.ruleset().Sheet(id).Name == "" || e.fact(id, rules.FactSide) != rules.SideMonsters {
			t.Errorf("monster %d: %+v", id, e.ruleset().Sheet(id))
		}
	}
	g := e.groundTerms() // what the engine hands the ruleset's ground: the block, the place's traits, the wilderness terrain
	kind := byte(br1Cell(t, e, "TravelKind"))
	if g.Block != 201 || g.Place != adventure.BRPlaceTraits(kind) || g.Terrain != rules.TerrainID(br1Cell(t, e, "WildTerrain")) {
		t.Errorf("BuildGround %+v: want block 201, traits of area kind %#x, the WildTerrain cell", g, kind)
	}
	if v := e.ruleset().Verdict(e.fight()); v != rules.FightOn {
		t.Fatalf("verdict %v at the start of the fight", v)
	}
	// What stops it: the first turn asks FightCardFact (the unit's turn record), which rulesxxvc does not answer
	// yet (Stage B of the fold: the facts of the turn loop). Assert the stop, so a later step moves it.
	defer func() {
		if got := fmt.Sprint(recover()); !strings.Contains(got, "rules.FightCardFact") {
			t.Fatalf("the fight stopped elsewhere than at FightCardFact: %s", got)
		}
	}()
	for i := 0; i < 400; i++ {
		if _, ok := e.build().(game.Fighting); ok {
			t.Fatal("Fighting reached: play the fight to its end")
		}
		if e.waiting(PendFight) {
			e.Step()
		} else {
			press(e)
		}
	}
	t.Fatal("no stop in 400 steps")
}
