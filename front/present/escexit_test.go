package present

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// TestEscIsExit: Esc (Cancel) on the rules' menu answers the reply its
// Esc names (BR2's fight View: GEAR SKILLS EXIT), never a silent nothing.
func TestEscIsExit(t *testing.T) {
	f := &fake{st: play.Fighting{Common: play.Common{Ask: &play.Question{Replies: []string{"GEAR", "SKILLS", "EXIT"}, Esc: play.EscTo(2)}}}}
	newRules(f, game.Data{}).send(play.Cancel{})
	if len(f.sent) != 1 || f.sent[0] != (play.Reply{N: 2}) {
		t.Fatalf("sent %+v", f.sent)
	}
}

// TestEscNoReply: a menu naming no Esc reply is not answered by Esc, even
// with an entry labelled Exit (labels drive nothing).
func TestEscNoReply(t *testing.T) {
	f := &fake{st: play.Fighting{Common: play.Common{Ask: &play.Question{Replies: []string{"Gear", "Exit"}}}}}
	newRules(f, game.Data{}).send(play.Cancel{})
	if len(f.sent) != 0 {
		t.Fatalf("sent %+v", f.sent)
	}
}

// TestEscToRange: EscTo names the index; the zero value names none.
func TestEscToRange(t *testing.T) {
	if _, ok := play.NoEscReply.Reply(); ok {
		t.Fatal("zero EscReply answers a reply")
	}
	if i, ok := play.EscTo(0).Reply(); !ok || i != 0 {
		t.Fatalf("EscTo(0) = %d %v", i, ok)
	}
}
