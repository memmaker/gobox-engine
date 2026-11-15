package rulesxxvc

import (
	"image"
	"testing"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// TimeOutEffects: 25 minutes in steps 10, 10, 5: a 20-minute effect runs out
// on the second step, a 30 one has 5 left, a lasting one (0) stays.
func TestElapse(t *testing.T) {
	r := New()
	a := r.Add(Character{Effects: []Effect{{ID: EffStunned, Minutes: 20}, {ID: EffDizzy, Minutes: 30}, {ID: EffHelpless}}})
	r.Elapse(rules.Party{a}, 25, nil)
	want := []Effect{{ID: EffDizzy, Minutes: 5}, {ID: EffHelpless}}
	if got := r.chars[a].Effects; len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("%+v", got)
	}
}

// effhndlr.pas by hand: the immunities clear a hit of their class, a melee
// hit on a laser-immune defender stands, UNKILLABLE caps at the hit points,
// DODGE_SLUG_HITS is the first 50% roll only beyond one square; unarmed no
// immunity applies.
func TestDefenderDamage(t *testing.T) {
	d := &Character{HP: 7, Effects: []Effect{{ID: EffImmuneToLaser}, {ID: EffImmuneToHeatAndPlasma}, {ID: EffUnkillable}}}
	for _, c := range []struct {
		wc    xxvc.WeaponClass
		armed bool
		want  int
	}{{xxvc.Laser, true, 0}, {xxvc.Heat, true, 0}, {xxvc.Plasma, true, 0}, {xxvc.Melee, true, 7}, {xxvc.Laser, false, 7}, {xxvc.Slug, true, 7}} {
		if got, _ := defenderDamage(d, 20, c.armed, c.wc, 1, nil); got != c.want {
			t.Errorf("class %d armed %v: %d, want %d", c.wc, c.armed, got, c.want)
		}
	}
	d = &Character{HP: 30, Effects: []Effect{{ID: EffDodgeSlugHits}}}
	if got, dodged := defenderDamage(d, 9, true, xxvc.Slug, 1, nil); got != 9 || dodged {
		t.Fatal("dodged at one square")
	}
	rng, want := rules.RNG{State: 1}, rules.RNG{State: 1}
	dodge := PercentileRoll(&want, 50)
	got, dodged := defenderDamage(d, 9, true, xxvc.Slug, 3, &rng)
	if dodged != dodge || dodged != (got == 0) || rng != want {
		t.Fatalf("dodge %v want %v, damage %d", dodged, dodge, got)
	}
}

// AddEffectWithMessage: a save cancels, UNKILLABLE and BLOCK_SCARED block,
// a held effect only lengthens. TURN_START: STUNNED ends the actions and
// SCARED goes.
func TestAddEffectAndTurnStart(t *testing.T) {
	r := New()
	src := &Character{}
	c := &Character{Enabled: true, Effects: []Effect{{ID: EffStunned, Minutes: 3}}}
	if r.addEffect(c, src, Effect{ID: EffScared, Minutes: 5}, true) || len(c.Effects) != 1 {
		t.Fatal("a made save did not cancel")
	}
	if !r.addEffect(c, src, Effect{ID: EffStunned, Minutes: 8}, false) || len(c.Effects) != 1 || c.Effects[0].Minutes != 8 {
		t.Fatalf("held effect: %+v", c.Effects)
	}
	b := &Character{Effects: []Effect{{ID: EffBlockScared}}}
	if r.addEffect(b, src, Effect{ID: EffScared}, false) || len(b.Effects) != 1 {
		t.Fatal("BLOCK_SCARED")
	}
	r.addEffect(c, src, Effect{ID: EffScared, Minutes: 5}, false)
	id := r.Add(*c)
	f := rules.Fight{Round: 1}
	r.fight = map[rules.CharID]*combatState{id: {initiative: 5}}
	rng := rules.RNG{State: 1}
	if turn := r.BeginTurn(f, id, &rng); !turn.Skipped || hasEffect(r.chars[id], EffScared) {
		t.Fatalf("turn %+v effects %+v", turn, r.chars[id].Effects)
	}
}

// MONSTER_SPECIAL: a STRIKE_ADJACENT monster next to its target gets the
// Maneuver first; it deals 8 on a failed save 6, else 4, once.
func TestMonsterSpecial(t *testing.T) {
	r := New()
	a := r.Add(member("a", 10, 0, 0, 0))
	r.chars[a].HP, r.chars[a].MaxHP = 20, 20
	m := r.Add(Character{Side: rules.SideMonsters, Enabled: true, HP: 5, MaxHP: 5, BaseMovement: 6, Scale: 2, BaseScale: 2, Effects: []Effect{{ID: EffStrikeAdjacent}}})
	f := rules.Fight{Party: rules.Party{a}, Board: rules.Board{Size: image.Pt(10, 10),
		Units: []rules.BoardUnit{{ID: a, At: rules.BoardPos{X: 1, Y: 5}}, {ID: m, At: rules.BoardPos{X: 2, Y: 5}}}}}
	rng := rules.RNG{State: 1}
	r.BeginFight(f, &rng)
	r.BeginRound(f, &rng)
	r.fight[m].initiative = 5
	r.BeginTurn(f, m, &rng)
	f.Actor = m
	act := r.Decide(f, m, &rng)
	if mv, ok := act.(rules.Maneuver); !ok || mv.ID != monsterSpecial {
		t.Fatalf("%#v", act)
	}
	want := rng
	dmg := 8
	if rollSave(r.chars[a], 6, 0, &want) {
		dmg = 4
	}
	if _, err := r.Resolve(f, act, &rng); err != nil || r.chars[a].HP != 20-dmg || rng != want {
		t.Fatal(err, r.chars[a].HP)
	}
	if _, err := r.Resolve(f, act, &rng); err == nil {
		t.Fatal("a second special resolved")
	}
	if act := r.Decide(f, m, &rng); act != (rules.Attack{Target: a}) {
		t.Fatalf("%#v", act)
	}
}

// TestOnRemove: an imported effect that calls on remove ends the actions when
// it goes (EffHelpless, EffStunned); one that does not, or EffScared, leaves them.
func TestOnRemove(t *testing.T) {
	for _, tc := range []struct {
		e    Effect
		ends bool
	}{
		{Effect{ID: EffStunned, CallsOnRemove: true}, true},
		{Effect{ID: 7, CallsOnRemove: true}, true},
		{Effect{ID: EffStunned}, false},
		{Effect{ID: EffScared, CallsOnRemove: true}, false},
	} {
		c := &Character{Effects: []Effect{tc.e}}
		s := &combatState{initiative: 5, movement: 3}
		removeEffect(c, s, tc.e.ID)
		if len(c.Effects) != 0 || (s.initiative == 0) != tc.ends {
			t.Errorf("%+v: effects %v, state %+v", tc.e, c.Effects, s)
		}
	}
	c := &Character{Effects: []Effect{{ID: EffHelpless, Minutes: 5, CallsOnRemove: true}}}
	s := &combatState{initiative: 5}
	timeOutEffects(c, s, 10)
	if len(c.Effects) != 0 || s.initiative != 0 {
		t.Errorf("time out: effects %v, state %+v", c.Effects, s)
	}
}
