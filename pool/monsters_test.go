package pool

import (
	"testing"

	"gobox/core/goldbox"
)

// TestMonstersLoad (bug class: pool read every game's MON records in POR's 285-byte
// layout, so cotab's, sotsb's and pod's first monster panicked): every monster of the
// game's data loads through LoadMonster's rules (core.MakeMonster) with its name.
func TestMonstersLoad(t *testing.T) { monstersLoad(t, quickStarted(t)) }

// TestMonstersLoadCOTAB: the same on cotab's imported folder.
func TestMonstersLoadCOTAB(t *testing.T) {
	monstersLoad(t, quickStartedOn(t, advEngine(t, advEnvCOTAB)))
}

// monstersLoad is TestMonstersLoad on e.
func monstersLoad(t *testing.T, e *Engine) {
	t.Helper()
	n := 0
	for s := goldbox.Set(1); s <= sets; s++ {
		e.set = s
		for id := range 256 {
			if _, ok := e.gdata().monster(s.File("MON"), id); !ok {
				continue
			}
			c, ok := e.LoadMonster(id)
			if !ok || e.name(c) == "" {
				t.Fatalf("%s %d: not loaded", s.File("MON"), id)
			}
			n++
		}
	}
	if n == 0 {
		t.Fatal("no monsters")
	}
}
