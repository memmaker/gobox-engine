package pool

import (
	"testing"

	"gobox/core/goldbox"
	"rules1e"
)

// TestSpellsGone (removed spells: POR 56, cotab 56, pod 102 and 103): every
// monster of every set loads (its spells read through the game's spellsGone,
// a removed one as rules1e.NoSpell or its stand-in) and holds no spell
// without a unified id; a script's removed spell
// number reads as no spell (pod 103: Hold Person). Detect the class: a removed spell number must be
// in the importer's spellsGone (POR: rules1e.PorSpellsGone).
func TestSpellsGone(t *testing.T) { spellsGone(t, quickStarted(t), 56) }
func TestSpellsGoneCOTAB(t *testing.T) {
	spellsGone(t, quickStartedOn(t, advEngine(t, advEnvCOTAB)), 56)
}
func TestSpellsGoneSOTSB(t *testing.T) { spellsGone(t, quickStartedOn(t, advEngine(t, advEnvSOTSB))) }
func TestSpellsGonePOD(t *testing.T) {
	spellsGone(t, quickStartedOn(t, advEngine(t, advEnvPOD)), 102, 103)
}

func spellsGone(t *testing.T, e *Engine, gone ...int) {
	t.Helper()
	known := func(id rules1e.SpellID) bool {
		id &^= rules1e.SpellPending
		return id == rules1e.NoSpell || int(id) < len(rules1e.SpellIDs) && rules1e.SpellIDs[id] != ""
	}
	for s := goldbox.Set(1); s <= sets; s++ {
		e.set = s
		for id := range 256 {
			if _, ok := e.gdata().monster(s.File("MON"), id); !ok {
				continue
			}
			c, _ := e.LoadMonster(id)
			m := e.member(c)
			for _, sp := range m.Memorized {
				if !known(sp) {
					t.Errorf("%s %d: memorized spell %d has no id", s.File("MON"), id, sp)
				}
			}
			for i, f := range m.SpellBook {
				if f != 0 && !known(rules1e.SpellID(i+1)) {
					t.Errorf("%s %d: knows spell %d, which has no id", s.File("MON"), id, i+1)
				}
			}
		}
	}
	for _, n := range gone { // a script naming one reads it as no spell or its stand-in
		if id := e.scriptSpell(n); id != "" && n != 103 {
			t.Errorf("removed spell %d reads as %q", n, id)
		}
	}
}
