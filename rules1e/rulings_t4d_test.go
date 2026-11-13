package rules1e

import "testing"

// TestRulingsT4d: Felix's affect rulings, the same for all games.
func TestRulingsT4d(t *testing.T) {
	// Feeblemind: Int and Wis 3.
	e, m, foe := aimScene(t, "feeblemindedSotsb", affFeebleminded)
	e.callAffect(affFeebleminded, foe, &Affect{AffectRecord: AffectRecord{Type: affFeebleminded}}, 0)
	if foe.Stats[StatInt] != 3 || foe.Stats[StatWis] != 3 {
		t.Errorf("feeblemind: Int %d Wis %d", foe.Stats[StatInt], foe.Stats[StatWis])
	}

	// Paralysing touch: 2d8 rounds on a foe close by, none from afar.
	held := false
	for seed := uint32(1); seed < 30 && !held; seed++ {
		e, m, foe = aimScene(t, "paralyse2d8", affParalyse2d8)
		e.RandSeed = seed
		m.Combat.Target = foe
		e.paralyse2d8(m, nil, 0)
		if a := e.FindAffect(foe, AffHeld); a != nil {
			held = true
			if a.Time < 2 || a.Time > 16 {
				t.Errorf("paralysing touch: %d rounds", a.Time)
			}
		}
	}
	if !held {
		t.Error("paralysing touch never held")
	}

	// A sting is blocked by an immunity.
	for seed := uint32(1); seed < 30; seed++ {
		e, m, foe = aimScene(t, "poisonBite", affPoisonBite)
		e.RandSeed = seed
		m.Combat.Target = foe
		e.AddAffect(foe, affBlockPoisonHold, 0, 0, 0)
		e.poisonAttack(m, 0)
		if foe.Status == StatusDead {
			t.Fatalf("seed %d: poison-immune foe killed", seed)
		}
	}

	// Slow Poison running out while poisoned: unconscious.
	e, _, foe = aimScene(t, "poisonEnds", affSlowPoison)
	e.AddAffect(foe, AffPoisoned, 0, 0xFF, 0)
	e.poisonEnds(foe, nil, 0)
	if foe.Status != StatusUnconscious {
		t.Errorf("slow poison ends: status %d", foe.Status)
	}

	// Paladin cures: the higher of the two levels.
	c := newChar("P", 1)
	c.Levels[ClsPaladin], c.OldLevels[ClsPaladin] = 12, 4
	(&RuleState{Tables: &POR}).paladinCures(c, nil, 1)
	if c.PaladinCures != 12/paladinCureLevels+1 {
		t.Errorf("paladin cures %d", c.PaladinCures)
	}

	// Ranger vs giants: + ranger level against giant kind, at any range.
	e, m, foe = aimScene(t, "rangerVsGiants", AffRangerVsGiants)
	m.Levels[ClsRanger], foe.Kind = 5, porGiants.Kind
	m.Combat.Target = foe
	e.AddAffect(m, AffRangerVsGiants, 0, 0xFF, 0)
	e.Rv.Damage = 10
	e.AffectGroup(m, 4)
	if e.Rv.Damage != 15 {
		t.Errorf("ranger vs giants: damage %d", e.Rv.Damage)
	}
	// The later games: the creature flag tells, not the Kind.
	tb := *e.T()
	tb.Giants = LaterGiants
	e.Tables = &tb
	e.Rv.Damage = 10
	e.AffectGroup(m, 4)
	if e.Rv.Damage != 10 {
		t.Errorf("ranger vs giants (flag rule, Kind giant, no flag): damage %d", e.Rv.Damage)
	}
	foe.CreatureFlags = LaterGiants.Flag
	e.AffectGroup(m, 4)
	if e.Rv.Damage != 15 {
		t.Errorf("ranger vs giants (flag rule): damage %d", e.Rv.Damage)
	}
}

// TestMirrorImageRuled (T4e; Felix's ruling, all games, POR included): the
// images sit in the data's high nibble; a blow or a spell flagged in its
// record's spellImageField may take one, an unflagged spell never does. POR's
// own rule (a data count, spells only) is gone; no Pascal vector covers it.
func TestMirrorImageRuled(t *testing.T) {
	foe := newChar("FOE", 1)
	e := cotabFirstScene(t, foe)
	if POR.Spells[SpMagicMissile][spellImageField] == 0 || POR.Spells[SpBless][spellImageField] != 0 {
		t.Fatal("POR's image flags: magic missile flagged, bless not")
	}
	lost := func(spell SpellID) bool {
		e.AddAffect(foe, affMirrorImage, 0, 2*imageStep, 0)
		e.CurSpell = spell
		taken := 0
		for range 50 {
			e.Rv.AffVal = 1
			e.checkAffect(foe, affMirrorImage)
			if e.Rv.AffVal == 0 {
				taken++
				a := e.FindAffect(foe, affMirrorImage)
				if a != nil && a.Data != imageStep {
					t.Errorf("spell %d: data %#x after one image, want %#x", spell, a.Data, imageStep)
				}
			}
			if e.FindAffect(foe, affMirrorImage) == nil {
				break
			}
		}
		e.RemoveAffect(foe, affMirrorImage, nil)
		return taken > 0
	}
	if !lost(0) || !lost(SpMagicMissile) {
		t.Error("a blow and a flagged spell take images")
	}
	if lost(SpBless) {
		t.Error("an unflagged spell took an image")
	}
}
