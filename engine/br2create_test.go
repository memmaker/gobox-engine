package engine

import (
	"slices"
	"testing"

	"gobox/game"
)

// TestBR2Create: a BR2 character made as a front makes it, kept, added and begun. BR2's chargen.pas differs from BR1's in
// the numbers the importer writes into xxvc.Creation: the character starts at level 7 with that level's experience and
// hit points rolled for seven levels, the career's starting skills are given (ApplyInitialSkills) and nothing is left
// to spend (EditSkills(0, 0, 105, 1)).
func TestBR2Create(t *testing.T) {
	e := brEngine(t, gameBR2)
	e.Open("")
	e.titleDone()
	e.Advance()
	state := func() *game.Draft { a, _ := e.build().(game.Assembling); return a.Draft }
	if err := e.createCharacter(); err != nil {
		t.Fatal(err)
	}
	d := state()
	if d == nil || len(d.Fields) != 3 {
		t.Fatalf("fields %+v", d)
	}
	if !slices.Equal(d.Fields[0].Options, []string{"Terran", "Martian", "Venusian", "Mercurian", "Tinker", "Desert Runner"}) { // CharGen's Pick Race: RaceTerran..RaceDesertRunner
		t.Fatalf("races %v", d.Fields[0].Options)
	}
	if err := e.SetField(0, slices.Index(d.Fields[0].Options, "Tinker")); err != nil { // medic or engineer only
		t.Fatal(err)
	}
	d = state()
	if d.Member.Class != "Medic" || d.Fields[2].Fits[slices.Index(d.Fields[2].Options, "Warrior")] {
		t.Fatalf("tinker career %+v", d.Fields[2])
	}
	if d.Member.MaxHP < 7 || d.Member.MaxHP > 7*(6+2) { // seven levels of the better of two d6, plus at most 2
		t.Fatalf("hit points of seven levels: %+v", d.Member)
	}
	for _, a := range d.Allotments {
		if a.Left != 0 {
			t.Errorf("BR2 leaves nothing to spend: %+v", a)
		}
	}
	if err := e.rename("Zed"); err != nil {
		t.Fatal(err)
	}
	if err := e.EndCreation(true); err != nil {
		t.Fatal(err)
	}
	list := e.BuildCharList()
	if _, err := e.doNow(game.AddCharacter{Roster: len(list) - 1}); err != nil || list[len(list)-1] != "Zed" {
		t.Fatalf("Add: %v %v", err, list)
	}
	sh := e.ruleset().Sheet(e.partyIDs()[0])
	if sh.Levels[sh.Classes[0]] != 7 || sh.Exp != 48000 /* a Medic's experienceNeeded[career, 7] */ || sh.Skills["diagnose"] != 15 || sh.Skills["treat-critical-wounds"] != 75 {
		t.Fatalf("created: level %d exp %d skills %v", sh.Levels[sh.Classes[0]], sh.Exp, sh.Skills)
	}
	if _, err := e.doNow(game.Begin{}); err != nil {
		t.Fatal(err)
	}
}
