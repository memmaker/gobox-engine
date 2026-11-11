package creation

import (
	"errors"
	"testing"

	"gobox/core"
)

// fake: one choice field (two options, option 1 illegal once rolled 18+)
// and one skills field; Roll gives strength 10 then counts up.
type fake struct{ made []core.Draft }

const fKind, fSkills core.FieldID = 0, 1

func (f *fake) Fields(d core.Draft) []core.FieldDef {
	sk := []string{"x", "y"}
	if d.Picks[fKind] == 1 { // the skills depend on the kind (BR career skills)
		sk = append(sk, "z")
	}
	return []core.FieldDef{{ID: fKind, Options: []string{"a", "b"}},
		{ID: fSkills, Kind: core.FieldSkills, Options: sk, Budget: 3}}
}
func (f *fake) Fits(d core.Draft, id core.FieldID, o int) bool { return true }
func (f *fake) Enforce(d core.Draft, id core.FieldID, o int, r *core.RNG) core.Draft {
	d.Picks[id] = o
	return d
}
func (f *fake) Roll(d core.Draft, k core.RollKind, r *core.RNG) core.Draft {
	if d.Sheet.Scores == nil {
		d.Sheet.Scores = map[core.Ability]int32{core.Strength: 9}
	}
	d.Sheet.Scores[core.Strength]++
	return d
}
func (f *fake) Minimums(d core.Draft) map[core.Ability]int32 { return nil }
func (f *fake) QuickParty() []core.Template {
	return []core.Template{{Picks: map[core.FieldID]int{fKind: 1}, Name: "Q", Need: map[core.Ability]int32{core.Strength: 15}}}
}
func (f *fake) SetScore(d core.Draft, a core.Ability, v int32) (core.Draft, error) {
	if v > 18 && a == core.Strength {
		return d, errors.New("range")
	}
	d.Sheet.Scores[a] = v
	return d, nil
}
func (f *fake) Make(d core.Draft) (core.CharID, error) {
	f.made = append(f.made, d)
	return core.CharID(len(f.made)), nil
}
func (f *fake) Draft(c core.CharID) (core.Draft, error) { return f.made[c-1], nil }

func TestDraft(t *testing.T) {
	r := &fake{}
	d := New(r, &core.RNG{})
	d.Rename("Ab~cdefghijklmnopq")
	if d.D.Name != "Abcdefghijklmno" {
		t.Errorf("name %q", d.D.Name)
	}
	if err := d.SetScore(core.Strength, "18/00"); err != nil || d.D.Sheet.Scores[core.StrengthPercent] != 100 {
		t.Errorf("18/00: %v %v", err, d.D.Sheet.Scores)
	}
	if d.SetScore(core.Strength, "19") == nil || d.SetScore(core.Strength, "18/5") == nil {
		t.Error("bad scores taken")
	}
	sk := d.Fields[1]
	if d.Spend(sk, 0, 2) != nil || d.Spend(sk, 1, 2) == nil || d.Spend(sk, 1, -1) == nil {
		t.Error("budget")
	}
	if d.SetField(fKind, 1); len(d.D.Points[fSkills]) != 3 || d.D.Of != core.NoChar {
		t.Errorf("fields after a pick: %v of %v", d.D.Points, d.D.Of)
	}
	for _, c := range [][2]int32{{18, 100}, {18, 5}, {12, 0}} {
		if v, p, _ := ParseScore(FormatScore(c[0], c[1])); v != c[0] || p != c[1] {
			t.Errorf("round trip %v", c)
		}
	}
	ids, err := QuickParty(r, &core.RNG{})
	if err != nil || len(ids) != 1 || r.made[0].Sheet.Scores[core.Strength] < 15 || r.made[0].Picks[fKind] != 1 || r.made[0].Name != "Q" {
		t.Errorf("quick party %v %v %+v", ids, err, r.made)
	}
	if m, _ := Modify(r, &core.RNG{}, ids[0]); m.D.Of != ids[0] {
		t.Errorf("modify of %v", m.D.Of)
	}
}
