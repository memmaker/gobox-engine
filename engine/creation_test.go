package engine

import (
	"errors"
	"slices"
	"strconv"
	"testing"

	"gobox/rules"
)

// fake: one choice field (two options, option 1 illegal once rolled 18+)
// and one skills field; Roll gives strength 10 then counts up.
type fake struct {
	made   []rules.Draft
	joined []rules.RosterID
}

const fKind, fSkills rules.FieldID = 0, 1

func (f *fake) Fields(d rules.Draft) []rules.FieldDef {
	sk := []string{"x", "y"}
	if d.Picks[fKind] == 1 { // the skills depend on the kind (BR career skills)
		sk = append(sk, "z")
	}
	return []rules.FieldDef{{ID: fKind, Options: []string{"a", "b"}},
		{ID: fSkills, Kind: rules.FieldSkills, Options: sk, Budget: 3}}
}
func (f *fake) Fits(d rules.Draft, id rules.FieldID, o int) bool { return true }
func (f *fake) Enforce(d rules.Draft, id rules.FieldID, o int, r *rules.RNG) rules.Draft {
	d.Picks[id] = o
	return d
}
func (f *fake) Roll(d *rules.Draft, k rules.RollKind, r *rules.RNG) error {
	if k == rules.RollMinimal {
		return &rules.MinimumError{Ability: rules.Strength, Want: 19, Best: 18}
	}
	if k == rules.RollPointBuy { // the bases, plus one (a roll's rules)
		d.Sheet.Scores = map[rules.Ability]int32{}
		for a, v := range d.Buy {
			d.Sheet.Scores[a] = v + 1
		}
		return nil
	}
	if d.Sheet.Scores == nil {
		d.Sheet.Scores = map[rules.Ability]int32{rules.Strength: 9}
	}
	d.Sheet.Scores[rules.Strength]++
	return nil
}
func (f *fake) Minimums(d rules.Draft) map[rules.Ability]int32 { return nil }
func (f *fake) Rolls(rules.Draft) []rules.RollKind {
	return []rules.RollKind{rules.RollFull, rules.RollMinimal, rules.RollPointBuy}
}

// Spend is the test double's pool: a point onto an option while the budget lasts, one back while the option has any.
func (f *fake) Spend(d rules.Draft, id rules.FieldID, o int, by int32) (rules.Draft, error) {
	p := slices.Clone(d.Points[id])
	var spent int32
	for _, v := range p {
		spent += v
	}
	if p[o]+by < 0 || spent+by > 3 {
		return d, &rules.SpendRefusal{Reason: rules.SpendNoPoints, Text: "no points left"}
	}
	p[o] += by
	d.Points = map[rules.FieldID][]int32{id: p}
	return d, nil
}
func (f *fake) QuickRerolls() int { return 2000 }

func (f *fake) QuickParty() []rules.Template {
	return []rules.Template{{Picks: map[rules.FieldID]int{fKind: 1}, Name: "Q", Need: map[rules.Ability]int32{rules.Strength: 15}}}
}
func (f *fake) SetScore(d rules.Draft, a rules.Ability, v int32) (rules.Draft, error) {
	if v > 18 && a == rules.Strength {
		return d, errors.New("range")
	}
	d.Sheet.Scores[a] = v
	return d, nil
}
func (f *fake) Make(d rules.Draft) (rules.CharID, rules.RosterID, error) {
	f.made = append(f.made, d)
	return rules.NoChar, rules.RosterID(len(f.made)), nil
}
func (f *fake) Join(id rules.RosterID) (rules.CharID, error) {
	f.joined = append(f.joined, id)
	return rules.CharID(len(f.joined)), nil
}
func (f *fake) Draft(c rules.CharID) (rules.Draft, error) {
	if c == rules.NoChar {
		return rules.Draft{}, nil
	}
	return f.made[c-1], nil
}

func TestDraft(t *testing.T) {
	r := &fake{}
	d, err := NewDraft(r, &rules.RNG{})
	if err != nil {
		t.Fatal(err)
	}
	var me *rules.MinimumError
	if err := d.Roll(rules.RollMinimal); !errors.As(err, &me) || me.Ability != rules.Strength || err.Error() != "STR 18/xx is for fighters only, max 18" {
		t.Errorf("minimal: %v", err)
	}
	d.Rename("Ab~cdefghijklmnopq")
	if d.D.Name != "Abcdefghijklmno" {
		t.Errorf("name %q", d.D.Name)
	}
	if err := d.SetScore(rules.Strength, "18/00"); err != nil || d.D.Sheet.Scores[rules.StrengthPercent] != 100 {
		t.Errorf("18/00: %v %v", err, d.D.Sheet.Scores)
	}
	if d.SetScore(rules.Strength, "19") == nil {
		t.Error("refusal lost")
	}
	if err := d.SetScore(rules.Strength, "18/5"); err != nil || d.D.Sheet.Scores[rules.StrengthPercent] != 5 {
		t.Errorf("18/5 (POR parseMinStat: one digit): %v", err)
	}
	for _, s := range []string{"17/50", "26", "18/100", "x"} {
		if _, _, err := ParseScore(s); err != ErrScore {
			t.Errorf("%q parsed", s)
		}
	}
	if d.Spend(fSkills, 0, 1) != nil || d.Spend(fSkills, 0, 1) != nil || d.Spend(fSkills, 1, 1) != nil || d.Spend(fSkills, 1, 1) == nil || d.Spend(fSkills, 1, -1) != nil {
		t.Error("budget")
	}
	if d.SetField(fKind, 1); len(d.D.Points[fSkills]) != 3 || d.D.Of != rules.NoChar {
		t.Errorf("fields after a pick: %v of %v", d.D.Points, d.D.Of)
	}
	for _, c := range [][2]int32{{18, 100}, {18, 5}, {12, 0}} {
		if v, p, _ := ParseScore(FormatScore(c[0], c[1])); v != c[0] || p != c[1] {
			t.Errorf("round trip %v", c)
		}
	}
	ids, err := QuickParty(r, &rules.RNG{})
	if err != nil || len(ids) != 1 || len(r.joined) != 1 || r.joined[0] != 1 || r.made[0].Sheet.Scores[rules.Strength] < 15 || r.made[0].Picks[fKind] != 1 || r.made[0].Name != "Q" {
		t.Errorf("quick party %v %v %+v", ids, err, r.made)
	}
	if m, _ := Modify(r, &rules.RNG{}, ids[0]); m.D.Of != ids[0] {
		t.Errorf("modify of %v", m.D.Of)
	}
}

// TestPointBuy: bought bases go through Roll(RollPointBuy); a buy over the
// budget or out of range is refused and the draft stays.
func TestDraftPointBuy(t *testing.T) {
	d, err := NewDraft(&fake{}, &rules.RNG{})
	if err != nil {
		t.Fatal(err)
	}
	was := d.D.Sheet.Scores[rules.Strength]
	for _, bad := range []map[rules.Ability]int32{{rules.Strength: 18, rules.Intelligence: 18, rules.Wisdom: 18}, {rules.Strength: 7}, {}} {
		if err := d.PointBuy(bad); !errors.Is(err, ErrPointBuy) || d.D.Sheet.Scores[rules.Strength] != was {
			t.Fatalf("%v: err %v, strength %d", bad, err, d.D.Sheet.Scores[rules.Strength])
		}
	}
	if err := d.PointBuy(map[rules.Ability]int32{rules.Strength: 18, rules.Intelligence: 16}); err != nil {
		t.Fatal(err)
	}
	if d.D.Sheet.Scores[rules.Strength] != 19 || d.D.Sheet.Scores[rules.Intelligence] != 17 || d.D.Buy != nil {
		t.Fatalf("scores %v, buy %v", d.D.Sheet.Scores, d.D.Buy)
	}
}

// FormatScore is ParseScore's inverse.
func FormatScore(v, pct int32) string {
	if pct == 0 {
		return strconv.Itoa(int(v))
	}
	return strconv.Itoa(int(v)) + "/" + strconv.Itoa(int(pct%100)/10) + strconv.Itoa(int(pct%10))
}
