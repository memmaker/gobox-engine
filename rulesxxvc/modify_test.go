package rulesxxvc

import (
	"slices"
	"testing"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// memberOf makes a character from a draft and puts it in the party with a carried item.
func memberOf(t *testing.T, career rules.Class, seed uint32) (creation, rules.CharID) {
	t.Helper()
	c := creation{r: specTables()}
	d := newDraft(t, c, 0, 1, slices.Index(pickCareers, career), seed)
	d.Name = "Zed"
	id, err := c.r.insertDraft(d, false)
	if err != nil {
		t.Fatal(err)
	}
	c.r.chars[id].Items = []Carried{{Ref: rules.ItemRef{ID: 1}, Readied: true}}
	c.r.party = append(c.r.party, id)
	return c, id
}

// TestModifiable: chargen.pas ModifyCharacterMenu "can't be modified": only a member whose experience is the
// career's starting one.
func TestModifiable(t *testing.T) {
	c, id := memberOf(t, xxvc.Warrior, 4)
	if !c.r.Check(rules.Modifiable{Who: id}, nil).Yes {
		t.Fatal("a new character is not modifiable")
	}
	c.r.chars[id].Exp++
	if c.r.Check(rules.Modifiable{Who: id}, nil).Yes {
		t.Error("a character with experience of its own is modifiable")
	}
}

// TestModifyMember: the member as a draft keeps its race, gender and career, its abilities stay in the range and
// the hit points in the range its constitution allows; Keep keeps the items and sets the hit points to the maximum.
func TestModifyMember(t *testing.T) {
	for seed := uint32(1); seed <= 60; seed++ {
		c, id := memberOf(t, xxvc.Medic, seed)
		d, err := c.Draft(id)
		if err != nil || d.Of != id || d.Name != "Zed" || d.Sheet.Race != xxvc.Terran || d.Sheet.Sex != rules.Female {
			t.Fatalf("draft %+v (%v)", d, err)
		}
		fields := c.Fields(d)
		if len(fields) != 3 {
			t.Fatalf("a member's fields %d: no skills, no weapon", len(fields))
		}
		for _, f := range fields {
			for o := range f.Options {
				if c.Fits(d, f.ID, o) != (o == d.Picks[f.ID]) {
					t.Fatalf("field %d option %d fits", f.ID, o)
				}
			}
		}
		if e := c.Enforce(d, fieldCareer, 0, &rules.RNG{State: 1}); e.Sheet.Scores[rules.Strength] != d.Sheet.Scores[rules.Strength] {
			t.Fatal("a member's career was changed")
		}
		if len(c.Rolls(d)) != 0 || c.Roll(&d, rules.RollFull, nil) == nil {
			t.Fatal("a member is rolled again")
		}
		hp := d.Sheet.MaxHP
		d, _ = c.SetScore(d, rules.Strength, 99) // Terran adjusts STR 0: at most 18
		d, _ = c.SetScore(d, rules.Intelligence, 1)
		d, _ = c.SetScore(d, rules.Constitution, 3)
		if d.Sheet.Scores[rules.Strength] != 18 || d.Sheet.Scores[rules.Intelligence] != 12 { // the medic's minimum
			t.Fatalf("scores %v", d.Sheet.Scores)
		}
		if d.Sheet.MaxHP < d.MinHP || d.Sheet.MaxHP > d.MaxHP || d.Sheet.HP != d.Sheet.MaxHP {
			t.Fatalf("hit points %d/%d in %d..%d (were %d)", d.Sheet.HP, d.Sheet.MaxHP, d.MinHP, d.MaxHP, hp)
		}
		d.Name = "Zeta"
		if _, _, err := c.Make(d); err != nil {
			t.Fatal(err)
		}
		m := c.r.chars[id]
		if m.Name != "Zeta" || m.Scores[rules.Strength] != 18 || m.MaxHP != int(d.Sheet.MaxHP) || m.HP != m.MaxHP || len(m.Items) != 1 {
			t.Fatalf("kept %+v", m)
		}
	}
}

// The sheet draws the member in place; HoldCopy and RestoreCopy put it back as it was.
func TestModifyDiscard(t *testing.T) {
	c, id := memberOf(t, xxvc.Warrior, 9)
	was := c.r.chars[id].Scores[rules.Dexterity]
	if _, err := c.r.Resolve(nil, rules.HoldCopy{Who: id}, nil); err != nil {
		t.Fatal(err)
	}
	d, _ := c.Draft(id)
	d, _ = c.SetScore(d, rules.Dexterity, 3+int32(raceStatAdjust[xxvc.Terran][1])+1)
	o, err := c.r.Resolve(nil, rules.SheetDraft{Draft: d}, nil)
	if err != nil || o.Created != id || c.r.chars[id].Scores[rules.Dexterity] == was && was != int(d.Sheet.Scores[rules.Dexterity]) {
		t.Fatalf("sheet draft %v %v", o, err)
	}
	if _, err := c.r.Resolve(nil, rules.RestoreCopy{Who: id}, nil); err != nil {
		t.Fatal(err)
	}
	if c.r.chars[id].Scores[rules.Dexterity] != was {
		t.Errorf("dexterity %d after the discard, was %d", c.r.chars[id].Scores[rules.Dexterity], was)
	}
}
