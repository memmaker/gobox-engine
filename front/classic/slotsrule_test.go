package classic

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// The rules' NoSave/NoLoad close the slot screen's Save and Load (camp, F10, party menu).
func TestSlotsHonourRules(t *testing.T) {
	camp := func(noSave, noLoad bool) *Game {
		return NewGame(&fake{st: play.Camping{Common: play.Common{Party: []game.Member{{Name: "A"}}, NoSave: noSave, NoLoad: noLoad}}}, game.Data{SaveDir: t.TempDir()})
	}
	if v := camp(false, false).SlotsView(false); v == nil || !v.CanSave {
		t.Fatal("camp save not offered")
	}
	if v := camp(true, false).SlotsView(false); v == nil || v.CanSave {
		t.Fatal("NoSave still offers saving")
	}
	if v := camp(true, true).SlotsView(false); v != nil {
		t.Fatal("NoSave+NoLoad still opens the slots")
	}
}
