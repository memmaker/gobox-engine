package playrules

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// TestEscIsExit: Esc (Cancel) on the rules' menu answers its Exit, as the
// originals' (BR2's fight View: GEAR SKILLS EXIT), never a silent nothing.
func TestEscIsExit(t *testing.T) {
	f := &fake{st: play.Fighting{Common: play.Common{Ask: &play.Question{Replies: []string{"GEAR", "SKILLS", "EXIT"}}}}}
	New(f, game.Data{}).Apply(game.Cancel{})
	if len(f.sent) != 1 || f.sent[0] != (play.Reply{N: 2}) {
		t.Fatalf("sent %+v", f.sent)
	}
}
