package rulesxxvc

import (
	"slices"
	"testing"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// dice is one draw of the Pascal's RollDice(n, sides).
type dice struct{ n, sides int }

// replay is the stream after the draws the Pascal makes, in its order (seed 1's: 1d6 5, 1d100 30, 3d8 14 ...).
func replay(seed uint32, draws ...dice) uint32 {
	r := rules.RNG{State: seed}
	for _, d := range draws {
		rules.RNGDice(&r, d.n, d.sides)
	}
	return r.State
}

const antidoteItem = rules.ItemID(7)

// treatParty is a party of the characters in order, a fight pending its treatment, and the antidote's item.
func treatParty(cs ...*Character) (*Rules, []rules.CharID) {
	r := New()
	r.tables.ItemTypes = make([]xxvc.ItemType, 32)
	r.tables.Items = map[rules.ItemID]xxvc.ItemTemplate{antidoteItem: {Type: 31}}
	var ids []rules.CharID
	for i, c := range cs {
		id := rules.CharID(i + 1)
		c.Enabled = true
		r.chars[id] = c
		r.party = append(r.party, id)
		ids = append(ids, id)
	}
	r.treatPending = true
	r.seed = rules.RNG{State: 1}
	return r, ids
}

func okay(name string, hp, max int) *Character {
	return &Character{Name: name, HP: hp, MaxHP: max, Status: StOkay}
}

func medic(c *Character, skill SkillNo, v int8) *Character {
	c.Career = xxvc.Medic
	c.Skills.Set(skill, v)
	return c
}

func texts(t *testing.T, o rules.Outcome) []string {
	var out []string
	for _, e := range o.Effects {
		n, ok := e.(rules.Narrated)
		if !ok {
			t.Fatalf("effect %T", e)
		}
		out = append(out, n.Text)
	}
	return out
}

// Seed 1: 1d6 5, 1d100 30, 1d6 5, 3d8 15, 3d8 9, 2d8 12, 2d8 4, 1d8 3, 1d8 6. First aid, pass 0 (postcbt.pas 584..592): the
// medic's 1d6 5 is cut to the 4 points the patient lacks to 6, its skill 50 makes the d100 30 (20 < 25); the patient, okay
// now, is a medic too and rolls its 1d6 (cut to 0: no skill roll); the other three passes roll their dice for both medics
// (the patient is done only after a pass above 0 heals), whose heals are cut to 0 by the maximum 6.
func TestTreatFirstAid(t *testing.T) {
	m := medic(okay("Mia", 20, 20), FirstAid, 50)
	p := &Character{Name: "Pat", HP: 2, MaxHP: 6, Status: StUnconscious}
	r, _ := treatParty(m, p)
	o, err := r.Resolve(nil, rules.TreatInjuries{}, nil)
	if err != nil || o.Confirm != "" {
		t.Fatalf("%v %q", err, o.Confirm)
	}
	if p.HP != 6 || p.Status != StOkay || !p.Enabled {
		t.Errorf("patient %d %v", p.HP, p.Status)
	}
	want := []string{treatWoundsHeading, "Mia Pat 4", "  Healing completed."}
	if got := texts(t, o); !slices.Equal(got, want) {
		t.Errorf("lines %q, want %q", got, want)
	}
	draws := []dice{{1, 6}, {1, 100}, {1, 6}, {3, 8}, {3, 8}, {2, 8}, {2, 8}, {1, 8}, {1, 8}}
	if r.seed.State != replay(1, draws...) {
		t.Error("the draws are not the Pascal's, in its order")
	}
	if o, _ := r.Resolve(nil, rules.TreatInjuries{}, nil); len(o.Effects) != 0 {
		t.Error("treated twice: the treatment pends once")
	}
}

// Seed 1: 1d6 5, 1d100 30, 3d8 14, 1d100 86, 2d8 8, 1d100 97, 1d8 7, 1d100 17. A member treating itself (pass 1) rolls
// at a quarter of its skill: 90 / 4 = 22 against the d100 86 fails; the rest of the passes fail for want of skill.
func TestTreatCriticalOnSelfIsHard(t *testing.T) {
	s := medic(okay("Sam", 3, 30), TreatCritical, 90)
	r, _ := treatParty(s)
	o, _ := r.Resolve(nil, rules.TreatInjuries{}, nil)
	if s.HP != 3 {
		t.Errorf("healed itself to %d", s.HP)
	}
	if got, want := texts(t, o), []string{"  No injuries were successfully treated."}; !slices.Equal(got, want) {
		t.Errorf("lines %q", got)
	}
	draws := []dice{{1, 6}, {1, 100}, {3, 8}, {1, 100}, {2, 8}, {1, 100}, {1, 8}, {1, 100}}
	if r.seed.State != replay(1, draws...) {
		t.Error("the draws are not the Pascal's, in its order")
	}
}

// Seed 1 as above: another medic's pass 1 is at the skill itself: 90 makes the d100 86; (3d8 14 + 3) * 2 = 34 is cut to
// the 27 the patient lacks; it is done and the passes 2 and 3 roll nothing.
func TestTreatCriticalOnAnother(t *testing.T) {
	s := medic(okay("Sam", 30, 30), TreatCritical, 90)
	p := &Character{Name: "Pat", HP: 3, MaxHP: 30, Status: StUnconscious}
	r, _ := treatParty(s, p)
	o, _ := r.Resolve(nil, rules.TreatInjuries{}, nil)
	if p.HP != 30 || p.Status != StOkay {
		t.Errorf("patient %d %v", p.HP, p.Status)
	}
	if got, want := texts(t, o), []string{treatWoundsHeading, "Sam Pat 27", "  Healing completed."}; !slices.Equal(got, want) {
		t.Errorf("lines %q", got)
	}
	if r.seed.State != replay(1, dice{1, 6}, dice{1, 100}, dice{3, 8}, dice{1, 100}) {
		t.Error("the draws are not the Pascal's, in its order")
	}
}

// Seed 1: 1d100 57. A medic with Treat Poisoning 90 makes it and cures the patient (unconscious, no hit points): no question.
func TestTreatPoisonCured(t *testing.T) {
	m := medic(okay("Mia", 10, 10), TreatPoisoning, 90)
	p := &Character{Name: "Pat", Status: StPoisoned}
	r, _ := treatParty(m, p)
	o, _ := r.Resolve(nil, rules.TreatInjuries{}, nil)
	if o.Confirm != "" || p.Status != StUnconscious || p.Enabled || p.HP != 0 {
		t.Errorf("%q %v", o.Confirm, p.Status)
	}
	if got, want := texts(t, o), []string{treatPoisoningHeading, "Mia Pat cured", "  Healing completed."}; !slices.Equal(got, want) {
		t.Errorf("lines %q", got)
	}
	if r.seed.State != replay(1, dice{1, 100}) {
		t.Error("the draws are not the Pascal's, in its order")
	}
}

// Seed 1: 1d100 57 fails a medic with no Treat Poisoning; the first antidote any member carries is offered (the carrier is
// the second member, a stack of two): Yes uses one and cures, No leaves it and the patient slips into a coma.
func TestTreatAntidote(t *testing.T) {
	for _, yes := range []bool{true, false} {
		m := medic(okay("Mia", 10, 10), TreatPoisoning, 0)
		c := okay("Cal", 10, 10)
		c.Items = []Carried{{Ref: rules.ItemRef{ID: antidoteItem, Count: 2}}}
		p := &Character{Name: "Pat", Status: StPoisoned}
		r, _ := treatParty(m, c, p)
		o, _ := r.Resolve(nil, rules.TreatInjuries{}, nil)
		if o.Confirm != "Use antidote on Pat? " {
			t.Fatalf("question %q", o.Confirm)
		}
		if got, want := texts(t, o), []string{treatPoisoningHeading, "Pat is dying of poison!"}; !slices.Equal(got, want) {
			t.Errorf("lines %q", got)
		}
		o, err := r.Resolve(nil, rules.Antidote{Yes: yes}, nil)
		if err != nil || o.Confirm != "" {
			t.Fatalf("%v %q", err, o.Confirm)
		}
		if yes {
			if p.Status != StUnconscious || c.Items[0].Ref.Count != 1 {
				t.Errorf("yes: %v, %d left", p.Status, c.Items[0].Ref.Count)
			}
			if got, want := texts(t, o), []string{"Pat is cured", "  Healing completed."}; !slices.Equal(got, want) {
				t.Errorf("lines %q", got)
			}
		} else {
			if p.Status != StComatose || p.Enabled || c.Items[0].Ref.Count != 2 {
				t.Errorf("no: %v, %d left", p.Status, c.Items[0].Ref.Count)
			}
			if got, want := texts(t, o), []string{"Pat slips into a coma.", "  No injuries were successfully treated."}; !slices.Equal(got, want) {
				t.Errorf("lines %q", got)
			}
		}
		if _, err := r.Resolve(nil, rules.Antidote{Yes: yes}, nil); err == nil {
			t.Error("answered twice")
		}
	}
}

// A last antidote is used up; with none anywhere the coma needs no question.
func TestTreatAntidoteUsedUpOrNone(t *testing.T) {
	c := okay("Cal", 10, 10)
	c.Items = []Carried{{Ref: rules.ItemRef{ID: antidoteItem, Count: 1}}}
	p := &Character{Name: "Pat", Status: StPoisoned}
	r, _ := treatParty(c, p)
	r.Resolve(nil, rules.TreatInjuries{}, nil)
	r.Resolve(nil, rules.Antidote{Yes: true}, nil)
	if len(c.Items) != 0 {
		t.Error("the last antidote stays")
	}
	q := &Character{Name: "Quin", Status: StPoisoned}
	r, _ = treatParty(okay("Cal", 10, 10), q)
	o, _ := r.Resolve(nil, rules.TreatInjuries{}, nil)
	if o.Confirm != "" || q.Status != StComatose {
		t.Errorf("%q %v", o.Confirm, q.Status)
	}
}

// The summary lines (postcbt.pas 690..698) and UNKILLABLE: a hurt member with no medic in the party; a member holding
// UNKILLABLE below 45 hit points is back on its feet with 45.
func TestTreatSummaryAndUnkillable(t *testing.T) {
	cases := []struct {
		name string
		cs   []*Character
		want string
	}{
		{"healthy", []*Character{okay("Al", 5, 5)}, "  No injuries were sustained."},
		{"no medic", []*Character{{Name: "Al", HP: 1, MaxHP: 5, Status: StUnconscious}}, "  No medics are available."},
	}
	for _, c := range cases {
		r, _ := treatParty(c.cs...)
		o, _ := r.Resolve(nil, rules.TreatInjuries{}, nil)
		if got := texts(t, o); !slices.Equal(got, []string{c.want}) {
			t.Errorf("%s: %q", c.name, got)
		}
	}
	u := &Character{Name: "Una", HP: 10, MaxHP: 60, Status: StUnconscious, Effects: []Effect{{ID: EffUnkillable}}}
	r, _ := treatParty(u)
	r.Resolve(nil, rules.TreatInjuries{}, nil)
	if u.Status != StOkay || !u.Enabled || u.HP != 45 {
		t.Errorf("unkillable: %v %d", u.Status, u.HP)
	}
}

// Skip, or no fight pending: nothing is rolled; the pending treatment is cleared by a skip.
func TestTreatSkips(t *testing.T) {
	p := &Character{Name: "Pat", HP: 1, MaxHP: 6, Status: StUnconscious}
	r, _ := treatParty(medic(okay("Mia", 5, 5), FirstAid, 50), p)
	if o, _ := r.Resolve(nil, rules.TreatInjuries{Skip: true}, nil); len(o.Effects) != 0 || p.HP != 1 || r.seed.State != 1 {
		t.Error("skipped, yet it treated")
	}
	if o, _ := r.Resolve(nil, rules.TreatInjuries{}, nil); len(o.Effects) != 0 || r.treatPending {
		t.Error("a skip leaves the treatment pending")
	}
	r.treatPending = false
	if o, _ := r.Resolve(nil, rules.TreatInjuries{}, nil); len(o.Effects) != 0 || r.seed.State != 1 {
		t.Error("treated with no fight pending")
	}
}
