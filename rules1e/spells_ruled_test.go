package rules1e

import (
	"slices"
	"strings"
	"testing"

	"gobox/core"
)

// ruledRecords are the spell records the ruled tests need: CotAB's rows
// (SotSB's for the spells CotAB lacks), field 0 as a SpellClass.
var ruledRecords = map[SpellID][16]SpellProp{
	spCureSeriousWounds:           {0, 4, 0, 0, 0, 0, 4, 2, 0, 4, 0, 2, 7, 1, 0, 0},
	spProtectionEvil10Radius:      {0, 4, 3, 0, 0, 10, 4, 2, 0, 4, 45, 2, 7, 2, 0, 0},
	spSticksToSnakes:              {0, 4, 3, 0, 0, 2, 4, 0, 0, 4, 3, 1, 7, 4, 1, 0},
	spDispelEvil:                  {0, 5, 0, 0, 0, 1, 0, 0, 0, 4, 145, 1, 8, 3, 0, 0},
	spFlameStrike:                 {0, 5, 6, 0, 0, 0, 4, 0, 2, 4, 0, 1, 8, 6, 1, 0},
	spSlayLiving:                  {0, 5, 3, 0, 0, 0, 4, 0, 1, 4, 0, 1, 10, 7, 1, 0},
	spFaerieFire:                  {2, 1, 8, 0, 0, 4, 5, 0, 1, 4, 7, 1, 3, 4, 1, 0},
	spConfusion:                   {1, 4, 12, 0, 2, 1, 11, 0, 1, 4, 35, 1, 4, 7, 1, 0},
	spFear:                        {1, 4, 6, 0, 0, 1, 8, 0, 1, 4, 142, 1, 4, 6, 1, 0},
	spFireShield:                  {1, 4, 0, 0, 2, 1, 0, 1, 0, 4, 0, 2, 4, 8, 0, 0},
	spFumble:                      {1, 4, 0, 1, 0, 1, 4, 0, 1, 4, 27, 1, 4, 4, 1, 0},
	spIceStorm:                    {1, 4, 0, 1, 0, 0, 10, 0, 0, 4, 0, 1, 4, 7, 1, 0},
	spConeOfCold:                  {1, 5, 6, 0, 0, 0, 8, 0, 2, 4, 0, 1, 5, 6, 1, 0},
	spFeeblemind:                  {1, 5, 16, 0, 0, 0, 4, 0, 1, 4, 68, 1, 5, 6, 1, 0},
	spHoldMonsters:                {1, 5, 0, 1, 0, 1, 7, 0, 1, 4, 52, 1, 5, 7, 1, 0},
	spCharmPersonOrMammal:         {2, 2, 8, 0, 60, 60, 4, 0, 1, 4, 11, 1, 4, 5, 1, 0},
	spMinorGlobeOfInvulnerability: {1, 4, 0, 1, 0, 1, 0, 1, 0, 4, 63, 2, 4, 5, 0, 0},
}

// ruledScene is cotabFirstScene with CotAB's affect ids, handlers and caster
// rule, the ruled records and a cleric-mage caster of level 9.
func ruledScene(t *testing.T, foes ...*Character) *RuleState {
	t.Helper()
	e := cotabFirstScene(t, foes...)
	tb := e.Tables
	for len(tb.Spells) < len(SpellIDs) {
		tb.Spells = append(tb.Spells, make([]SpellProp, len(POR.Spells[1])))
	}
	for id, r := range ruledRecords {
		tb.Spells[id] = slices.Clone(r[:])
	}
	var err error
	if tb.Affects, err = withVariants(tb.Affects, CotabAffectVariants); err != nil {
		t.Fatal(err)
	}
	tb.AffectIDs, tb.CasterLevel = CotabAffectIDs, CotabCasterLevel
	e.Party[0].Levels[ClsCleric] = 9
	return e
}

// said is the text the cast narrated.
func said(e *RuleState) string {
	var b []string
	for _, f := range e.Fx {
		if n, ok := f.(core.Narrated); ok {
			b = append(b, n.Text)
		}
	}
	return strings.Join(b, "|")
}

// TestAreaSpellsRuled: the area spells mark the cast as an area one, so no
// mirror image is used up.
func TestAreaSpellsRuled(t *testing.T) {
	for _, sp := range []SpellID{spFlameStrike, spConfusion, spFear, spIceStorm, spConeOfCold, spDeathSpell, spMassInvisibility, spDelayedBlastFireball} {
		foe := newChar("FOE", 1)
		foe.Side = 1
		e := ruledScene(t, foe)
		e.TgtX, e.TgtY = 11, 10
		castOn(e, sp, foe)
		if e.areaSpell != 1 {
			t.Errorf("spell %d: no area flag", sp)
		}
	}
	foe := newChar("FOE", 1)
	e := ruledScene(t, foe)
	e.AddAffect(foe, affMirrorImage, 0, 3<<imageShiftCotab, 0)
	castOn(e, spFlameStrike, foe)
	if a := e.FindAffect(foe, affMirrorImage); a == nil || a.Data != 3<<imageShiftCotab || foe.HP == foe.MaxHP {
		t.Errorf("Flame Strike: images %v, HP %d", a, foe.HP)
	}
}

// TestHealWakeRuled: 68/74/102 wake the unconscious out of combat and do not
// heal the mummy-diseased.
func TestHealWakeRuled(t *testing.T) {
	for _, sp := range []SpellID{spCureSeriousWounds, spCureCriticalWounds, spPotionOfExtraHealing} {
		c := newChar("HURT", 1)
		c.HP, c.Okay, c.Status = 0, 0, StatusUnconscious
		e := ruledScene(t, c)
		e.BeginFlag = 0
		castOn(e, sp, c)
		if c.Okay == 0 || c.HP == 0 {
			t.Errorf("spell %d: not awake (okay %d HP %d)", sp, c.Okay, c.HP)
		}
		d := newChar("MUMMY", 1)
		d.HP = 1
		e.AddAffect(d, affMummyDisease, 0, 0, 0)
		castOn(e, sp, d)
		if d.HP != 1 {
			t.Errorf("spell %d healed the mummy-diseased", sp)
		}
	}
}

// TestNeutralizePoisonRuled: the poisons go in cure mode, the target stands.
func TestNeutralizePoisonRuled(t *testing.T) {
	c := newChar("SICK", 1)
	c.HP, c.Okay = 0, 0
	e := ruledScene(t, c)
	e.AddAffect(c, AffPoisoned, 0, 0, 0)
	e.AddAffect(c, affSlowPoison, 5, 0, 0)
	castOn(e, spNeutralizePoisonCleric, c)
	if len(c.Affects) != 0 || c.HP != 1 || c.Okay != 1 || e.Rv.Cure != 0 || !strings.Contains(said(e), "is unpoisoned") {
		t.Errorf("affects %d HP %d okay %d cure %d %q", len(c.Affects), c.HP, c.Okay, e.Rv.Cure, said(e))
	}
}

// TestLowerCaseMessagesRuled: 72, 80, 83, 89, 91 speak lower case.
func TestLowerCaseMessagesRuled(t *testing.T) {
	for sp, want := range map[SpellID]string{spProtectionEvil10Radius: "is protected", spMinorGlobeOfInvulnerability: "is protected", spFumble: "is clumsy"} {
		c := newChar("C", 1)
		e := ruledScene(t, c)
		c.Saves[SaveVsSpell] = 30 // never saves
		castOn(e, sp, c)
		if !strings.Contains(said(e), want) {
			t.Errorf("spell %d: %q, want %q", sp, said(e), want)
		}
	}
}

// TestSticksToSnakesRuled: the caster's message and pod's level cap; a big
// helpless target says nothing.
func TestSticksToSnakesRuled(t *testing.T) {
	small := newChar("SMALL", 1)
	small.Saves[SaveVsSpell] = 30
	e := ruledScene(t, small)
	e.Tables.CasterLevel.Cap = maxSpellLevelPod
	e.Party[0].Levels[ClsCleric] = 40
	castOn(e, spSticksToSnakes, small)
	a := e.FindAffect(small, affSticksToSnakes)
	if a == nil || a.Data != maxSpellLevelPod || !strings.Contains(said(e), "turns sticks into snakes") {
		t.Errorf("affect %v, %q", a, said(e))
	}
	big := newChar("BIG", 9)
	e = ruledScene(t, big)
	e.AddAffect(big, AffHeld, 0, 0, 0)
	castOn(e, spSticksToSnakes, big)
	if strings.Contains(said(e), "flat") {
		t.Error("a helpless big target smashed them")
	}
}

// TestRaiseDeadRuled: collapse and poison go, a constitution point is spent.
func TestRaiseDeadRuled(t *testing.T) {
	c := newChar("DEAD", 1)
	c.Status, c.Okay, c.HP, c.Stats[StatCon] = StatusDead, 0, 0, 10
	e := ruledScene(t, c)
	e.AddAffect(c, affAnimated, 0, 0, 0)
	e.AddAffect(c, AffPoisoned, 0, 0, 0)
	castOn(e, spRaiseDead, c)
	if c.Status != StatusOkay || c.HP != 1 || c.Stats[StatCon] != 9 || len(c.Affects) != 0 {
		t.Errorf("status %d HP %d con %d affects %d", c.Status, c.HP, c.Stats[StatCon], len(c.Affects))
	}
}

// TestSlayLivingRuled: the breathless are unaffected.
func TestSlayLivingRuled(t *testing.T) {
	c := newChar("UNDEAD", 1)
	c.MonsterFlags = monsterNoBreath
	e := ruledScene(t, c)
	seed := e.RandSeed
	castOn(e, spSlayLiving, c)
	if c.Okay == 0 || c.HP != c.MaxHP || e.RandSeed != seed || !strings.Contains(said(e), "is unaffected") {
		t.Errorf("okay %d HP %d %q", c.Okay, c.HP, said(e))
	}
}

// TestHoldTargetsRuled: Faerie Fire's first target never misses... the
// last in the list takes the free save; Hold Monsters gives big ones none.
func TestHoldTargetsRuled(t *testing.T) {
	a, b := newChar("A", 1), newChar("B", 1)
	a.Saves[SaveVsSpell], b.Saves[SaveVsSpell] = 30, 30
	e := ruledScene(t, a, b)
	castOn(e, spFaerieFire, a, b)
	if e.FindAffect(a, e.spellAffect(spFaerieFire)) == nil || e.FindAffect(b, e.spellAffect(spFaerieFire)) != nil {
		t.Errorf("Faerie Fire: a %v b %v", a.Affects, b.Affects)
	}
	big := newChar("BIG", 9)
	big.Size, big.Saves[SaveVsSpell] = 2, 30
	e = ruledScene(t, big)
	castOn(e, spHoldMonsters, big)
	if e.FindAffect(big, AffHeld) == nil {
		t.Error("Hold Monsters: a big target got a free save")
	}
}

// TestDimensionDoorRuled: a hugged caster is let go and moves.
func TestDimensionDoorRuled(t *testing.T) {
	bear := newChar("BEAR", 3)
	e := ruledScene(t, bear)
	caster := e.Party[0]
	e.AddAffect(caster, affHeldFast, 0, 0, 0)
	e.AddAffect(bear, affHuggingCotab, 0, byte(e.CombatIndex(caster)), 0)
	e.TgtX, e.TgtY = 20, 12
	castOn(e, spDimensionDoor)
	if e.FindAffect(bear, affHuggingCotab) != nil || e.CombatX(caster) != 20 || e.CombatY(caster) != 12 {
		t.Errorf("hug %v, at %d,%d", bear.Affects, e.CombatX(caster), e.CombatY(caster))
	}
}

// TestFearRuled: one Frightened affect (CotAB's RunningAway merged), the
// foe runs out of its side's hands.
func TestFearRuled(t *testing.T) {
	foe := newChar("FOE", 1)
	foe.Side, foe.Saves[SaveVsSpell] = 1, 30
	e := ruledScene(t, foe)
	e.TgtX, e.TgtY = 11, 10
	castOn(e, spFear)
	if e.FindAffect(foe, affFrightened) == nil || foe.Combat.Turned != 1 || foe.Auto != 1 || !strings.Contains(said(e), "runs in terror") {
		t.Errorf("affects %v turned %d auto %d %q", foe.Affects, foe.Combat.Turned, foe.Auto, said(e))
	}
}

// TestFireShieldRuled: the player's choice of flame, one message.
func TestFireShieldRuled(t *testing.T) {
	for choice, body := range map[int32]AffectType{fireShieldHot: affFireShieldHot, fireShieldCold: affFireShieldCold} {
		e := ruledScene(t)
		e.Fight.Choice = choice
		castOn(e, spFireShield)
		c := e.Party[0]
		if e.FindAffect(c, body) == nil || e.FindAffect(c, affZapsAttackers) == nil || strings.Count(said(e), "is protected") != 1 {
			t.Errorf("choice %d: %v %q", choice, c.Affects, said(e))
		}
	}
}

// TestRemoveCurseMageRuled: Bestow Curse (36) is lifted.
func TestRemoveCurseMageRuled(t *testing.T) {
	c := newChar("C", 1)
	e := ruledScene(t, c)
	e.AddAffect(c, AffBestowCurse, 0, 0, 0)
	castOn(e, spRemoveCurseMage, c)
	if len(c.Affects) != 0 {
		t.Errorf("affects %v", c.Affects)
	}
}

// TestFeeblemindRuled: fighters are unaffected; a feebleminded mage loses
// the spells still being memorised.
func TestFeeblemindRuled(t *testing.T) {
	f := newChar("FIGHTER", 5)
	f.Class, f.Saves[SaveVsSpell] = ClsFighter, 30
	e := ruledScene(t, f)
	castOn(e, spFeeblemind, f)
	if len(f.Affects) != 0 || !strings.Contains(said(e), "is unaffected") {
		t.Errorf("fighter: %v %q", f.Affects, said(e))
	}
	m := newChar("MAGE", 5)
	m.Class, m.Saves[SaveVsSpell] = ClsMage, 30
	m.Memorized[0], m.Memorized[1] = SpMagicMissile, SpMagicMissile|SpellPending
	e = ruledScene(t, m)
	castOn(e, spFeeblemind, m)
	if m.Memorized[0] != SpMagicMissile || m.Memorized[1] != 0 {
		t.Errorf("mage memory %v", m.Memorized[:2])
	}
}

// TestHealRuled: blindness cured, maximum less 1d4, lower case.
func TestHealRuled(t *testing.T) {
	c := newChar("C", 5)
	c.HP = 3
	e := ruledScene(t, c)
	e.AddAffect(c, AffBlind, 0, 0, 0)
	castOn(e, spHeal, c)
	if len(c.Affects) != 0 || c.HP < c.MaxHP-4 || c.HP >= c.MaxHP || !strings.Contains(said(e), "is healed") {
		t.Errorf("affects %v HP %d %q", c.Affects, c.HP, said(e))
	}
}

// TestCharmPersonOrMammalRuled: humans are charmed (pod's race test), at
// the capped level; a plain monster is not.
func TestCharmPersonOrMammalRuled(t *testing.T) {
	data := func(ranger Level) *Affect {
		h := newChar("MAN", 1)
		h.Race, h.Saves[SaveVsSpell] = RaceHuman, 30
		e := ruledScene(t, h)
		e.Tables.CasterLevel.Cap = maxSpellLevelPod
		e.Party[0].Levels[ClsRanger] = ranger // druid level ranger - 7
		castOn(e, spCharmPersonOrMammal, h)
		return e.FindAffect(h, AffCharmed)
	}
	if a, b := data(38), data(60); a == nil || b == nil || a.Data != b.Data {
		t.Errorf("human: level 31 %v, level 53 %v", a, b)
	}
	m := newChar("ORC", 1)
	e := ruledScene(t, m)
	castOn(e, spCharmPersonOrMammal, m)
	if len(m.Affects) != 0 || !strings.Contains(said(e), "is not charmed") {
		t.Errorf("monster: %v %q", m.Affects, said(e))
	}
}

// TestStoneToFleshRuled: in camp one not stoned keeps the spell, no
// question; in combat a statue stands and its body leaves the map.
func TestStoneToFleshRuled(t *testing.T) {
	c := newChar("C", 1)
	e := ruledScene(t, c)
	e.BeginFlag = 0
	castOn(e, spStoneToFlesh, c)
	if !e.spellKept || !strings.Contains(said(e), "is not stoned") {
		t.Errorf("camp: kept %v %q", e.spellKept, said(e))
	}
	s := newChar("STATUE", 1)
	e = ruledScene(t, s)
	e.combatRemoveFromMap(s)
	s.Status, s.Okay = StatusStoned, 0
	castOn(e, spStoneToFlesh, s)
	if s.Status != StatusOkay || s.HP != 1 || e.Bodies[1].C != nil || e.Cmap.Ground[10][11] != PorGroundIDs.Floor {
		t.Errorf("combat: status %d HP %d body %v size %d", s.Status, s.HP, e.Bodies[1].C, e.Cmap.Ground[10][11])
	}
}

// TestAffectRenewalRuled: the longer one wins, never two.
func TestAffectRenewalRuled(t *testing.T) {
	e := &RuleState{Tables: &POR}
	c := newChar("C", 1)
	e.AddAffectMsg(c, AffHeld, 5, 1, 0, 0, false, "")
	e.AddAffectMsg(c, AffHeld, 3, 2, 0, 0, false, "")
	if len(c.Affects) != 1 || c.Affects[0].Time != 5 {
		t.Errorf("shorter: %v", c.Affects[0])
	}
	e.AddAffectMsg(c, AffHeld, 9, 3, 0, 0, false, "")
	if len(c.Affects) != 1 || c.Affects[0].Time != 9 {
		t.Errorf("longer: %v", c.Affects[0])
	}
}
