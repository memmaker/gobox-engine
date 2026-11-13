package rules4e

import (
	"gobox/rules1e"
	"slices"
	"testing"
)

func foe(name string) *rules1e.Character {
	c := &rules1e.Character{}
	c.SetName(name)
	c.Side = 1
	c.HP = 20
	return c
}

func TestLookup(t *testing.T) {
	for name, want := range map[string]Rank{"GOBLIN": Minion, "GOBLIN LEADER": Standard, "ORC LEADER": Elite, "HUGE SCORPION": Standard} {
		if m, ok := Lookup(name); !ok || m.Rank != want {
			t.Errorf("%s: %v %v, want rank %v", name, m, ok, want)
		}
	}
	if _, ok := Lookup("BUCCANEER"); ok {
		t.Error("a human foe has no 4e kind")
	}
}

func TestMinionDiesFromAnyDamage(t *testing.T) {
	r, e := New(), &rules1e.RuleState{}
	if d := r.Damage(e, foe("GOBLIN"), 1); d < 30 {
		t.Errorf("minion took %d, want death", d)
	}
	if d := r.Damage(e, foe("GOBLIN LEADER"), 1); d != 1 {
		t.Errorf("leader took %d, want 1", d)
	}
	p := foe("GOBLIN")
	p.Side = 0 // a player who named a character Goblin
	if d := r.Damage(e, p, 1); d != 1 {
		t.Errorf("party member took %d, want 1", d)
	}
}

func TestSpecialRecharges(t *testing.T) {
	r, e := New(), &rules1e.RuleState{}
	c := foe("BASILISK")
	gaze := &rules1e.Affect{}
	gaze.Type = 0x53 // affStoneGaze
	c.Affects = []*rules1e.Affect{gaze}
	if !r.Special(e, c) || r.Special(e, c) {
		t.Fatal("special: want once, then spent")
	}
	for i := 0; i < 100 && r.spent[c]; i++ {
		r.RoundStart(e, c)
	}
	if !r.Special(e, c) {
		t.Error("special never recharged")
	}
	if !r.Special(e, foe("BUCCANEER")) {
		t.Error("a foe without a special attack is never held back")
	}
}

func member(levels map[rules1e.Class]byte) *rules1e.Character {
	c := &rules1e.Character{Combat: &rules1e.CombatInfo{}}
	c.Stats = [6]rules1e.Score{10, 10, 10, 10, 10, 10}
	for k, l := range levels {
		c.Levels[k] = rules1e.Level(l)
	}
	return c
}

func names(ms []rules1e.Maneuver) (out []string) {
	for _, m := range ms {
		out = append(out, m.Name)
	}
	return
}

func TestPicks(t *testing.T) {
	r, e := New(), &rules1e.RuleState{}
	for _, k := range []rules1e.Class{rules1e.Fighter, rules1e.Paladin, rules1e.Ranger, rules1e.Cleric, rules1e.Mage, rules1e.Thief} {
		c := member(map[rules1e.Class]byte{k: 1})
		for range 10 {
			ps := r.Picks(e, c)
			if len(ps) == 0 {
				break
			}
			c.Powers = append(c.Powers, ps[0].Options[0].ID)
		}
		if ps := r.Picks(e, c); len(ps) > 0 {
			t.Errorf("class %d still owes %v", k, ps)
		}
		if n := len(r.Maneuvers(e, c)); n < 2 && k != rules1e.Thief { // Second Wind and a pick a member without a weapon can use (the rogue's need a light blade)
			t.Errorf("class %d: %v, want Second Wind and its picks", k, names(r.Maneuvers(e, c)))
		}
	}
	h := member(map[rules1e.Class]byte{rules1e.Fighter: 3})
	h.Race = rules1e.RaceHuman
	n := 0
	for ps := r.Picks(e, h); len(ps) > 0; ps = r.Picks(e, h) {
		h.Powers, n = append(h.Powers, ps[0].Options[0].ID), n+1
	}
	want := 0
	for _, kd := range []kind{atWill, encounter} {
		k := 0
		for _, p := range powers {
			if p.kind == kd && slices.Contains(p.classes, rules1e.Fighter) {
				k++
			}
		}
		want += min(owed(kd, 3, true), k)
	}
	want++ // Fighter Weapon Talent
	if n != want || want > 6 {
		t.Errorf("a 3rd-level human fighter picked %d powers, want %d (3 at will, 2 encounter as the list has them, a Weapon Talent)", n, want)
	}
	if ms := r.Maneuvers(e, foe("ORC")); ms != nil {
		t.Errorf("a foe has maneuvers %v", names(ms))
	}
	fm := member(map[rules1e.Class]byte{rules1e.Fighter: 1, rules1e.Mage: 1})
	r.spend(fm, powerOf(1))
	if slices.Contains(names(r.Maneuvers(e, fm)), "Second Wind") {
		t.Error("Second Wind offered twice in a fight")
	}
	r.CombatStart(e)
	if !slices.Contains(names(r.Maneuvers(e, fm)), "Second Wind") {
		t.Error("Second Wind not back in the next fight")
	}
}

func TestGuardRaisesTheNumberToHit(t *testing.T) {
	r, e := New(), &rules1e.RuleState{}
	a, b := foe("ORC"), member(nil)
	r.lay(b, guard2, b, false) // Second Wind: until the start of b's next turn
	r.TurnEnd(e, b)
	if n := r.HitNeed(e, a, b, 10); n != 12 {
		t.Errorf("need %d against a guarded member, want 12", n)
	}
	r.TurnStart(e, b)
	if n := r.HitNeed(e, a, b, 10); n != 10 {
		t.Errorf("need %d after the guard's turn began, want 10", n)
	}
}

func TestRolesPickTargets(t *testing.T) {
	r, e := New(), &rules1e.RuleState{}
	tank, mage, hurt := member(map[rules1e.Class]byte{rules1e.Fighter: 1}), member(map[rules1e.Class]byte{rules1e.Mage: 1}), member(nil)
	tank.AC[0], tank.HP = 20, 30
	mage.AC[0], mage.HP = 10, 12
	hurt.AC[0], hurt.HP = 15, 2
	foes := []*rules1e.Character{tank, mage, hurt}
	for name, want := range map[string]*rules1e.Character{
		"OGRE": mage, "MINOTAUR": tank, "WIGHT": hurt, "GOBLIN LEADER": mage, "WRAITH": hurt,
	} {
		if got := r.Target(e, foe(name), foes); got != want {
			t.Errorf("%s attacked %v, want %v", name, got, want)
		}
	}
	if r.Target(e, foe("BUCCANEER"), foes) != nil {
		t.Error("a foe without a role takes the original's random pick")
	}
	if !r.Holds(e, foe("OGRE")) || !r.Holds(e, foe("GOBLIN")) || r.Holds(e, foe("WIGHT")) {
		t.Error("brutes and minions hold, skirmishers check morale")
	}
}

func TestSoldierMarks(t *testing.T) {
	r, e := New(), &rules1e.RuleState{}
	s, m, other := foe("MINOTAUR"), member(nil), member(nil)
	m.Okay, s.Okay = 1, 1
	r.AfterAttack(e, s, m, true)
	if n := r.HitNeed(e, m, other, 10); n != 12 {
		t.Errorf("marked member needs %d to hit another, want 12", n)
	}
	if n := r.HitNeed(e, m, s, 10); n != 10 {
		t.Errorf("marked member needs %d to hit the soldier, want 10", n)
	}
	r.TurnEnd(e, s) // the turn it marked in
	r.TurnStart(e, s)
	if n := r.HitNeed(e, m, other, 10); n != 12 {
		t.Errorf("the mark ended before the end of the soldier's next turn: %d", n)
	}
	r.TurnEnd(e, s)
	if n := r.HitNeed(e, m, other, 10); n != 10 {
		t.Errorf("the mark outlived the soldier's next turn: %d", n)
	}
}

func TestPowerState(t *testing.T) {
	r, e := New(), &rules1e.RuleState{}
	cl, ally, f := member(map[rules1e.Class]byte{rules1e.Cleric: 1}), member(nil), foe("ORC")
	r.lay(ally, guard1, cl, true) // Priest's Shield, in the cleric's turn
	r.lay(ally, guard2, ally, false)
	if n := r.HitNeed(e, f, ally, 10); n != 12 {
		t.Errorf("need %d against +1 and +2 guards, want 12 (they do not add up)", n)
	}
	r.TurnEnd(e, cl)
	r.TurnStart(e, ally)
	if n := r.HitNeed(e, f, ally, 10); n != 11 {
		t.Errorf("need %d after the ally's turn began, want 11", n)
	}
	r.TurnStart(e, cl)
	if n := r.HitNeed(e, f, ally, 10); n != 11 {
		t.Errorf("need %d in the cleric's next turn, want 11", n)
	}
	r.TurnEnd(e, cl)
	if n := r.HitNeed(e, f, ally, 10); n != 10 {
		t.Errorf("need %d after the cleric's next turn, want 10", n)
	}

	r.shield(ally, 3)
	r.shield(ally, 2)
	if d := r.Damage(e, ally, 5); d != 2 || r.temp[ally] != 0 {
		t.Errorf("5 damage on 3 temporary hit points: %d through, %d left", d, r.temp[ally])
	}

	rg := member(map[rules1e.Class]byte{rules1e.Ranger: 1})
	r.hitRun[rg] = f
	if r.Provokes(e, rg, f) || !r.Provokes(e, rg, foe("ORC")) {
		t.Error("Hit and Run spares only its foe's free attack")
	}
	r.TurnEnd(e, rg)
	if !r.Provokes(e, rg, f) {
		t.Error("Hit and Run outlived the ranger's turn")
	}

	r.lay(f, dazed, rg, true)
	if r.Provokes(e, rg, f) {
		t.Error("a dazed foe made a free attack")
	}
	f.Attacks = [2]rules1e.Count{1, 1}
	f.Combat = &rules1e.CombatInfo{CanCast: 1}
	r.Moved(e, f)
	if f.Attacks != [2]rules1e.Count{} || f.Combat.CanCast != 0 {
		t.Error("a dazed foe attacks after moving")
	}

	r.quarry[rg] = f
	if r.Dealt(e, rg, f, rules1e.Hit{}) == 0 {
		t.Error("no quarry damage")
	}
	if r.Dealt(e, rg, f, rules1e.Hit{}) != 0 {
		t.Error("quarry damage twice in a round")
	}
}

func TestWeaponDamage(t *testing.T) {
	r, e := New(), &rules1e.RuleState{}
	a, f := member(map[rules1e.Class]byte{rules1e.Fighter: 1}), foe("BUCCANEER") // no quarry, no minion
	a.Stats[0] = 18                                                              // Strength +4
	h := rules1e.Hit{Dice: 1, Sides: 1, Roll: 1, Bonus: 3, Ability: 2}           // 1d1, +1 magic, +2 the original's strength
	if d := r.Dealt(e, a, f, h); d != 1+1+4 {
		t.Errorf("basic: %d, want 1[W] + 1 + Strength 4", d)
	}
	if d := weaponDamage(e, a, powerOf(4), h); d != 2+1+4 { // Steel Serpent Strike
		t.Errorf("2[W]: %d, want 2 + 1 + 4", d)
	}
	if d := weaponDamage(e, a, powerOf(3), h); d != 1+1 { // Sure Strike
		t.Errorf("Sure Strike: %d, want 1[W] + 1", d)
	}
	if o := defence(e, foe("ORC"), ref); o != -4 {
		t.Errorf("orc drudge Reflex %d against AC, want -4", o)
	}
}

func TestConditions(t *testing.T) {
	r, e := New(), &rules1e.RuleState{}
	a, b, ally := foe("BUCCANEER"), member(nil), member(nil)
	h := rules1e.Hit{Roll: 7}
	r.lay(a, weakened, b, true)
	if d := r.Dealt(e, a, b, h); d != 3 {
		t.Errorf("weakened: %d, want half of 7", d)
	}
	r.lay(a, feeble, b, true) // Enfeebling Strike
	if n := r.HitNeed(e, a, b, 10); n != 12 {
		t.Errorf("enfeebled foe needs %d, want 12", n)
	}
	r.lay(ally, glow, b, true) // Divine Glow
	r.lance[a] = b             // Lance of Faith: an ally's next attack
	if n := r.HitNeed(e, ally, a, 10); n != 6 {
		t.Errorf("glowing ally with the lance needs %d, want 6", n)
	}
	r.AfterAttack(e, ally, a, false)
	if n := r.HitNeed(e, ally, a, 10); n != 8 {
		t.Errorf("the lance's +2 was not spent: need %d, want 8", n)
	}
}

func TestStyles(t *testing.T) {
	r, e := New(), &rules1e.RuleState{}
	c := member(map[rules1e.Class]byte{rules1e.Ranger: 1})
	c.MaxHP, c.HP = 10, 10
	ps := r.Picks(e, c)
	if len(ps) == 0 || len(ps[0].Options) != 2 || ps[0].Options[1].ID != twoBlade {
		t.Fatalf("the ranger's first pick %v, want the Fighting Style", ps)
	}
	c.Powers = append(c.Powers, twoBlade)
	if n := r.Toughness(c); n != 5 || !picked(c, twoBlade) {
		t.Errorf("Two-Blade: Toughness %d, want 5", n)
	}
	if slices.Contains(names(r.Maneuvers(e, c)), "Two-Blade Style") {
		t.Error("a style on the bar")
	}
}

func TestSurges(t *testing.T) {
	r, e := New(), &rules1e.RuleState{}
	c := member(map[rules1e.Class]byte{rules1e.Fighter: 1})
	c.Stats[4] = 14 // Constitution +2
	c.MaxHP, c.HP = 20, 20
	if n := surgesPerDay(e, c); n != 11 {
		t.Errorf("fighter surges %d, want 9 + 2", n)
	}
	c.Surges = 11
	if slices.Contains(names(r.Maneuvers(e, c)), "Second Wind") {
		t.Error("Second Wind without a surge left")
	}
	if surge(e, c, 0) || c.Surges != 11 {
		t.Error("spent a surge it does not have")
	}
}

func TestZoneLasts(t *testing.T) {
	r, e := New(), &rules1e.RuleState{}
	w := member(map[rules1e.Class]byte{rules1e.Mage: 1})
	r.zones = []zone{{by: w, dmg: 1}} // cast in w's turn
	r.TurnEnd(e, w)
	if len(r.zones) != 1 {
		t.Fatal("the cloud ended with the turn it was cast in")
	}
	r.TurnStart(e, w)
	r.TurnEnd(e, w)
	if len(r.zones) != 0 {
		t.Error("the cloud outlasted the end of the caster's next turn")
	}
}

// The sheet's defences are 4e's: 10 + the better ability + the class's.
func TestSheet(t *testing.T) {
	r, e := New(), &rules1e.RuleState{}
	c := member(map[rules1e.Class]byte{rules1e.Fighter: 1})
	c.AC[0] = 50 // armour class 10: no armour
	c.Stats[rules1e.Strength], c.Stats[rules1e.Constitution] = 16, 12
	got := r.Sheet(e, c)
	if len(got) != 5 || got[0].Name != "Fortitude" || got[0].Value != "15" {
		t.Errorf("sheet %v, want Fort 15 first of 5", got)
	}
}

// The reactions a member has are listed for its policies; a Wand of
// Accuracy held for a hit (OnHit) is spent only when its bonus made it.
func TestPolicies(t *testing.T) {
	r, e := New(), &rules1e.RuleState{}
	c, f := member(map[rules1e.Class]byte{rules1e.Mage: 1}), foe("GOBLIN")
	c.Powers = append(c.Powers, wandMastery)
	if rs := r.Reactions(c); len(rs) != 1 || rs[0].ID != wandMastery {
		t.Fatalf("reactions %v, want the wand", rs)
	}
	r.CombatStart(e)
	r.wandTry[c] = 2
	r.Hits(e, c, f, 2)
	if r.used[c][wandMastery] {
		t.Error("spent on a hit it did not make")
	}
	r.Hits(e, c, f, 1)
	if !r.used[c][wandMastery] {
		t.Error("not spent on the hit it made")
	}
}

// A Wand of Accuracy is not spent on a hit the Staff of Defense turns into a miss.
func TestWandAfterStaff(t *testing.T) {
	r, e := New(), &rules1e.RuleState{}
	c, b := member(map[rules1e.Class]byte{rules1e.Mage: 1}), member(map[rules1e.Class]byte{rules1e.Mage: 1})
	c.Powers = append(c.Powers, wandMastery)
	b.Side, b.Okay, b.Powers, b.Stats[rules1e.Constitution] = 1, 1, append(b.Powers, staffMastery), 16
	b.Slots[0] = &rules1e.Item{Type: 0x21} // a quarterstaff
	r.CombatStart(e)
	r.wandTry[c] = 2
	if r.Hits(e, c, b, 1) {
		t.Fatal("the staff did not ward")
	}
	if r.used[c][wandMastery] {
		t.Error("the wand spent on a hit the staff undid")
	}
}

// Evasive Strike: a player's ranged member that chose to stay is not shifted
// before the attack; the aim reaches its real shift (1 + Wisdom modifier).
func TestEvasiveStay(t *testing.T) {
	r, e := New(), &rules1e.RuleState{}
	c := member(map[rules1e.Class]byte{rules1e.Ranger: 1})
	c.Stats[rules1e.Wisdom] = 14
	r.CombatStart(e)
	powerOf(43).before(r, e, c)
	if r.shifted[c] {
		t.Error("shifted a member that stayed")
	}
	c.Powers = append(c.Powers, 43)
	for _, m := range r.Maneuvers(e, c) {
		if m.ID == 43 && m.Move != 3 {
			t.Errorf("move %d, want 3", m.Move)
		}
	}
}
