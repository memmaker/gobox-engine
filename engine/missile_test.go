package engine

import (
	"gobox/rules"
	"gobox/rules1e"
	"os"
	"slices"
	"testing"
)

// testEngine is an engine on the game files (the test skips without them),
// its saves under a temp dir: tests never read or write the game's own.
func testEngine(t *testing.T) *Engine {
	t.Helper()
	if _, err := os.Stat("../data/por/sources/dos/TITLE.DAX"); err != nil {
		t.Skip("no game files")
	}
	e, err := New(rules1e.Rules1e, porTables, "../data/por", "../data/por/sources")
	if err != nil {
		t.Fatal(err)
	}
	e.Cfg.SaveDir = t.TempDir()
	if _, err := e.r1e().Resolve(nil, rules.LoadItemTable{}, nil); err != nil { // RandomItem converts through it
		t.Fatal(err)
	}
	return e
}

// quickStarted: an engine with the Quick Start party, its roster written
// under a temp dir: tests never read or write the game's own saves.
func quickStarted(t *testing.T) *Engine { t.Helper(); return quickStartedOn(t, testEngine(t)) }

// quickStartedOn is quickStarted on the engine e (any game's: advEngine).
func quickStartedOn(t *testing.T, e *Engine) *Engine {
	t.Helper()
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

// fighterWith: a Quick Start member who may use all of types, selected
// and acting, its hands emptied.
func fighterWith(t *testing.T, e *Engine, types ...rules1e.ItemType) (*rules1e.RuleState, *rules1e.Character) {
	t.Helper()
	r := e.state()
	i := slices.IndexFunc(r.Party, func(c *rules1e.Character) bool {
		return !slices.ContainsFunc(types, func(ty rules1e.ItemType) bool { return r.Base(ty).Classes&c.ClassMask == 0 })
	})
	if i < 0 {
		t.Fatal("nobody for the items")
	}
	e.selectMember(i)
	c := e.member(e.selected())
	e.as(e.sel)
	r.ApplyItems(c)
	for _, it := range slices.Clone(c.Items) {
		if it.Readied != 0 {
			r.ReadyItem(it)
			r.ApplyItems(c)
		}
	}
	return r, c
}

// ready adds a copy of type ty to c and readies it.
func ready(r *rules1e.RuleState, c *rules1e.Character, ty rules1e.ItemType) *rules1e.Item {
	it := r.AddItemCopy(c, r.RandomItem(ty))
	r.ReadyItem(it)
	r.ApplyItems(c)
	return it
}

// TestLauncherMissilesSlots (extra): a bow readies into the Launcher slot
// and a thrown-only weapon into the Missiles slot, by their base records,
// beside sword and shield (they take no hands); Attack uses the sword,
// Ranged (Wield) is the bow with its arrows, or the thrown flask once it has
// taken the arrows' Missiles box.
func TestLauncherMissilesSlots(t *testing.T) {
	e := quickStarted(t)
	r, c := fighterWith(t, e, rules1e.ItemLongSword, rules1e.ItemShield, rules1e.ItemShortBow, rules1e.ItemOil)
	sword, shield := ready(r, c, rules1e.ItemLongSword), ready(r, c, rules1e.ItemShield)
	bow, arrows := ready(r, c, rules1e.ItemShortBow), ready(r, c, rules1e.ItemArrow)
	if c.Slots[rules1e.SlotWeapon] != sword || c.Slots[rules1e.SlotShield] != shield || c.Slots[rules1e.SlotLauncher] != bow || c.Slots[rules1e.SlotArrows] != arrows {
		t.Fatalf("sword %v shield %v bow %v arrows %v", c.Slots[rules1e.SlotWeapon] == sword, c.Slots[rules1e.SlotShield] == shield, c.Slots[rules1e.SlotLauncher] == bow, c.Slots[rules1e.SlotArrows] == arrows)
	}
	if d := r.DollSlot(c, bow); d != rules.DollLauncher {
		t.Errorf("bow on doll slot %d", d)
	}
	e.setPhase(rules.PhaseFight)
	cID := e.charID(c)
	tw := &Turn{}
	e.turnWeapons(cID, tw)
	if tw.Weapon == nil || tw.Missile == nil || tw.Weapon.Item != e.r1e().Handle(sword) || tw.Missile.Item != e.r1e().Handle(bow) {
		t.Errorf("turn weapons %+v / %+v", tw.Weapon, tw.Missile)
	}
	var ammo *rules1e.Item
	if w, ok := r.Ranged(c); !ok || w.Weapon != bow || !r.AmmoFor(c, w, &ammo) || ammo != arrows || r.Hand(c).Weapon != sword {
		t.Errorf("Ranged: the bow %v, ammo arrows %v, the sword still in hand %v", w.Weapon == bow, ammo == arrows, r.Hand(c).Weapon == sword)
	}
	oil := r.AddItemCopy(c, r.RandomItem(rules1e.ItemOil))
	r.SwapReady(c, oil) // the Missiles box: the arrows come off
	r.ApplyItems(c)
	if c.Slots[rules1e.SlotMissiles] != oil || c.Slots[rules1e.SlotArrows] != nil || c.Slots[rules1e.SlotWeapon] != sword || c.Slots[rules1e.SlotShield] != shield {
		t.Fatalf("oil in Missiles %v, arrows %v, sword %v (%s)", c.Slots[rules1e.SlotMissiles] == oil, c.Slots[rules1e.SlotArrows], c.Slots[rules1e.SlotWeapon] == sword, e.ext.note)
	}
	if w, ok := r.Ranged(c); !ok || w.Weapon != oil || !r.AmmoFor(c, w, &ammo) || ammo != oil || r.Hand(c).Weapon != sword {
		t.Errorf("Ranged: the flask %v, used up %v, the sword still in hand %v", w.Weapon == oil, ammo == oil, r.Hand(c).Weapon == sword)
	}
}

// TestOldSaveMissileSlots: an old save's Readied 2/3 items load into the
// Missiles and Launcher slots their base records name; anything else
// there goes back to the pack (saveimport.go, U10).
func TestOldSaveMissileSlots(t *testing.T) {
	e := quickStarted(t)
	r, c := fighterWith(t, e, rules1e.ItemLongSword, rules1e.ItemShortBow, rules1e.ItemJavelin)
	sword := ready(r, c, rules1e.ItemLongSword)
	bow, jav, extra := r.RandomItem(rules1e.ItemShortBow), r.RandomItem(rules1e.ItemJavelin), r.RandomItem(rules1e.ItemLongSword)
	c.Items = append(c.Items, bow, jav, extra) // as a save holds them
	bow.Readied, jav.Readied, extra.Readied = rules1e.ReadiedOldLauncher, rules1e.ReadiedOldMissiles, rules1e.ReadiedOldLauncher
	r.ImportCharacter(c)
	r.ApplyItems(c)
	if c.Slots[rules1e.SlotLauncher] != bow || c.Slots[rules1e.SlotMissiles] != jav || c.Slots[rules1e.SlotWeapon] != sword || extra.Readied != 0 {
		t.Errorf("bow %v javelin %v sword %v, second sword readied %d", c.Slots[rules1e.SlotLauncher] == bow, c.Slots[rules1e.SlotMissiles] == jav, c.Slots[rules1e.SlotWeapon] == sword, extra.Readied)
	}
}
