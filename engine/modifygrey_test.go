package engine

import (
	"gobox/game/play"
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
	if !playAllows(e.build(), play.Modify{Member: 0}) {
		t.Error("Modify not offered for a member with no experience")
	}
	if playAllows(e.build(), play.Modify{Member: 1}) {
		t.Error("Modify offered for a member with experience of its own")
	}
	if _, err := e.doNow(play.Modify{Member: 1}); err == nil {
		t.Error("Modify taken for a member with experience")
	}
}
