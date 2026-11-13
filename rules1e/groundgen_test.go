package rules1e

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestGroundGenJSON: rules.json combatGround carries every field of a GroundGen,
// and POR's terrain table is TerrainKinds' ranges (first match) per number.
func TestGroundGenJSON(t *testing.T) {
	b, err := json.Marshal(PorGroundGen)
	if err != nil {
		t.Fatal(err)
	}
	var back GroundGen
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(back, PorGroundGen) {
		t.Errorf("GroundGen changed in rules.json:\n%s", b)
	}
	if len(POR.Terrain) != 256 || POR.Terrain[1] != 1 || POR.Terrain[0xB2] != 0x60 {
		t.Errorf("POR.Terrain: %v", POR.Terrain[:4])
	}
}
