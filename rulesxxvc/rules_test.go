package rulesxxvc

import (
	"testing"

	"gobox/rules"
	"gobox/rules/xxvc"
)

func member(name string, dex, tech int, firstAid int8, coins rules.Copper) Character {
	c := Character{Name: name, Career: xxvc.Warrior, Enabled: true, Status: StOkay, Coins: coins,
		Scores: map[rules.Ability]int{rules.Dexterity: dex, xxvc.Tech: tech}}
	c.Skills.Set(13, firstAid)
	return c
}

// Felix 2026-11-11: off per member; on pools; off again: all to the first.
func TestPurse(t *testing.T) {
	r := New()
	a, b := r.Add(member("a", 10, 0, 0, 5)), r.Add(member("b", 10, 0, 0, 7))
	if r.Purse(a) != 5 || r.Purse(b) != 7 {
		t.Fatal("off: own coins")
	}
	r.Begin(rules.Campaign{}, rules.Options{SharedPurse: true})
	if r.Purse(a) != 12 || r.Purse(b) != 12 {
		t.Fatal("on: one purse")
	}
	r.Begin(rules.Campaign{}, rules.Options{})
	if r.Purse(a) != 12 || r.Purse(b) != 0 {
		t.Fatalf("off again: %d %d", r.Purse(a), r.Purse(b))
	}
	saved, err := r.Save()
	if err != nil {
		t.Fatal(err)
	}
	q := New()
	if err := q.Load(saved); err != nil || q.Purse(a) != 12 {
		t.Fatal("Load", err)
	}
}

// Seed 1 d100s 57, 30 (outside Go): a 20+10-57 = -27, b 40+5-30 = 15 wins;
// the band is the last candidate's 45: 15 < 22, Made.
func TestCheckSkillRoll(t *testing.T) {
	r := New()
	a, b := r.Add(member("a", 10, 10, 20, 0)), r.Add(member("b", 10, 5, 40, 0))
	rng := rules.RNG{State: 1}
	got := r.Check(rules.SkillRoll{Whom: []rules.CharID{a, b}, Skill: "first-aid", Hardness: rules.Normal}, &rng)
	if got.Grade != rules.Made || got.Who != b || !got.Yes {
		t.Errorf("%+v", got)
	}
}

// Seed 1: a d10 7, d127 42; b d10 7, d127 107: both 8, b's tie-breaker wins.
func TestBeginRound(t *testing.T) {
	r := New()
	a, b := r.Add(member("a", 10, 0, 0, 0)), r.Add(member("b", 10, 0, 0, 0))
	rng := rules.RNG{State: 1}
	f := rules.Fight{Party: rules.Party{a, b}}
	r.BeginRound(f, &rng)
	for _, want := range []rules.CharID{b, a} {
		if id, ok := r.NextActor(f, &rng); !ok || id != want {
			t.Fatalf("NextActor %d, want %d", id, want)
		}
	}
	if _, ok := r.NextActor(f, &rng); ok {
		t.Error("round not over")
	}
}
