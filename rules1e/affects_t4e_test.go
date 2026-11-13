package rules1e

import (
	"gobox/core"
	"strings"
	"testing"
)

// TestOldAgeDeaths (T4e, M7): a hasted member ages; after the fight PoD's
// and SotSB's rule kill those past their race's limit, the others mourn;
// POR (no rule) only ages.
func TestOldAgeDeaths(t *testing.T) {
	for _, k := range []struct {
		rule *OldAgeRule
		dies bool
	}{{nil, false}, {&PodOldAge, true}, {&SotsbOldAge, true}} {
		tb := POR
		tb.OldAge = k.rule
		e := &RuleState{Tables: &tb}
		old, young := newChar("OLD", 1), newChar("YOUNG", 1)
		old.Race, old.Age, old.Okay = RaceHuman, 120, 1
		young.Race, young.Age, young.Okay = RaceHuman, 20, 1
		e.Party = []*Character{old, young}
		e.agesOnce(old, &Affect{}, 0)
		if old.Age != 121 || !e.ageCheckDue {
			t.Fatalf("haste: age %d, check due %v", old.Age, e.ageCheckDue)
		}
		e.oldAgeDeaths()
		if dead := old.Status == StatusGone; dead != k.dies || young.Status != StatusOkay || e.ageCheckDue {
			t.Errorf("rule %v: old status %d, young %d", k.rule, old.Status, young.Status)
		}
		if k.dies {
			last := e.Fx[len(e.Fx)-1].(core.Narrated).Text
			if last != "YOUNG weeps at the loss of such a venerable companion." {
				t.Errorf("mourning %q", last)
			}
		}
	}
}

// TestSpiritualHammerPod: PoD's hammer (its type and word) comes, and the
// item listed after it is readied when no weapon is; flag 1 takes it away.
func TestSpiritualHammerPod(t *testing.T) {
	tb := POR
	tb.Hammer = &PodHammer
	e := &RuleState{Tables: &tb}
	c := newChar("CLERIC", 1)
	e.Party = []*Character{c}
	c.Items = []*Item{{Type: 6, Words: [3]WordID{0, 6, 1}}}
	e.spiritualHammerPod(c, &Affect{}, 0)
	var hammer *Item
	for _, it := range e.Carried(c) {
		if it.Type == PodHammer.Type && it.Words[2] == PodHammer.Words[2] {
			hammer = it
		}
	}
	if hammer == nil || hammer.Spells[2] != ItemPower(AffItem0) {
		t.Fatalf("no PoD hammer: %+v", e.Carried(c))
	}
	e.spiritualHammerPod(c, &Affect{}, 1)
	for _, it := range e.Carried(c) {
		if it.Words[2] == PodHammer.Words[2] {
			t.Errorf("hammer kept after flag 1")
		}
	}
}

// TestSwallowsWholeSotsb (M5): a roll of 75 or more swallows the target, its
// body under the swallower; below, nothing.
func TestSwallowsWholeSotsb(t *testing.T) {
	for _, roll := range []int8{74, 75} {
		e := &RuleState{Tables: &POR}
		c, v := newChar("WORM", 1), newChar("HERO", 1)
		v.HP, v.Okay = 10, 1
		c.Combat, v.Combat = &CombatInfo{Target: v}, &CombatInfo{}
		e.Party = []*Character{v}
		e.Rv.HitRoll = roll
		e.swallowsWholeSotsb(c, &Affect{}, 0)
		if dead := v.Status == StatusDead && v.HP == 0; dead != (roll >= swallowRoll) {
			t.Errorf("roll %d: status %d hp %d", roll, v.Status, v.HP)
		}
	}
}

// TestMagicResistGroups: a later game's magic resistance stops an affect in
// the add-affect group; PoD's poison immune take no poison damage.
func TestMagicResistGroups(t *testing.T) {
	tb := POR
	tb.Resist = &PodMagicResist
	e := &RuleState{Tables: &tb}
	c := newChar("FOE", 1)
	c.MagicResistance = 200
	e.Rv.AffVal = AffHeld
	e.AffectGroup(c, grpAddAffect)
	if e.Rv.AffVal != 0 {
		t.Errorf("affect %x not resisted", e.Rv.AffVal)
	}
	c.MagicResistance, c.MonsterFlags = 0, monsterPoisonImmune
	e.Rv.Damage, e.Rv.DmgType = 9, dmgPoison
	e.AffectGroup(c, grpDamage)
	if e.Rv.Damage != 0 {
		t.Errorf("poison immune took %d", e.Rv.Damage)
	}
}

// TestNoLevelDrain (Felix: no game drains levels): no handler, variant or
// spell is a level drain; the former drain touches weaken and Energy Drain
// holds. Detects the class: any new drain handler or spell id fails here.
func TestNoLevelDrain(t *testing.T) {
	isDrain := func(name string) bool {
		n := strings.ToLower(name)
		return strings.Contains(n, "drain") && !strings.Contains(n, "blooddrain")
	}
	for _, hs := range []map[HandlerName]AffectHandler{porHandlers, cotabHandlers, sotsbHandlers, podHandlers, aimHandlers} {
		for n := range hs {
			if isDrain(string(n)) {
				t.Errorf("drain handler %s", n)
			}
		}
	}
	for _, vs := range []map[AffectType]HandlerName{CotabAffectVariants, SotsbAffectVariants, PodAffectVariants} {
		for id, n := range vs {
			if isDrain(string(n)) {
				t.Errorf("affect %x: drain variant %s", id, n)
			}
		}
	}
	for id, s := range SpellIDs {
		if isDrain(string(s)) {
			t.Errorf("spell %d is %s", id, s)
		}
	}
	if PodAffectIDs[101] != affWeakeningTouch || PodAffectIDs[102] != affWeakeningTouch2 || PodSpellsGone[103] != 23 || PodSpellsGone[102] != 0 {
		t.Errorf("PoD's former drains: %x %x %d", PodAffectIDs[101], PodAffectIDs[102], PodSpellsGone[103])
	}
}

// TestStatItemsCotab (M3): a readied CotAB stat item changes the score in
// use from the character's own; put away, the own score comes back.
func TestStatItemsCotab(t *testing.T) {
	e := &RuleState{Tables: &POR}
	c := newChar("ROGUE", 1)
	c.Stats[StatDex], c.Base[StatDex] = 12, 12
	it := &Item{Readied: 1}
	it.Spells[2] = ItemPower(itemPowerDexterityDown)
	c.Items = []*Item{it}
	e.dexterityDownItemCotab(c, &Affect{Item: it}, 0)
	if c.Stats[StatDex] != 10 {
		t.Errorf("readied: dex %d, want 10", c.Stats[StatDex])
	}
	it.Readied = 0
	e.dexterityDownItemCotab(c, &Affect{Item: it}, 1)
	if c.Stats[StatDex] != 12 {
		t.Errorf("put away: dex %d, want 12", c.Stats[StatDex])
	}
	it.Readied, it.Spells[2] = 1, ItemPower(itemPowerLowStrengthMind)
	c.Stats[StatStr], c.Base[StatStr], c.Stats[StatInt], c.Base[StatInt] = 16, 16, 14, 14
	e.lowStrengthMindItemCotab(c, &Affect{Item: it}, 0)
	if c.Stats[StatStr] != lowScore || c.Stats[StatInt] != lowScore {
		t.Errorf("low strength and mind: %d %d", c.Stats[StatStr], c.Stats[StatInt])
	}
}
