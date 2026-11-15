package engine

import (
	"gobox/game"
	"slices"
	"testing"
)

// TestModifyGreyedWithExperience: the party menu offers Modify only for a
// character with no experience of its own (core.Modifiable), not listed and refused.
func TestModifyGreyedWithExperience(t *testing.T) {
	e := quickStarted(t)
	if len(e.state().Party) < 2 {
		t.Skip("needs two party members")
	}
	e.state().Party[0].Exp, e.state().Party[1].Exp = 0, 1<<20
	e.push(cont{kind: kDrawMain})
	e.partyMenu(true, true)
	e.Advance()
	if !playAllows(e.build(), game.Modify{Member: 0}) {
		t.Error("Modify not offered for a member with no experience")
	}
	if playAllows(e.build(), game.Modify{Member: 1}) {
		t.Error("Modify offered for a member with experience of its own")
	}
	if _, err := e.doNow(game.Modify{Member: 1}); err == nil {
		t.Error("Modify taken for a member with experience")
	}
	// the hall's menu: Modify only while the selected member may be modified, never Drop
	for sel, want := range []bool{true, false} {
		e.selectMember(sel)
		fs := e.partyFunctions()
		if got := slices.Contains(fs, partyModify); got != want {
			t.Errorf("member %d selected: Modify on the menu %v, want %v", sel, got, want)
		}
		if slices.Contains(fs, partyDrop) {
			t.Error("the hall's menu has Drop")
		}
	}
	e.state().Party[0].OldLevels[0] = 3 // dual-classed: experience 0 again, but it has adventured
	if e.selectMember(0); slices.Contains(e.partyFunctions(), partyModify) {
		t.Error("Modify on the menu for a dual-classed member")
	}
}
