package pool

import (
	"gobox/char"
	"gobox/core"
	"os"
	"rules1e"
	"slices"
	"testing"
)

// testEngine is an engine on the game files (the test skips without them),
// its saves under a temp dir: tests never read or write the game's own.
func testEngine(t *testing.T) *Engine {
	t.Helper()
	if _, err := os.Stat("../../por-go/data/dos/TITLE.DAX"); err != nil {
		t.Skip("no game files")
	}
	e, err := New(char.Rules1e, porTables, "../../por-go/data/assets", "../../por-go/data")
	if err != nil {
		t.Fatal(err)
	}
	e.Cfg.SaveDir = t.TempDir()
	return e
}

// quickStarted: an engine with the Quick Start party, its roster written
// under a temp dir: tests never read or write the game's own saves.
func quickStarted(t *testing.T) *Engine {
	t.Helper()
	e := testEngine(t)
	seed := uint32(7) // the same party every run
	e.FixedSeed = &seed
	e.Init()
	if err := e.QuickStart(); err != nil {
		t.Fatal(err)
	}
	if len(e.state().Party) == 0 {
		t.Fatal("no Quick Start party")
	}
	return e
}

// TestNoMissileSlot: the Missiles and Launcher slots were por additions.
// A thrown weapon or a launcher readies in hand, as the original's, and
// Attack aims with it; arrows go beside the bow. An old save's Readied 2
// or 3 item goes back to the pack.
func TestNoMissileSlot(t *testing.T) {
	e := quickStarted(t)
	i := slices.IndexFunc(e.state().Party, func(c *rules1e.Character) bool {
		return e.state().Base(rules1e.ItemShortBow).Classes&e.state().Base(rules1e.ItemJavelin).Classes&c.ClassMask != 0
	})
	if i < 0 {
		t.Fatal("nobody for a bow and a javelin")
	}
	e.selectMember(i)
	c := e.member(e.selected())
	e.as(e.sel) // the rules, c acting
	r := e.state()
	r.ApplyItems(c)
	for _, s := range []int{rules1e.SlotWeapon, rules1e.SlotShield, rules1e.SlotArrows, rules1e.SlotBolts} {
		if c.Slots[s] != nil {
			r.ReadyItem(c.Slots[s])
			r.ApplyItems(c)
		}
	}
	j := e.state().AddItemCopy(c, e.state().RandomItem(rules1e.ItemJavelin))
	if e.state().DollSlot(c, j) != core.DollWeapon {
		t.Errorf("javelin: doll slot %d, want the weapon hand", e.state().DollSlot(c, j))
	}
	r.ReadyItem(j)
	r.ApplyItems(c)
	if j.Readied != 1 || c.Slots[rules1e.SlotWeapon] != j {
		t.Errorf("javelin readied %d, in hand %v", j.Readied, c.Slots[rules1e.SlotWeapon] == j)
	}
	r.ReadyItem(j)
	r.ApplyItems(c)
	bow, arrows := e.state().AddItemCopy(c, e.state().RandomItem(rules1e.ItemShortBow)), e.state().AddItemCopy(c, e.state().RandomItem(rules1e.ItemArrow))
	r.ReadyItem(bow)
	r.ReadyItem(arrows)
	r.ApplyItems(c)
	if bow.Readied != 1 || c.Slots[rules1e.SlotWeapon] != bow || c.Slots[rules1e.SlotArrows] != arrows {
		t.Errorf("bow readied %d in hand %v, arrows %v (%s)", bow.Readied, c.Slots[rules1e.SlotWeapon] == bow, c.Slots[rules1e.SlotArrows] == arrows, e.ext.note)
	}
	// an old save's Missiles/Launcher items: its import puts them back in the
	// pack (saveimport.go, U10; ApplyItems reads the current format only)
	bow.Readied, arrows.Readied = rules1e.ReadiedOldLauncher, rules1e.ReadiedOldMissiles
	r.ImportCharacter(c)
	r.ApplyItems(c)
	if bow.Readied != 0 || arrows.Readied != 0 || c.Slots[rules1e.SlotWeapon] != nil || c.Slots[rules1e.SlotArrows] != nil {
		t.Errorf("old save: bow %d, arrows %d, weapon %v", bow.Readied, arrows.Readied, c.Slots[rules1e.SlotWeapon])
	}
}
