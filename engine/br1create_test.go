package engine

import (
	"slices"
	"testing"

	"gobox/game"
	"gobox/rules"
	"gobox/rules/xxvc"
)

// TestBR1Create: a BR1 character made as a front makes it (the engine's creation actions, one seed), kept to the
// roster, added to the party and the game begun. The rules are chargen.pas CreateCharacterMenu: race, gender and
// career fields, abilities with the race's adjustment and the career's minimum, hit points from the free training to
// level 2, 1000 credits and the free ammunition.
func TestBR1Create(t *testing.T) {
	e := br1Engine(t)
	e.Open("")
	e.titleDone()
	e.Advance()
	draft := func() *game.Draft { a, _ := e.build().(game.Assembling); return a.Draft }
	if err := e.createCharacter(); err != nil {
		t.Fatal(err)
	}
	d := draft()
	if d == nil || len(d.Fields) != 3 || d.Fields[0].Name != "Race" || d.Fields[1].Name != "Gender" || d.Fields[2].Name != "Class" {
		t.Fatalf("fields %+v", d)
	}
	if !slices.Equal(d.Fields[0].Options, []string{"Terran", "Martian", "Venusian", "Mercurian", "Tinker", "Desert Runner"}) ||
		!slices.Equal(d.Fields[2].Options, []string{"Rocket Jock", "Medic", "Warrior", "Engineer", "Rogue"}) {
		t.Fatalf("options %v / %v", d.Fields[0].Options, d.Fields[2].Options)
	}
	// A Tinker may be a medic or an engineer only: the career moves to the first of them.
	if err := e.SetField(0, slices.Index(d.Fields[0].Options, "Tinker")); err != nil {
		t.Fatal(err)
	}
	d = draft()
	if d.Fields[2].Options[d.Fields[2].Pick] != "Medic" || d.Fields[2].Fits[slices.Index(d.Fields[2].Options, "Warrior")] {
		t.Fatalf("tinker career %+v", d.Fields[2])
	}
	if d.Member.Class != "Medic" || d.Member.MaxHP < 2 || d.Member.MaxHP > 2*(10+3) {
		t.Fatalf("draft member %+v", d.Member)
	}
	if !slices.Equal(d.Rolls, []rules.RollKind{rules.RollFull}) {
		t.Fatalf("BR offers the original's one roll: %v", d.Rolls)
	}
	var labels []string
	for _, st := range d.Stats {
		labels = append(labels, st.Label)
	}
	if !slices.Equal(labels, []string{"STR", "DEX", "CON", "INT", "WIS", "CHA", "TCH"}) {
		t.Fatalf("the draft lists the seven abilities in the original's order: %v", labels)
	}
	for _, k := range []string{"str", "dex", "con", "int", "wis", "cha", "tch"} {
		if v := d.Member.Stats[k]; v < 1 {
			t.Errorf("ability %s %d", k, v)
		}
	}
	if d.Member.Age < 19 || d.Member.Age > 24 {
		t.Fatalf("age %d, want 1d6+18", d.Member.Age)
	}
	if err := e.rename("Zed"); err != nil {
		t.Fatal(err)
	}
	if err := e.EndCreation(true); err != nil {
		t.Fatal(err)
	}
	list := e.BuildCharList()
	if len(list) == 0 || list[len(list)-1] != "Zed" {
		t.Fatalf("roster %v", list)
	}
	if _, err := e.doNow(game.AddCharacter{Roster: len(list) - 1}); err != nil || e.partySize() != 1 {
		t.Fatalf("Add: %v, party %d", err, e.partySize())
	}
	id := e.partyIDs()[0]
	sh := e.ruleset().Sheet(id)
	if sh.Name != "Zed" || sh.Levels[sh.Classes[0]] != 2 || sh.Age != int32(d.Member.Age) {
		t.Fatalf("created: %+v", sh)
	}
	if its := e.itemsState(id); len(its) != 2 || its[0].Name != "Spacesuit" || its[1].Name != "Gun Bolt" || !its[0].Readied || !its[1].Readied || its[1].Count != 100 {
		t.Fatalf("the free items: %+v", its)
	}
	if _, err := e.doNow(game.Begin{}); err != nil {
		t.Fatal(err)
	}
}

// TestBR1CreateSkillPoints: a Medic made as a front makes it: the skill points spent through the engine's Spend
// action under the original's limits (skills.pas EditSkills(80, 40, 30)), then kept, added to the party, begun.
func TestBR1CreateSkillPoints(t *testing.T) {
	e := br1Engine(t)
	e.Open("")
	e.titleDone()
	e.Advance()
	state := func() *game.Draft { a, _ := e.build().(game.Assembling); return a.Draft }
	if err := e.createCharacter(); err != nil {
		t.Fatal(err)
	}
	if err := e.SetField(0, slices.Index(state().Fields[0].Options, "Tinker")); err != nil { // a Tinker is a Medic first
		t.Fatal(err)
	}
	d := state()
	if len(d.Fields) != 3 || len(d.Allotments) != 2 || d.Member.Class != "Medic" {
		t.Fatalf("fields %d, allotments %d, class %q", len(d.Fields), len(d.Allotments), d.Member.Class)
	}
	career, general := d.Allotments[0], d.Allotments[1]
	if career.Left != 80 || career.Max != 30 || len(career.Rows) != 8 || general.Left != 40 || general.Max != 30 {
		t.Fatalf("allotments %+v / %+v", career, general)
	}
	row := func(a game.Allotment, key rules.Skill) int {
		i := slices.IndexFunc(a.Rows, func(r game.AllotRow) bool { return r.Name == xxvc.SkillNames[key] })
		if i < 0 {
			t.Fatalf("no row %s", xxvc.SkillNames[key])
		}
		return i
	}
	spend := func(alloc, r, by int) error {
		_, err := e.doNow(game.Spend{Allotment: alloc, Row: r, By: by})
		return err
	}
	light, serious := row(career, "treat-light-wounds"), row(career, "treat-serious-wounds")
	if career.Rows[serious].Open || len(career.Rows[serious].Wants) != 1 || career.Rows[serious].Wants[0].Level != 30 {
		t.Fatalf("serious %+v", career.Rows[serious])
	}
	if err := spend(0, serious, 1); err == nil || err.Error() != "Prerequisites not satisfied" {
		t.Fatalf("serious before light: %v", err)
	}
	for range 30 {
		if err := spend(0, light, 1); err != nil {
			t.Fatal(err)
		}
	}
	if err := spend(0, light, 1); err == nil || err.Error() != "Skill at maximum for this level" {
		t.Fatalf("31 points: %v", err)
	}
	for range 4 {
		if err := spend(0, serious, 1); err != nil {
			t.Fatal(err)
		}
	}
	maths := row(general, "mathematics")
	if err := spend(1, maths, 1); err != nil {
		t.Fatal(err)
	}
	if err := spend(1, maths, 1); err != nil {
		t.Fatal(err)
	}
	if err := spend(1, maths, -1); err != nil {
		t.Fatal(err)
	}
	d = state()
	if d.Allotments[0].Left != 46 || d.Allotments[0].Rows[light].Value != 30 || d.Allotments[0].Rows[serious].Value != 4 ||
		d.Allotments[1].Left != 39 || d.Allotments[1].Rows[maths].Value != 1 || !d.Allotments[0].Rows[serious].Open {
		t.Fatalf("after spending: %+v", d.Allotments)
	}
	if err := e.rename("Doc"); err != nil {
		t.Fatal(err)
	}
	if err := e.EndCreation(true); err != nil {
		t.Fatal(err)
	}
	list := e.BuildCharList()
	if _, err := e.doNow(game.AddCharacter{Roster: len(list) - 1}); err != nil {
		t.Fatal(err)
	}
	sh := e.ruleset().Sheet(e.partyIDs()[0])
	if sh.Name != "Doc" || sh.Skills["treat-light-wounds"] != 30 || sh.Skills["treat-serious-wounds"] != 4 || sh.Skills["mathematics"] != 1 || sh.Skills["diagnose"] != 0 {
		t.Fatalf("the kept character's skills: %v", sh.Skills)
	}
	if _, err := e.doNow(game.Begin{}); err != nil {
		t.Fatal(err)
	}
}
