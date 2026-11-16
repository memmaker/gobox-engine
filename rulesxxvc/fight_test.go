package rulesxxvc

import (
	"image"
	"testing"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// gametime.pas RestAndHeal: unconscious wake, okay gain a HP a day to the max.
func TestRestAndHeal(t *testing.T) {
	r := New()
	a := r.Add(Character{Status: StUnconscious, HP: 3, MaxHP: 10})
	b := r.Add(Character{Status: StOkay, HP: 9, MaxHP: 10})
	d := r.Add(Character{Status: StDead, HP: 0, MaxHP: 10})
	o, err := r.Resolve(rules.Camping{Party: rules.Party{a, b, d}}, rules.Rest{Minutes: 2 * MinutesPerDay}, nil)
	if err != nil || r.chars[a].HP != 5 || r.chars[a].Status != StOkay || !r.chars[a].Enabled ||
		r.chars[b].HP != 10 || r.chars[d].HP != 0 || len(o.Effects) != 2 {
		t.Errorf("%v %+v %+v %+v", err, r.chars[a], r.chars[b], r.chars[d])
	}
}

// The party surprised: each initiative 6 less (Initiative), drawn after
// RollTactics' two d100s (both untrained, both fail, neither stops the side).
func TestSurprise(t *testing.T) {
	r := New()
	a, b := r.Add(member("a", 10, 0, 0, 0)), r.Add(member("b", 10, 0, 0, 0))
	rng := rules.RNG{State: 1}
	want := rng
	want.N(100)
	want.N(100)
	ia, _ := Initiative(&want, true, 10, true)
	ib, _ := Initiative(&want, true, 10, true)
	r.BeginRound(rules.Fight{Party: rules.Party{a, b}, Terms: rules.FightTerms{Surprised: []rules.Side{rules.SideParty}}}, &rng)
	if r.fight[a].initiative != ia || r.fight[b].initiative != ib || r.tactics != [2]bool{} {
		t.Errorf("%+v %+v", r.fight[a], r.fight[b])
	}
}

// RollZeroG: no Maneuver in 0g is value 0, which always misses by more than
// its band of 0: failed.
func TestZeroG(t *testing.T) {
	r := New()
	a := r.Add(member("a", 10, 0, 0, 0))
	rng := rules.RNG{State: 1}
	f := rules.Fight{Party: rules.Party{a}, Board: rules.Board{ZeroG: true}}
	r.BeginFight(f, &rng)
	r.BeginRound(f, &rng)
	if turn := r.BeginTurn(f, a, &rng); turn.Skipped || r.chars[a].Status != StFailedZeroG {
		t.Errorf("%+v %v", turn, r.chars[a].Status)
	}
}

// DetermineLootExp: 100 xp over two enabled members, the credits taken.
func TestEndFight(t *testing.T) {
	r := New()
	a, b := r.Add(member("a", 10, 0, 0, 0)), r.Add(member("b", 10, 0, 0, 0))
	m := r.Add(Character{Side: rules.SideMonsters, Status: StDead, XPAward: 100, Coins: 30})
	f := rules.Fight{Party: rules.Party{a, b}, Board: rules.Board{Units: []rules.BoardUnit{{ID: m}}}}
	if v := r.Verdict(f); v != rules.PartyWon {
		t.Fatal(v)
	}
	res := r.EndFight(f, nil)
	if res.Exp != 50 || res.Money.Copper != 30 || res.Beaten != 1 || r.chars[a].Exp != 50 {
		t.Errorf("%+v", res)
	}
	r.chars[m].Enabled, r.chars[m].Status = true, StOkay
	r.chars[a].Enabled, r.chars[b].Enabled = false, false
	r.chars[b].Status = StFleeing
	if v := r.Verdict(f); v != rules.PartyFled {
		t.Error(v)
	}
}

func TestFactsAndExport(t *testing.T) {
	r := New()
	a := r.Add(member("a", 12, 0, 7, 9))
	if v, err := r.Fact(a, rules.SkillFact{Skill: "first-aid"}); err != nil || v != rules.Number(7) {
		t.Error(v, err)
	}
	if v, _ := r.Fact(a, rules.KnownFact{Skill: "first-aid"}); v != rules.Yes(true) {
		t.Error(v)
	}
	if err := r.SetFact(a, rules.FactHP, rules.Number(4)); err != nil || r.chars[a].HP != 4 {
		t.Error(err)
	}
	if _, err := r.Fact(a, rules.FactPaladin); err == nil {
		t.Error("paladin")
	}
	f, err := r.Export(a)
	if err != nil {
		t.Fatal(err)
	}
	b, err := r.Import(f)
	if err != nil || r.Sheet(b).HP != 4 || r.Purse(b) != 9 {
		t.Error(err, r.Sheet(b))
	}
}

// charhlp.pas InflictDamage: overkill 30 dead, 10 comatose, 1 dying, 0 unconscious.
func TestInflictDamage(t *testing.T) {
	for _, c := range []struct {
		hp, dmg int
		want    Status
		down    bool
	}{{5, 35, StDead, true}, {5, 15, StComatose, true}, {5, 6, StDying, true}, {5, 5, StUnconscious, true}, {5, 4, StOkay, false}} {
		ch := Character{HP: c.hp, Status: StOkay, Enabled: true}
		if InflictDamage(&ch, c.dmg) != c.down || ch.Status != c.want || ch.Enabled == c.down {
			t.Errorf("%d-%d: %+v", c.hp, c.dmg, ch)
		}
	}
}

// One attack of a phase-2 rate 2 at bonus 100: hits unless the die is 1; the
// 1d4 is the next draw.
func TestAttackAndFlee(t *testing.T) {
	r := New()
	a := member("a", 10, 0, 0, 0)
	a.AttackBonus, a.Scale, a.Damage[1], a.Movement = 100, 2, rules.Dice{N: 1, Sides: 4}, 12
	a.BaseMovement = 12
	aid := r.Add(a)
	m := r.Add(Character{Side: rules.SideMonsters, Status: StOkay, Enabled: true, HP: 50, BaseMovement: 6})
	f := rules.Fight{Party: rules.Party{aid}, Board: rules.Board{Units: []rules.BoardUnit{{ID: m}}}, Actor: aid}
	rng := rules.RNG{State: 1}
	r.BeginFight(f, &rng)
	r.BeginRound(f, &rng)
	want := rng
	die := want.N(20) + 1
	dmg := 0
	if die != 1 {
		dmg = want.N(4) + 1
	}
	if _, err := r.Resolve(f, rules.Attack{Target: m}, &rng); err != nil || r.chars[m].HP != 50-dmg {
		t.Errorf("%v hp %d, die %d dmg %d", err, r.chars[m].HP, die, dmg)
	}
	if o, _ := r.Resolve(f, rules.Flee{}, &rng); o.Confirm != "Flee:" || r.chars[aid].Status == StFleeing {
		t.Errorf("Flee asks first: %+v", o)
	}
	if o, _ := r.Resolve(f, rules.Confirmed{Action: rules.Flee{}}, &rng); r.chars[aid].Status != StFleeing || len(o.Effects) != 2 {
		t.Errorf("flee: 6 < 3? %+v", o)
	}
}

// CombatAiControl: a monster three squares east steps west toward its one
// target, then, adjacent, attacks it and ends its turn.
func TestDecide(t *testing.T) {
	r := New()
	a := r.Add(member("a", 10, 0, 0, 0))
	m := r.Add(Character{Side: rules.SideMonsters, Enabled: true, HP: 5, MaxHP: 5, BaseMovement: 6, Scale: 2, BaseScale: 2})
	f := rules.Fight{Party: rules.Party{a}, Board: rules.Board{Size: image.Pt(10, 10),
		Units: []rules.BoardUnit{{ID: a, At: rules.BoardPos{X: 1, Y: 5}}, {ID: m, At: rules.BoardPos{X: 3, Y: 5}}}}}
	rng := rules.RNG{State: 1}
	r.BeginFight(f, &rng)
	r.BeginRound(f, &rng)
	r.fight[m].initiative = 5
	r.BeginTurn(f, m, &rng)
	f.Actor = m
	act := r.Decide(f, m, &rng)
	mv, ok := act.(rules.Move)
	if !ok || distance(mv.To, rules.BoardPos{X: 1, Y: 5}) != 1 {
		t.Fatalf("%#v", act)
	}
	if _, err := r.Resolve(f, mv, &rng); err != nil || r.fight[m].movement >= 12 {
		t.Fatal(err, r.fight[m].movement)
	}
	f.Board.Units[1].At = mv.To
	if act := r.Decide(f, m, &rng); act != (rules.Attack{Target: a}) {
		t.Fatalf("%#v", act)
	} else if _, err := r.Resolve(f, act, &rng); err != nil {
		t.Fatal(err)
	}
	// a down with attacks left: the next pass finds no target and guards (FinishRound)
	if act := r.Decide(f, m, &rng); r.chars[a].Enabled || r.fight[m].initiative == 0 || act != (rules.Guard{}) {
		t.Fatalf("%#v %v", act, r.chars[a].Enabled)
	}
}

// IsRearAttack: a second attack from the square the defender turns its back
// on strikes its rear armour (0 here; its front cannot be hit), and with no
// attacks left the actions end.
func TestRearAttack(t *testing.T) {
	r := New()
	a := r.Add(member("a", 10, 0, 0, 0))
	m := r.Add(Character{Side: rules.SideMonsters, Enabled: true, HP: 50, MaxHP: 50, ArmorClass: [2]int{200, 0}})
	f := rules.Fight{Party: rules.Party{a}, Actor: a, Board: rules.Board{Size: image.Pt(10, 10), Units: []rules.BoardUnit{
		{ID: a, At: rules.BoardPos{X: 1, Y: 5}}, {ID: m, At: rules.BoardPos{X: 2, Y: 5}, Facing: 2}}}}
	rng := rules.RNG{State: 1}
	r.BeginFight(f, &rng)
	r.BeginRound(f, &rng)
	r.fight[m].attacksDrawn = 1
	r.fight[a].attacks = [2]int{0, 3}
	r.chars[a].Damage[1] = rules.Dice{N: 1, Sides: 1}
	o, err := r.Resolve(f, rules.Attack{Target: m}, &rng)
	if err != nil || r.chars[m].HP >= 50 || r.fight[a].initiative != 0 {
		t.Errorf("%v %+v %d", err, o, r.chars[m].HP)
	}
}

// MoveMenu's reach: movement 4 pays two straight steps or one diagonal.
func TestReach(t *testing.T) {
	r := New()
	f := rules.Fight{Board: rules.Board{Size: image.Pt(3, 3)}}
	rs := r.reach(f, rules.BoardPos{}, 4)
	if len(rs) != 5 || pathCost(f.Board, rules.BoardPos{}, rs[1].Path) != 4 {
		t.Errorf("%+v", rs)
	}
}

// DetermineLootExp's items (NotTreasure stays) and revivesParty.
func TestEndFightLootRevive(t *testing.T) {
	r := New()
	r.tables = xxvc.Tables{Items: map[rules.ItemID]xxvc.ItemTemplate{1: {}, 2: {NotTreasure: true}}}
	a := r.Add(member("a", 10, 0, 0, 0))
	b := r.Add(Character{Status: StDying})
	m := r.Add(Character{Side: rules.SideMonsters, Status: StDead, Items: []Carried{{Slot: 1, Ref: rules.ItemRef{ID: 1}}, {Slot: 2, Ref: rules.ItemRef{ID: 2}}}})
	f := rules.Fight{Party: rules.Party{a, b}, Terms: rules.FightTerms{RevivesParty: true}, Board: rules.Board{Units: []rules.BoardUnit{{ID: m}}}}
	res := r.EndFight(f, nil)
	if len(res.Items) != 1 || r.loot[res.Items[0]].ID != 1 || r.chars[b].HP != 1 || r.chars[b].Status != StOkay {
		t.Errorf("%+v %+v", res, r.chars[b])
	}
}

// CombatMenu offers Use Jetpack to a member with a readied jetpack and
// movement left only in a wilderness area (areaKind Venus/Mars wilderness).
func TestJetpackOffer(t *testing.T) {
	r := New()
	r.tables = xxvc.Tables{ItemTypes: []xxvc.ItemType{{Slot: xxvc.SlotJetpack}}, Items: map[rules.ItemID]xxvc.ItemTemplate{1: {}}}
	a := member("a", 10, 0, 0, 0)
	a.BaseMovement, a.Items = 12, []Carried{{Slot: 1, Ref: rules.ItemRef{ID: 1}, Readied: true}}
	aid := r.Add(a)
	f := rules.Fight{Party: rules.Party{aid}, Actor: aid, Board: rules.Board{Size: image.Pt(3, 3),
		Units: []rules.BoardUnit{{ID: aid, At: rules.BoardPos{X: 1, Y: 1}}}}}
	rng := rules.RNG{State: 1}
	r.BeginFight(f, &rng)
	r.BeginRound(f, &rng)
	jets := func() (n int) {
		for _, x := range r.allowedFight(f).Aimed {
			if _, ok := x.Action.(rules.Jetpack); ok {
				n++
			}
		}
		return n
	}
	if n := jets(); n != 0 {
		t.Errorf("plain area: %d jetpack squares", n)
	}
	f.Terms.Area = rules.AreaWilderness
	if n := jets(); n != 8 {
		t.Errorf("wilderness: %d jetpack squares, want the 8 free ones", n)
	}
}

// RollTactics: Battle Tactics 120 always makes the roll (120 - d100 >= 0);
// the party side then hits at +1 from its next RecalcStats; the second
// member does not roll for a side that has it.
func TestRollTactics(t *testing.T) {
	r := New()
	ca := member("a", 10, 0, 0, 0)
	ca.Skills.Set(BattleTacticsNo, 120)
	a, b := r.Add(ca), r.Add(member("b", 10, 0, 0, 0))
	rng := rules.RNG{State: 1}
	o := r.BeginRound(rules.Fight{Party: rules.Party{a, b}}, &rng)
	won := r.tactics
	r.tactics = [2]bool{}
	r.RecalcStats(r.chars[a])
	before := r.chars[a].AttackBonus
	r.tactics = won
	r.RecalcStats(r.chars[a])
	if !r.tactics[0] || r.tactics[1] || len(o.Effects) != 1 || r.chars[a].AttackBonus != before+1 {
		t.Fatalf("%v %+v %d %d", r.tactics, o, before, r.chars[a].AttackBonus)
	}
}

// RecalcMovement: the zone's movementBonus (PartyMove 2) adds to the team's
// base 6 only, doubled: 16 half steps; a monster keeps 12; EndFight clears it.
func TestMovementBonus(t *testing.T) {
	r := New()
	ca := member("a", 10, 0, 0, 0)
	ca.BaseMovement = 6
	a := r.Add(ca)
	m := r.Add(Character{Side: rules.SideMonsters, Enabled: true, BaseMovement: 6})
	f := rules.Fight{Party: rules.Party{a}, Terms: rules.FightTerms{PartyMove: 2}}
	r.BeginFight(f, nil)
	if r.RecalcMovement(r.chars[a]) != 16 || r.RecalcMovement(r.chars[m]) != 12 {
		t.Fatal(r.RecalcMovement(r.chars[a]), r.RecalcMovement(r.chars[m]))
	}
	r.EndFight(f, nil)
	if r.RecalcMovement(r.chars[a]) != 12 {
		t.Fatal(r.RecalcMovement(r.chars[a]))
	}
}

// AiMoveAttack's flight: a fleeing combatant with movement left but less
// than a whole step does not fight on; AiMoveStep's gate fails and
// FinishRound ends its round (guarding), even with an enemy adjacent.
func TestFleeHalfStep(t *testing.T) {
	r := New()
	a := r.Add(member("a", 10, 0, 0, 0))
	m := r.Add(Character{Side: rules.SideMonsters, Enabled: true, HP: 5, MaxHP: 5, BaseMovement: 6, Scale: 2, BaseScale: 2})
	f := rules.Fight{Party: rules.Party{a}, Actor: m, Board: rules.Board{Size: image.Pt(10, 10),
		Units: []rules.BoardUnit{{ID: a, At: rules.BoardPos{X: 1, Y: 5}}, {ID: m, At: rules.BoardPos{X: 2, Y: 5}}}}}
	rng := rules.RNG{State: 1}
	r.BeginFight(f, &rng)
	r.BeginRound(f, &rng)
	r.fight[m].initiative = 5
	r.BeginTurn(f, m, &rng)
	s := r.fight[m]
	s.fleeing, s.movement, s.aiBegun, s.aiPrepare, s.target = true, 1, true, false, a
	if act := r.Decide(f, m, &rng); act != (rules.Guard{}) {
		t.Fatalf("%#v", act)
	}
}

// HitFact is what an attack uses: for each seed the attack's roll hits exactly when HitFact counts it, with and
// without the fight's PartyHit bonus (BeginFight keeps it), and the fact changes nothing (a blind defender's armour).
func TestHitFactIsTheAttacks(t *testing.T) {
	var facts []int
	for _, bonus := range []int32{0, 3, -4} {
		hitsBy := map[int]bool{} // attack roll -> the attack hit
		var fact int
		for seed := uint32(1); seed <= 300; seed++ {
			r := New()
			a := member("a", 10, 0, 0, 0)
			a.AttackBonus, a.Scale, a.Damage[1] = 3, 2, rules.Dice{N: 1, Sides: 4}
			aid := r.Add(a)
			m := r.Add(Character{Side: rules.SideMonsters, Status: StOkay, Enabled: true, HP: 500, BaseMovement: 6, BaseArmorClass: 17})
			r.chars[m].Effects = []Effect{{ID: EffBlind}}
			f := rules.Fight{Party: rules.Party{aid}, Actor: aid, Terms: rules.FightTerms{PartyHit: bonus},
				Board: rules.Board{Units: []rules.BoardUnit{{ID: aid, At: rules.BoardPos{X: 5, Y: 5}, Size: 1}, {ID: m, At: rules.BoardPos{X: 6, Y: 5}, Size: 1}}}}
			rng := rules.RNG{State: seed}
			r.BeginFight(f, &rng)
			r.BeginRound(f, &rng)
			r.chars[aid].AttackBonus = 3
			before := r.chars[m].ArmorClass
			v, err := r.Fact(aid, rules.HitFact{Target: m})
			if err != nil || r.chars[m].ArmorClass != before {
				t.Fatalf("fact %v: %v, armor %v then %v", v, err, before, r.chars[m].ArmorClass)
			}
			fact = int(v.(rules.Number))
			die := rng
			roll, missed := AttackDie(&die)
			o, err := r.Resolve(f, rules.Attack{Target: m}, &rng)
			if err != nil {
				t.Fatal(err)
			}
			if missed {
				continue
			}
			for _, e := range o.Effects {
				if at, ok := e.(rules.Attacked); ok {
					hitsBy[roll] = at.Hit
				}
			}
		}
		want := 0
		for roll := 2; roll <= naturalAttackMax; roll++ {
			if hit, seen := hitsBy[roll]; seen && hit {
				want += 5
			}
		}
		if len(hitsBy) < 19 || fact != want {
			t.Errorf("PartyHit %d: HitFact %d, the attack hits on %d%% (%d rolls seen)", bonus, fact, want, len(hitsBy))
		}
		facts = append(facts, fact)
	}
	if facts[1] <= facts[0] || facts[0] <= facts[2] { // the bonus counts
		t.Errorf("HitFact for PartyHit 0, 3, -4: %v", facts)
	}
}
