package engine

import (
	"testing"

	"gobox/rules"
)

// TestBR1ShipCells: a script reads and writes the party's ship and the last foe's summary by name
// (Ship*, Foe*; the text's D_ form addresses the fuel word through a 32-bit operand, which the text form of a
// named variable does not give a word); the ruleset's ship and foe summary hold the numbers, the engine keeps none.
func TestBR1ShipCells(t *testing.T) {
	e := br1Begun(t)
	cell := func(f rules.ShipCellFact) int64 { return int64(e.fact(rules.NoChar, f).(rules.Number)) }
	fuel := cell(rules.ShipCellFact{Part: rules.ShipSystemStatus, Index: 4})
	if fuel == 0 {
		t.Fatal("the party's ship starts without fuel")
	}
	br1OwnBlock(t, e, 201, "    D_4CE2 = D_4D1E\n    D_4D1E = 321\n    ShipMount0Ammo = 5\n    FoeCrew = 9")
	if got := br1Cell(t, e, "Word4CE2"); got != fuel {
		t.Errorf("script read fuel %d, ruleset has %d", got, fuel)
	}
	for _, c := range []struct {
		name string
		f    rules.ShipCellFact
		want int64
	}{
		{"fuel", rules.ShipCellFact{Part: rules.ShipSystemStatus, Index: 4}, 321},
		{"mount 0 ammo", rules.ShipCellFact{Part: rules.ShipMountAmmo}, 5},
		{"foe crew", rules.ShipCellFact{Foe: true, Part: rules.ShipCrew}, 9},
	} {
		if got := cell(c.f); got != c.want {
			t.Errorf("%s: ruleset has %d, want %d", c.name, got, c.want)
		}
	}
}
