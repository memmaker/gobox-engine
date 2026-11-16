package engine

import (
	"fmt"
	"gobox/rules"
	"hash/fnv"
	"image"
	"testing"
)

// TestClutterDice pins the combat ground's dice: the same seed lays the same
// map and clutter as the engine's dungeonMap/wildMap did before the
// BuildGround port (hashes taken from por-go 72e4b64^ with the same setup).
func TestClutterDice(t *testing.T) {
	for _, c := range []struct {
		viewMode int
		areaKind byte
		want     uint64
	}{
		{1, 0, 0x4d7f6ef238dec47b},
		{1, 0xA, 0xa8e50419f974917f},
		{2, 0, 0x5fde4711f45315fc},
		{2, 0xA, 0x337953e4003feeab},
		{3, 0, 0x5fde4711f45315fc},
		{4, 0xA, 0x337953e4003feeab},
	} {
		tables := porTables
		tables.NoRemainsBlocks = []int{0xA} // POR's (combatst.pas: eclScriptId <> 10), whatever the installed folder holds
		e := &Engine{given: &tables}
		e.initVM()
		for i := range 0x200 {
			e.gs.geo.Plane[i] = byte(i*7) & 0x33
		}
		e.posX, e.posY = 8, 8
		e.viewMode = c.viewMode
		e.gs.areaKind = c.areaKind
		e.ruleset().BeginFight(rules.Fight{Terms: rules.FightTerms{PartyAt: image.Pt(8, 8)}}, nil)
		e.state().RandSeed = 12345
		if _, err := e.ruleset().Resolve(rules.Fight{}, e.groundTerms(), nil); err != nil {
			t.Fatal(err)
		}
		h := fnv.New64a()
		fmt.Fprint(h, e.state().Cmap.Ground)
		if got := h.Sum64(); got != c.want {
			t.Errorf("viewMode %d areaKind %#x: ground %#x, want %#x", c.viewMode, c.areaKind, got, c.want)
		}
	}
}
