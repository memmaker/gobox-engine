package ui

import (
	"strings"
	"testing"

	"gobox/gpu"
	"gobox/ui/combat"
)

// TestFightRowWeaponAndStatus: the unit's row shows its weapon and its
// status both, status first (BR2 O099:0560: OKAY BOLT GUN).
func TestFightRowWeaponAndStatus(t *testing.T) {
	dst := gpu.NewImage(1280, 800)
	defer dst.Deallocate()
	var drawn strings.Builder
	g := func(c byte) [8]byte { drawn.WriteByte(c); return [8]byte{} }
	b := &combat.Board{Turn: &combat.Turn{Unit: 1}, Units: []combat.Unit{{ID: 1, Name: "GIZMO", HP: 43, AC: 2, Weapon: "BOLT GUN", Status: "OKAY", Party: true}}}
	(&Presenter{}).drawFightRows(dst, Note{Glyph: g}, b, 1)
	if s := drawn.String(); !strings.Contains(s, "BOLT GUN") || !strings.Contains(s, "OKAY") || strings.Index(s, "OKAY") > strings.Index(s, "BOLT GUN") {
		t.Fatalf("drawn %q", s)
	}
}
