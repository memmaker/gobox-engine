package engine

import (
	"gobox/rules"
	"testing"

	"gobox/game"
)

// TestWatchFightTurn: Quick Start, camp in Phlan, rest until the watch
// comes, Stay: the party gets a turn before the fight ends, on the
// original's skewed maps and deskewed (where the foes once stood among
// the party at distance 0 and killed it before anyone acted, seed 1).
func TestWatchFightTurn(t *testing.T) {
	for _, deskew := range []bool{false, true} {
		for _, sd := range []uint32{1, 2} {
			e := testEngine(t)
			e.Deskew.Store(deskew)
			e.FixedSeed = &sd
			e.Open("")
			playOn(e, 10, func(s game.State) bool { _, ok := s.(game.Assembling); return ok })
			e.doNow(game.QuickStart{})
			e.doNow(game.Begin{})
			exploring := func(s game.State) bool { x, ok := s.(game.Exploring); return ok && x.Ask == nil }
			if !playOn(e, 3000, exploring) {
				t.Fatalf("seed %d: not exploring (%T)", sd, e.build())
			}
			e.doNow(game.MakeCamp{})
			if _, err := e.doNow(game.CampRest{Kind: RestWait}); err != nil {
				t.Fatal("CampRest:", err)
			}
			watch := func(s game.State) bool { q := commonOf(s).Ask; return q != nil && len(q.Replies) == 2 }
			for i := 0; i < 50 && !watch(e.build()); i++ { // rest on, an hour at a time, till the watch comes
				if _, ok := e.build().(game.Camping); ok {
					e.doNow(game.CampRest{Kind: RestWait})
				} else if !playOn(e, 100, watch) {
					break
				}
			}
			q := commonOf(e.build()).Ask
			if q == nil || len(q.Replies) != 2 || q.Replies[1] != "Stay" {
				t.Fatalf("deskew %v seed %d: no watch (%+v)", deskew, sd, q)
			}
			e.doNow(game.Reply{N: 1})
			if f, ok := e.build().(game.Fighting); deskew && ok { // BUG-CLASSES 29: distance 0, the sides apart
				if in, box := foesInParty(f.Field.Units); in > 0 {
					t.Errorf("seed %d: %d of the watch inside the party's box %v", sd, in, box)
				}
			}
			turn := playOn(e, 400, func(s game.State) bool {
				f, ok := s.(game.Fighting)
				return ok && f.Field.Turn != 0 && e.cs.actor != rules.NoChar
			})
			if !turn {
				t.Errorf("deskew %v seed %d: the watch fight ended with no party turn", deskew, sd)
			}
		}
	}
}
