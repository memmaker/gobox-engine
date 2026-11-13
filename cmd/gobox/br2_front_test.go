package main

import (
	"testing"

	"gobox/ui"

	"buckrogers2/rules"
)

// TestCreateSheetRows: gobox's creation sheet on the shipped path
//
//	shows the adnd2-br sheet's seven abilities ending in TECH
//
// and the rules' figures, as BR1's (play.Draft.Stats, Derived).
func TestBR2CreateSheetRows(t *testing.T) {
	e, err := rules.New(br2Adv(), t.TempDir(), 1)
	if err != nil {
		t.Skip(err)
	}
	d := e.Data()
	g := ui.NewGame(e, d)
	g.TitleDone()
	if err := g.CreateCharacter(); err != nil {
		t.Fatal(err)
	}
	dr := g.Draft()
	if dr == nil || len(dr.Stats) != 7 || dr.Stats[6].Label != "TECH" {
		t.Fatalf("sheet abilities %+v", dr)
	}
	if len(dr.Derived) == 0 || dr.Derived[0].Label != "Age" {
		t.Fatalf("sheet figures %+v", dr.Derived)
	}
}
