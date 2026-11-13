package engine

import (
	"gobox/game"
	"gobox/game/play"
	"gobox/rules1e"
	"slices"
	"testing"
)

type toughRules struct{ rules1e.ClassicRules }

func (toughRules) Toughness(*rules1e.Character) int { return 5 }

// Toughness comes with the rule set and goes with it.
func TestToughen(t *testing.T) {
	e, c := &Engine{}, &rules1e.Character{}
	c.MaxHP, c.HP = 10, 3
	e.state().SetRules(toughRules{})
	e.state().Toughen(c)
	if c.MaxHP != 15 || c.HP != 8 || c.Tough != 5 {
		t.Fatalf("on: %d/%d tough %d", c.HP, c.MaxHP, c.Tough)
	}
	e.state().SetRules(nil)
	e.state().Toughen(c)
	e.state().Toughen(c)
	if c.MaxHP != 10 || c.HP != 3 || c.Tough != 0 {
		t.Errorf("off: %d/%d tough %d", c.HP, c.MaxHP, c.Tough)
	}
}

// pickRules owes two at-wills, picked from three.
type pickRules struct{ rules1e.ClassicRules }

func (pickRules) Picks(_ *rules1e.RuleState, c *rules1e.Character) []rules1e.PowerOwed {
	var free []rules1e.Maneuver
	for id := 1; id <= 3; id++ {
		if !slices.Contains(c.Powers, id) {
			free = append(free, rules1e.Maneuver{ID: id})
		}
	}
	if len(c.Powers) >= 2 {
		return nil
	}
	return []rules1e.PowerOwed{{What: "an at-will power", Options: free}}
}

// The creation's power picks: made ones keep their slot, a change
// replaces the power in it, a power held elsewhere does not fit.
func TestPowerSlots(t *testing.T) {
	e, c := &Engine{}, &rules1e.Character{}
	e.state().SetRules(pickRules{})
	if ss := e.state().PowerSlots(c); len(ss) != 1 || ss[0].Pick != -1 {
		t.Fatalf("fresh: %v", ss)
	}
	e.state().SetPower(c, 0, 1) // id 2
	e.state().SetPower(c, 1, 0) // id 1 (3 and 1 left)
	if ss := e.state().PowerSlots(c); len(ss) != 2 || !slices.Equal(c.Powers, []int{2, 1}) {
		t.Fatalf("picked: %v %v", ss, c.Powers)
	}
	e.state().SetPower(c, 1, 0) // slot 1 holds 1, its options 1 and 3: 1 again, no change
	e.state().SetPower(c, 0, 0) // slot 0's options 1,2,3: 1 is held by slot 1
	if !slices.Equal(c.Powers, []int{2, 1}) {
		t.Errorf("held elsewhere: %v", c.Powers)
	}
	e.state().SetPower(c, 0, 2) // 3 in place of 2
	if !slices.Equal(c.Powers, []int{3, 1}) {
		t.Errorf("replaced: %v", c.Powers)
	}
}

// Cast out of combat (a card's or the sheet's): the pick list comes back
// as the question to put (pending data), each spell making the call again with it.
func TestCastAsksSpell(t *testing.T) {
	e := quickStarted(t)
	i := 0
	e.state().Party[i].Memorized[0] = 1 // a spell memorized
	if _, err := e.doNow(play.UsePower{Member: i, Kind: game.PowerCast, Target: -1}); err != nil {
		t.Fatal(err)
	}
	if q := commonOf(e.build()).Ask; q == nil || e.ext.pq.of != askSpell || len(q.Replies) == 0 {
		t.Fatalf("want a spell pick, got %+v", q)
	}
	if pq := e.ext.pq; pq == nil || pq.calls[0].kind != callPower || pq.calls[0].spell == 0 || pq.calls[0].member != i {
		t.Errorf("option 0 makes %+v, want the cast again with its spell", pq)
	}
}

// Quick Start's power picks come answered: each open one its first free option.
func TestDefaultPowers(t *testing.T) {
	e, c := &Engine{}, &rules1e.Character{}
	e.state().SetRules(pickRules{})
	e.state().DefaultPowers(c)
	if !slices.Equal(c.Powers, []int{1, 2}) {
		t.Errorf("defaults: %v", c.Powers)
	}
}

// TestPaladinPowerAllowed: in camp, where Cast is, a paladin's Heal is allowed
// as play.UsePower and asks whom through Do.
func TestPaladinPowerAllowed(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	e.Advance()
	e.state().Party[0].Levels[rules1e.ClsPaladin] = 1
	if _, err := e.doNow(play.MakeCamp{}); err != nil {
		t.Fatal(err)
	}
	heal := play.UsePower{Member: 0, Kind: game.PowerHeal, Target: -1}
	if !slices.Contains(e.build().(play.Camping).Allowed.Plain, play.Action(heal)) {
		t.Fatal("Heal not allowed in camp")
	}
	if _, err := e.doNow(heal); err != nil {
		t.Fatal(err)
	}
	if e.ext.pq == nil {
		t.Fatal("Heal asked no one")
	}
}
