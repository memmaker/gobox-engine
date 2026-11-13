package present

import (
	"strconv"
	"strings"
	"testing"

	"gobox/game"
	"gobox/gpu"
)

// TestSheetLongRulesStopAboveFooter: a long rule-set list (BR2's skills)
// stops above AC/THAC0/Damage and counts the rest.
func TestSheetLongRulesStopAboveFooter(t *testing.T) {
	dst := gpu.NewImage(1280, 800)
	defer dst.Deallocate()
	var drawn strings.Builder
	g := func(c byte) [8]byte { drawn.WriteByte(c); return [8]byte{} }
	note := Note{Glyph: g, Font: func(c byte, _ bool) [8]byte { return g(c) }}
	var rules []game.Stat
	for i := range 40 {
		rules = append(rules, game.Stat{Name: "SKILL" + strconv.Itoa(i), Value: "1"})
	}
	sheetOn(&game.SheetPage{Rules: rules, THAC0: 9}).drawSheet(dst, note)
	if s := drawn.String(); strings.Contains(s, "SKILL39") || !strings.Contains(s, "more") {
		t.Errorf("long list drawn to the end, into the footer")
	}
}

// TestSheetNineSkills: BR2's longest list (tech and 8 career skills, 9
// rows) fits whole above the footer: no "+N more".
func TestSheetNineSkills(t *testing.T) {
	dst := gpu.NewImage(1280, 800)
	defer dst.Deallocate()
	var drawn strings.Builder
	g := func(c byte) [8]byte { drawn.WriteByte(c); return [8]byte{} }
	note := Note{Glyph: g, Font: func(c byte, _ bool) [8]byte { return g(c) }}
	var rules []game.Stat
	for i := range 9 {
		rules = append(rules, game.Stat{Name: "SKILL" + strconv.Itoa(i), Value: "1"})
	}
	sheetOn(&game.SheetPage{Rules: rules, THAC0: 9}).drawSheet(dst, note)
	if s := drawn.String(); !strings.Contains(s, "SKILL8") || strings.Contains(s, "more") {
		t.Errorf("9 skills don't fit")
	}
}
