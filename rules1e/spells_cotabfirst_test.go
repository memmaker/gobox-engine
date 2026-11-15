package rules1e

import (
	"gobox/rules"
	"slices"
	"testing"
)

// cotabFirstScene is a caster (Party[0]) and foes on a bare combat map,
// with spell records for the unified ids past POR's.
func cotabFirstScene(t *testing.T, foes ...*Character) *RuleState {
	t.Helper()
	tb := POR
	tb.Spells = slices.Clone(POR.Spells)
	for len(tb.Spells) <= int(spBestowCurseMage) {
		tb.Spells = append(tb.Spells, make([]SpellProp, len(POR.Spells[1])))
	}
	tb.Spells[spEntangle][0] = SpellProp(SpellClassDruid)
	tb.Spells[spCloudKill][0] = SpellProp(SpellClassMage)
	tb.Affects = map[AffectType]HandlerName{}
	for k, v := range POR.Affects {
		tb.Affects[k] = v
	}
	tb.Affects[affCloudkill] = "cloudkill"
	tb.Grounds = CotabGroundIDs
	e := &RuleState{Tables: &tb, Phase: rules.PhaseFight, RandSeed: 1}
	caster := newChar("CASTER", 0)
	caster.Levels[ClsMage], caster.Levels[ClsRanger] = 9, 9
	e.Party = append([]*Character{caster}, foes...)
	e.Cmap = &CombatMap{}
	for y := range e.Cmap.Ground {
		for x := range e.Cmap.Ground[y] {
			e.Cmap.Ground[y][x] = PorGroundIDs.Floor
		}
	}
	for i, c := range e.Party {
		if c.Combat == nil {
			c.Combat = &CombatInfo{Moves: 6}
		}
		e.CombatantSlots = append(e.CombatantSlots, c)
		e.Cpos[i+1] = CombatPos{X: byte(10 + i), Y: 10, Size: 1}
	}
	e.Ncomb = len(e.Party)
	e.CombatRedrawMap()
	return e
}

// newChar is a fighter of hd hit dice with 10 hit points a die (at least 10).
func newChar(name string, hd byte) *Character {
	c := &Character{}
	c.Okay, c.MaxLevel = 1, hd
	c.CharName = name
	c.HP = byte(max(10, 10*int(hd)))
	c.MaxHP = c.HP
	return c
}

func castOn(e *RuleState, spell SpellID, targets ...*Character) {
	e.CurSpell, e.SpellCaster = spell, e.Party[0]
	e.NSpellTargets = len(targets)
	for i, c := range targets {
		e.SetTgt(i+1, c)
	}
	e.spellEffect(spell)
}

// TestCloudKillRuling: Felix's Cloud Kill: weak foes die at once, strong
// ones take 1d10 a round; the cloud's end gives the squares back.
func TestCloudKillRuling(t *testing.T) {
	weak := newChar("WEAK", 3)
	strong := newChar("STRONG", 9)
	e := cotabFirstScene(t, weak, strong)
	e.TgtX, e.TgtY = 12, 10 // weak at 11, strong at 12: both inside the nine squares
	castOn(e, spCloudKill)
	if weak.Status != StatusDead || weak.Okay != 0 {
		t.Errorf("weak foe: status %d, want dead", weak.Status)
	}
	if strong.HP >= 90 || strong.HP < 90-cloudKillDamageSides || strong.Okay == 0 {
		t.Errorf("strong foe: HP %d, want 1d10 poison damage", strong.HP)
	}
	if e.Cmap.Ground[10][12] != e.grounds().PoisonCloud || len(e.Clouds) != 1 {
		t.Fatalf("cloud not laid: ground %#x, %d clouds", e.Cmap.Ground[10][12], len(e.Clouds))
	}
	hp := strong.HP
	e.inClouds(strong) // a round in the cloud
	if strong.HP >= hp {
		t.Errorf("no damage in the second round: HP %d", strong.HP)
	}
	a := e.FindAffect(e.Party[0], affCloudkill)
	if a == nil {
		t.Fatal("caster has no Cloud Kill affect")
	}
	e.RemoveAffect(e.Party[0], affCloudkill, a)
	e.callAffect(affCloudkill, e.Party[0], a, 0)
	if len(e.Clouds) != 0 || e.Cmap.Ground[10][12] != PorGroundIDs.Floor {
		t.Errorf("cloud not cleared: %d clouds, ground %#x", len(e.Clouds), e.Cmap.Ground[10][12])
	}
}

// TestEntangleRuling: Felix's Entangle: outdoors, a failed save vs spell
// Befuddles for 6 + 3/level rounds and the turn is lost; in a dungeon nothing.
func TestEntangleRuling(t *testing.T) {
	foe := newChar("FOE", 1)
	foe.Saves[SaveVsSpell] = 19
	e := cotabFirstScene(t, foe)
	e.Dungeon = true
	seed := e.RandSeed
	castOn(e, spEntangle, foe)
	if len(foe.Affects) != 0 || e.RandSeed != seed {
		t.Fatal("Entangle worked in a dungeon")
	}
	e.Dungeon = false
	want := entangleBase + entanglePerLevel*e.casterLevel(int(spEntangle))
	for s := uint32(1); s < 20; s++ {
		foe.Affects = nil
		e.RandSeed = s
		castOn(e, spEntangle, foe)
		if a := e.FindAffect(foe, affBefuddledPod); a != nil {
			if int(a.Time) != want {
				t.Errorf("rounds %d, want %d", a.Time, want)
			}
			foe.Combat.Moves = 6
			e.callAffect(affBefuddledPod, foe, a, 0)
			if foe.Combat.Moves != 0 {
				t.Error("Befuddled did not end the turn")
			}
			return
		}
	}
	t.Error("no seed failed the save")
}
