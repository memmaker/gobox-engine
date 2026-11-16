package rulesxxvc

import (
	"image"
	"strings"
	"testing"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// parJetpack is a lone member with a readied jetpack on a 5x5 open board in a wilderness, its turn begun.
func parJetpack(t *testing.T, skill int8) (*Rules, rules.CharID, rules.Fight, *rules.RNG) {
	t.Helper()
	r := New()
	r.tables = xxvc.Tables{ItemTypes: []xxvc.ItemType{{Slot: xxvc.SlotJetpack}}, Items: map[rules.ItemID]xxvc.ItemTemplate{1: {}}}
	a := member("a", 10, 0, 0, 0)
	a.BaseMovement, a.Items = 12, []Carried{{Slot: 1, Ref: rules.ItemRef{ID: 1}, Readied: true}}
	a.Skills.Set(jetpackSkill, skill)
	id := r.Add(a)
	f := rules.Fight{Party: rules.Party{id}, Actor: id, Terms: rules.FightTerms{Area: rules.AreaWilderness},
		Board: rules.Board{Size: image.Pt(5, 5), Units: []rules.BoardUnit{{ID: id, At: rules.BoardPos{X: 1, Y: 1}}}}}
	rng := rules.RNG{State: 1}
	r.BeginFight(f, &rng)
	r.BeginRound(f, &rng)
	return r, id, f, &rng
}

func jetpackSquares(r *Rules, f rules.Fight) (n int) {
	for _, x := range r.allowedFight(f).Aimed {
		if _, ok := x.Action.(rules.Jetpack); ok {
			n++
		}
	}
	return n
}

// TestJetpackOfferedOnlyWhen ports rules TestJetpack's offer (combat.pas CombatMenu): the jetpack is offered with
// movement left, not Dodging, not after a flight (Used Jetpack), a readied jetpack and a wilderness area.
func TestJetpackOfferedOnlyWhen(t *testing.T) {
	r, id, f, _ := parJetpack(t, 100)
	if n := jetpackSquares(r, f); n != 24 {
		t.Fatalf("offered at %d squares, want the 24 free ones", n)
	}
	for _, c := range []struct {
		what string
		undo func() func()
	}{
		{"Dodging", func() func() { r.chars[id].Status = StDodging; return func() { r.chars[id].Status = StOkay } }},
		{"Used Jetpack", func() func() { r.chars[id].Status = StUsedJetpack; return func() { r.chars[id].Status = StOkay } }},
		{"no movement left", func() func() {
			m := r.fight[id].movement
			r.fight[id].movement = 0
			return func() { r.fight[id].movement = m }
		}},
		{"jetpack not readied", func() func() {
			r.chars[id].Items[0].Readied = false
			return func() { r.chars[id].Items[0].Readied = true }
		}},
		{"not a wilderness", func() func() { f.Terms.Area = rules.AreaPlain; return func() { f.Terms.Area = rules.AreaWilderness } }},
	} {
		redo := c.undo()
		if n := jetpackSquares(r, f); n != 0 {
			t.Errorf("%s: offered at %d squares", c.what, n)
		}
		redo()
		if n := jetpackSquares(r, f); n != 24 {
			t.Fatalf("after %s undone: %d squares", c.what, n)
		}
	}
}

// TestJetpackFlight ports rules TestJetpack's flights (cbtacts.pas CombatUseJetpack): a made roll flies to the chosen
// square; a failed one says so, tumbles to a free square inside the board, and either way the movement is gone and
// the status Used Jetpack.
func TestJetpackFlight(t *testing.T) {
	to := rules.BoardPos{X: 4, Y: 3}
	r, id, f, rng := parJetpack(t, 100)
	o, err := r.Resolve(f, rules.Jetpack{To: to}, rng)
	if err != nil {
		t.Fatal(err)
	}
	var walked []rules.Walked
	for _, e := range o.Effects {
		if w, ok := e.(rules.Walked); ok {
			walked = append(walked, w)
		}
	}
	if len(walked) != 1 || walked[0].Path[len(walked[0].Path)-1] != to || r.chars[id].Status != StUsedJetpack || r.fight[id].movement != 0 {
		t.Errorf("made: walked %+v status %v movement %d", walked, r.chars[id].Status, r.fight[id].movement)
	}

	r, id, f, rng = parJetpack(t, 0) // a skill of 0 always fails
	o, err = r.Resolve(f, rules.Jetpack{To: to}, rng)
	if err != nil {
		t.Fatal(err)
	}
	said, landed := false, rules.BoardPos{X: 1, Y: 1}
	for _, e := range o.Effects {
		switch e := e.(type) {
		case rules.Narrated:
			said = said || strings.Contains(e.Text, "fails at use jetpack")
		case rules.Walked:
			landed = e.Path[len(e.Path)-1]
		}
	}
	if !said || r.chars[id].Status != StUsedJetpack || r.fight[id].movement != 0 ||
		landed.X < 0 || landed.Y < 0 || landed.X >= 5 || landed.Y >= 5 {
		t.Errorf("failed: said %v status %v movement %d landed %v", said, r.chars[id].Status, r.fight[id].movement, landed)
	}
}
