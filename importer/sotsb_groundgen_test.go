package importer

import (
	"encoding/json"
	"testing"
)

// TestCombatGroundData pins the data written to rules.json combatGround
// (the game's own engine it was once compared with is gone).
func TestCombatGroundData(t *testing.T) {
	got, err := json.Marshal(SotsbCombatGround())
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"fill":0,"corners":{"nwDoorOpen":13,"neWallOpen":7,"neDoorOpen":1},"clutter":null,"rubble":{"cellFlag":64,"tile":22,"heap":26,"edge":27,"heapOdds":5,"edgeOdds":9},"wild":null}`
	if string(got) != want {
		t.Fatalf("combatGround is\n%s\nwant\n%s", got, want)
	}
}
