package playrules

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// TestViewMemberOpensSheet: View (a shop's, temple's) is front state: the
// member's sheet comes up with no call to the game, and CloseSheet ends it.
func TestViewMemberOpensSheet(t *testing.T) {
	f := &fake{st: play.Exploring{Common: play.Common{Party: []play.Member{{Name: "ARIN"}}}}}
	r := New(f, game.Data{})
	r.Apply(game.ViewMember{Member: 0})
	if r.State().SheetPage == nil || len(f.sent) != 0 {
		t.Fatalf("sheet %v, sent %v", r.State().SheetPage != nil, f.sent)
	}
	r.Apply(game.CloseSheet{})
	if r.State().SheetPage != nil {
		t.Fatal("sheet still up")
	}
}

// Camp with no AdvanceTimeBy allowed offers no timed rest (Camp.NoRest).
func TestCampNoRest(t *testing.T) {
	for _, allow := range []bool{false, true} {
		st := play.Camping{}
		if allow {
			st.Allowed.Plain = []play.Action{play.AdvanceTimeBy{}}
		}
		if got := New(&fake{st: st}, game.Data{}).State().Allowed["NoRest"]; got == allow {
			t.Errorf("AdvanceTimeBy allowed %v: NoRest %v", allow, got)
		}
	}
}
