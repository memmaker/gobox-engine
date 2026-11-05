package classic

import (
	"testing"

	"gobox/game"
)

// The rules' NoSave/NoLoad close the slot screen's Save and Load (camp, F10, party menu).
func TestSlotsHonourRules(t *testing.T) {
	st := game.State{Mode: game.ModeCamp, Party: []game.Member{{Name: "A"}}}
	if v := (&Game{R: &loreRules{s: st}}).SlotsView(false); v == nil || !v.CanSave {
		t.Fatal("camp save not offered")
	}
	st.Flags = &game.Flags{NoSave: true}
	if v := (&Game{R: &loreRules{s: st}}).SlotsView(false); v == nil || v.CanSave {
		t.Fatal("NoSave still offers saving")
	}
	st.Flags.NoLoad = true
	if v := (&Game{R: &loreRules{s: st}}).SlotsView(false); v != nil {
		t.Fatal("NoSave+NoLoad still opens the slots")
	}
}
