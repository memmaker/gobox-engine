package engine

import (
	"fmt"
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

func TestMonstersLoadSOTSB(t *testing.T) {
	monstersLoad(t, quickStartedOn(t, advEngine(t, advEnvSOTSB)))
}

func TestMonstersLoadPOD(t *testing.T) {
	monstersLoad(t, quickStartedOn(t, advEngine(t, advEnvPOD)))
}

// TestItemsLoad (bug class: pod's import wrote its items as raw 17-byte library
// records, which pool reads as char.Item): every game's ITEM<n> blocks read as
// char.Item (a broken file panics), and some item has a type or a name.
// Detect the class: every dax.ImportData call passes char.SheetOfRecord and char.ItemOfRecord.
func TestItemsLoad(t *testing.T)      { itemsLoad(t, testEngine(t)) }
func TestItemsLoadCOTAB(t *testing.T) { itemsLoad(t, advEngine(t, advEnvCOTAB)) }
func TestItemsLoadSOTSB(t *testing.T) { itemsLoad(t, advEngine(t, advEnvSOTSB)) }
func TestItemsLoadPOD(t *testing.T)   { itemsLoad(t, advEngine(t, advEnvPOD)) }

func itemsLoad(t *testing.T, e *Engine) {
	t.Helper()
	n := 0
	for s := 0; s <= sets; s++ {
		for id := range 256 {
			for _, it := range e.gdata().itemBlock(fmt.Sprintf("ITEM%d", s), id) {
				if it.Type != 0 || it.PName[0] != 0 {
					n++
				}
			}
		}
	}
	if n == 0 {
		t.Fatal("no items")
	}
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
