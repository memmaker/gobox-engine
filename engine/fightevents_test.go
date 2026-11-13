package engine

import (
	"gobox/rules1e"
	"image"
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// TestFightEventsOut (BUG-CLASSES 20): a headless fight (the computer for
// both sides) hands out a play event for each kind of the rules' facts, and
// a unit's damage shows as lower HP in the next State.
func TestFightEventsOut(t *testing.T) { testFightEventsOut(t, testEngine) }

func testFightEventsOut(t *testing.T, mk func(*testing.T) *Engine) {
	e := mk(t)
	e.Explore = true
	scene := e.DevScene("autocombat")
	e.DevAction = func() {
		scene()
		for _, c := range e.state().Party { // 20b: casters with Magic Missile, so spells go out too
			c.Memorized[0] = rules1e.SpMagicMissile
		}
	}
	seed := uint32(3)
	e.FixedSeed = &seed
	e.Init()
	e.Start(e.Play)
	hp := func(s play.State, id int) int {
		if f, ok := s.(play.Fighting); ok {
			for _, u := range f.Field.Units {
				if u.ID == id {
					return u.HP
				}
			}
		}
		return -1
	}
	unit := func(s play.State, id int) *game.Unit {
		if f, ok := s.(play.Fighting); ok {
			for i := range f.Field.Units {
				if f.Field.Units[i].ID == id {
					return &f.Field.Units[i]
				}
			}
		}
		return nil
	}
	kind := func(ev play.Event) string {
		switch ev.(type) {
		case play.CastSpell:
			return "spell"
		case play.Shot:
			return "missile"
		case play.Walked:
			return "move"
		case play.Attacked:
			return "attack"
		case play.Struck:
			return "damage"
		case play.Fell:
			return "down"
		}
		return ""
	}
	seen := map[string]int{}
	casts37b, fought := 0, false
	prev := e.build()
	for i := 0; i < 2000; i++ {
		_, fighting := prev.(play.Fighting)
		fought = fought || fighting || seen["attack"] > 0
		if fought && !fighting {
			break
		}
		var out play.Outcome
		var err error
		if i == 0 && fighting { // the dev start opens the fight before any Do: it runs on as Do's would
			err = e.stepOn(&out, nil)
		} else if q := commonOf(prev).Ask; q != nil {
			n := 0
			if q.Text == "Continue Battle:" && !q.YesNo { // No: the fight is won
				n = 1
			}
			out, err = e.doNow(play.Reply{N: n})
		} else if playAllows(prev, play.Continue{}) {
			out, err = e.doNow(play.Continue{})
		} else {
			t.Fatalf("a fight at rest with nothing to do: %T %+v", prev, commonOf(prev).Allowed) // Do runs the computer's turns on
		}
		if err != nil {
			t.Fatal(err)
		}
		s := e.build()
		for j, ev := range out.Events {
			seen[kind(ev)]++
			switch ev := ev.(type) {
			case play.Shot: // 20b
				if ev.From == 0 {
					t.Errorf("missile to %v has no shooter", ev.To)
				}
			case play.CastSpell:
				if len(ev.Ats) == 0 || ev.Ats[0] == (image.Point{}) {
					t.Errorf("spell %d by %d aimed at (0,0)", ev.Spell, ev.From)
				}
				for _, p := range out.Events[:j] { // 37b: the cast's pose before its flight
					if p, ok := p.(play.Shot); ok && p.From == ev.From && len(ev.Ats) > 0 && p.To == ev.Ats[0] {
						t.Errorf("spell %d by %d: its missile flew before the cast", ev.Spell, ev.From)
						casts37b--
						break
					}
				}
				casts37b++
			case play.Attacked:
				if !ev.Hit {
					break
				}
			hit:
				for _, n := range out.Events[j+1:] { // the hit's damage is its own: no Struck for it before the next action
					switch n := n.(type) {
					case play.Attacked, play.CastSpell: // another swing's or a spell's harm is its own (cotab: a spell on the one just hit)
						break hit
					case play.Struck:
						if n.Unit == ev.To {
							t.Errorf("unit %d's hit on %d shows its damage twice", ev.From, ev.To)
						}
					}
				}
			}
		}
		took, fell := map[int]int{}, map[int]bool{} // the whole outcome's harm, per unit
		for _, ev := range out.Events {
			switch ev := ev.(type) {
			case play.Struck:
				took[ev.Unit] += ev.Damage
			case play.Attacked:
				took[ev.To] += ev.Damage
			case play.Fell:
				fell[ev.Unit] = true
			}
		}
		for id, d := range took {
			if b, a := hp(prev, id), hp(s, id); d > 0 && b >= 0 && a >= 0 && a >= b {
				t.Errorf("unit %d took %d: HP %d then %d", id, d, b, a)
			}
		}
		for id := range fell {
			if _, on := s.(play.Fighting); !on {
				break // the fight is over: no field to look at
			}
			if u := unit(s, id); u != nil && !u.Dead && !u.Gone && u.HP > 0 {
				t.Errorf("unit %d fell but stands: %+v", id, u)
			}
		}
		prev = s
	}
	t.Logf("events per kind: %v", seen)
	if !fought {
		t.Fatal("no fight")
	}
	for _, k := range []string{"spell", "missile", "move", "attack", "down"} {
		if seen[k] == 0 {
			t.Errorf("no %s event out of the fight (seen %v)", k, seen)
		}
	}
	if casts37b == 0 {
		t.Error("37b: no cast checked")
	}
}
