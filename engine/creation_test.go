package engine

import (
	"errors"
	"testing"

	"gobox/core"
)

// fake: one choice field (two options, option 1 illegal once rolled 18+)
// and one skills field; Roll gives strength 10 then counts up.
type fake struct {
	made   []core.Draft
	joined []core.RosterID
}

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
func (f *fake) Roll(d *core.Draft, k core.RollKind, r *core.RNG) error {
	if k == core.RollMinimal {
		return &core.MinimumError{Ability: core.Strength, Want: 19, Best: 18}
	}
	if k == core.RollPointBuy { // the bases, plus one (a roll's rules)
		d.Sheet.Scores = map[core.Ability]int32{}
		for a, v := range d.Buy {
			d.Sheet.Scores[a] = v + 1
		}
		return nil
	}
	if d.Sheet.Scores == nil {
		d.Sheet.Scores = map[core.Ability]int32{core.Strength: 9}
	}
	d.Sheet.Scores[core.Strength]++
	return nil
}
func (f *fake) Minimums(d core.Draft) map[core.Ability]int32 { return nil }
func (f *fake) QuickRerolls() int { return 2000 }

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
func (f *fake) Make(d core.Draft) (core.CharID, core.RosterID, error) {
	f.made = append(f.made, d)
	return core.NoChar, core.RosterID(len(f.made)), nil
}
func (f *fake) Join(id core.RosterID) (core.CharID, error) {
	f.joined = append(f.joined, id)
	return core.CharID(len(f.joined)), nil
}
func (f *fake) Draft(c core.CharID) (core.Draft, error) {
	if c == core.NoChar {
		return core.Draft{}, nil
	}
	return f.made[c-1], nil
}

func TestDraft(t *testing.T) {
	r := &fake{}
	d, err := NewDraft(r, &core.RNG{})
	if err != nil {
		t.Fatal(err)
	}
	var me *core.MinimumError
	if err := d.Roll(core.RollMinimal); !errors.As(err, &me) || me.Ability != core.Strength || err.Error() != "STR 18/xx is for fighters only, max 18" {
		t.Errorf("minimal: %v", err)
	}
	d.Rename("Ab~cdefghijklmnopq")
	if d.D.Name != "Abcdefghijklmno" {
		t.Errorf("name %q", d.D.Name)
	}
	if err := d.SetScore(core.Strength, "18/00"); err != nil || d.D.Sheet.Scores[core.StrengthPercent] != 100 {
		t.Errorf("18/00: %v %v", err, d.D.Sheet.Scores)
	}
	if d.SetScore(core.Strength, "19") == nil {
		t.Error("refusal lost")
	}
	if err := d.SetScore(core.Strength, "18/5"); err != nil || d.D.Sheet.Scores[core.StrengthPercent] != 5 {
		t.Errorf("18/5 (POR parseMinStat: one digit): %v", err)
	}
	for _, s := range []string{"17/50", "26", "18/100", "x"} {
		if _, _, err := ParseScore(s); err != ErrScore {
			t.Errorf("%q parsed", s)
		}
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
	if err != nil || len(ids) != 1 || len(r.joined) != 1 || r.joined[0] != 1 || r.made[0].Sheet.Scores[core.Strength] < 15 || r.made[0].Picks[fKind] != 1 || r.made[0].Name != "Q" {
		t.Errorf("quick party %v %v %+v", ids, err, r.made)
	}
	if m, _ := Modify(r, &core.RNG{}, ids[0]); m.D.Of != ids[0] {
		t.Errorf("modify of %v", m.D.Of)
	}
}

// TestPointBuy: bought bases go through Roll(RollPointBuy); a buy over the
// budget or out of range is refused and the draft stays.
func TestDraftPointBuy(t *testing.T) {
	d, err := NewDraft(&fake{}, &core.RNG{})
	if err != nil {
		t.Fatal(err)
	}
	was := d.D.Sheet.Scores[core.Strength]
	for _, bad := range []map[core.Ability]int32{{core.Strength: 18, core.Intelligence: 18, core.Wisdom: 18}, {core.Strength: 7}, {}} {
		if err := d.PointBuy(bad); !errors.Is(err, ErrPointBuy) || d.D.Sheet.Scores[core.Strength] != was {
			t.Fatalf("%v: err %v, strength %d", bad, err, d.D.Sheet.Scores[core.Strength])
		}
	}
	if err := d.PointBuy(map[core.Ability]int32{core.Strength: 18, core.Intelligence: 16}); err != nil {
		t.Fatal(err)
	}
	if d.D.Sheet.Scores[core.Strength] != 19 || d.D.Sheet.Scores[core.Intelligence] != 17 || d.D.Buy != nil {
		t.Fatalf("scores %v, buy %v", d.D.Sheet.Scores, d.D.Buy)
	}
}
