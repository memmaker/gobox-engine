package engine

import (
	"slices"
	"testing"

	"gobox/game"
	"gobox/rules"
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
	if sh.Name != "Zed" || sh.Levels[sh.Classes[0]] != 2 {
		t.Fatalf("created: %+v", sh)
	}
	if its := e.itemsState(id); len(its) != 2 || its[0].Name != "Spacesuit" || its[1].Name != "Gun Bolt" || !its[0].Readied || !its[1].Readied || its[1].Count != 100 {
		t.Fatalf("the free items: %+v", its)
	}
	if _, err := e.doNow(game.Begin{}); err != nil {
		t.Fatal(err)
	}
}
