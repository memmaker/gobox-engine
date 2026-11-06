package present

import (
	"strings"
	"testing"

	"gobox/front"
	"gobox/gpu"
)

// TestFightRowWeaponAndStatus: the unit's row shows its weapon and its
// status both (BR2 always states one: OKAY).
func TestFightRowWeaponAndStatus(t *testing.T) {
	dst := gpu.NewImage(1280, 800)
	defer dst.Deallocate()
	var drawn strings.Builder
	g := func(c byte) [8]byte { drawn.WriteByte(c); return [8]byte{} }
	b := &front.Battle{Turn: &front.Turn{Unit: 1}, Units: []front.Unit{{ID: 1, Name: "GIZMO", HP: 43, AC: 2, Weapon: "BOLT GUN", Status: "OKAY", Party: true}}}
	(&Presenter{}).drawFightRows(dst, Note{Glyph: g}, b, 1)
	if s := drawn.String(); !strings.Contains(s, "BOLT GUN") || !strings.Contains(s, "OKAY") || strings.Index(s, "BOLT GUN") > strings.Index(s, "OKAY") {
		t.Fatalf("drawn %q", s)
	}
}
