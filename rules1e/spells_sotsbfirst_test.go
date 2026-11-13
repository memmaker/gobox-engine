package rules1e

import "testing"

// delayedBlastScene: the cotab scene with a Delayed Blast Fireball record
// and its handler.
func delayedBlastScene(t *testing.T, foes ...*Character) *RuleState {
	e := cotabFirstScene(t, foes...)
	for len(e.T().Spells) <= int(spDelayedBlastFireball) {
		e.T().Spells = append(e.T().Spells, make([]SpellProp, len(POR.Spells[1])))
	}
	e.T().Spells[spDelayedBlastFireball][0] = SpellProp(SpellClassMage)
	e.T().Affects[affDelayedBlastSotsb] = "delayedBlastSotsb"
	e.T().Grounds = SotsbGroundIDs
	e.Fight.On = true
	return e
}

// TestDelayedBlastRuling: delay 0 goes off at once as a fireball of
// (level)d6+1 per die; a delay of rounds waits on the square as the caster's
// affect and bursts on everyone within three squares when it ends.
func TestDelayedBlastRuling(t *testing.T) {
	foe := newChar("FOE", 9)
	e := delayedBlastScene(t, foe)
	lvl := int(e.Party[0].Levels[ClsMage])
	e.Fight.Choice = 0
	castOn(e, spDelayedBlastFireball, foe)
	if lost := 90 - int(foe.HP); lost < lvl/2 || lost > 7*lvl || len(e.Clouds) != 0 {
		t.Fatalf("instant: lost %d HP, %d records", lost, len(e.Clouds))
	}

	foe = newChar("FOE", 9)
	e = delayedBlastScene(t, foe)
	e.Fight.Choice = 9 // past the most: delayedBlastMaxRounds
	e.TgtX, e.TgtY = 13, 10
	castOn(e, spDelayedBlastFireball)
	a := e.FindAffect(e.Party[0], affDelayedBlastSotsb)
	if foe.HP != 90 || a == nil || a.Time != delayedBlastMaxRounds || len(e.Clouds) != 1 || e.Cmap.Ground[10][13] != SotsbGroundIDs.DelayedBlast {
		t.Fatalf("waiting: HP %d, affect %+v, %d records, ground %#x", foe.HP, a, len(e.Clouds), e.Cmap.Ground[10][13])
	}
	e.RemoveAffect(e.Party[0], affDelayedBlastSotsb, a)
	if lost := 90 - int(foe.HP); lost < lvl/2 || lost > 7*lvl {
		t.Errorf("burst: lost %d HP", lost)
	}
	if len(e.Clouds) != 0 || e.Cmap.Ground[10][13] != PorGroundIDs.Floor {
		t.Errorf("after the burst: %d records, ground %#x", len(e.Clouds), e.Cmap.Ground[10][13])
	}
}

// TestDelayedBlastAsks: in a fight the spell's aim carries its question.
func TestDelayedBlastAsks(t *testing.T) {
	e := delayedBlastScene(t)
	a := (&Rules{S: e}).spellAim(SpellIDs[spDelayedBlastFireball])
	if a.Choose == "" || len(a.Choices) != delayedBlastMaxRounds+1 {
		t.Errorf("aim %+v: no delay question", a)
	}
}

// TestDelayedBlastLevel: the burst rolls the level it was cast at, past
// 15 too, whatever the caster's level when it goes off.
func TestDelayedBlastLevel(t *testing.T) {
	foe := newChar("FOE", 9)
	foe.HP = 200
	e := delayedBlastScene(t, foe)
	const lvl = 20
	e.Party[0].Levels[ClsMage] = lvl
	e.Fight.Choice = 1
	e.TgtX, e.TgtY = 13, 10
	castOn(e, spDelayedBlastFireball)
	a := e.FindAffect(e.Party[0], affDelayedBlastSotsb)
	if a == nil || len(e.Clouds) != 1 || e.Clouds[0].Level != lvl {
		t.Fatalf("waiting: affect %+v, records %d", a, len(e.Clouds))
	}
	e.Party[0].Levels[ClsMage] = 1
	e.RemoveAffect(e.Party[0], affDelayedBlastSotsb, a)
	if lost := 200 - int(foe.HP); lost < lvl {
		t.Errorf("burst: lost %d HP, want at least %d (half of %dd6+%d)", lost, lvl, lvl, lvl)
	}
}
