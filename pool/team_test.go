package pool

import (
	"gobox/game/play"
	"gobox/rules1e"
	"testing"
)

// TestTeamJoined (BUG-CLASSES 64): a party built from the menu has joined
// (ovr18:1446): PartyCount is the party's size before the first fight,
// every member is on the party's side with its own icon slot (Order < 8,
// COMSPR's, not CPIC's), and Quick Start over a party is refused.
func TestTeamJoined(t *testing.T) { teamJoined(t, testEngine(t)) }

// TestTeamJoinedCOTAB: the same on cotab's imported folder.
func TestTeamJoinedCOTAB(t *testing.T) { teamJoined(t, advEngine(t, advEnvCOTAB)) }

// teamJoined is TestTeamJoined on e.
func teamJoined(t *testing.T, e *Engine) {
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Init()
	e.Start(e.Play)
	playOn(e, 20, func(s play.State) bool { _, ok := s.(play.Assembling); return ok })
	if _, err := e.doNow(play.QuickStart{}); err != nil {
		t.Fatalf("QuickStart: %v", err)
	}
	if _, err := e.doNow(play.QuickStart{}); err == nil {
		t.Error("QuickStart over a party taken")
	}
	if n := len(e.state().Party); n == 0 || e.state().PartyCount() != n || len(commonOf(e.build()).Party) != n {
		t.Fatalf("party %d, PartyCount %d, published %d", n, e.state().PartyCount(), len(commonOf(e.build()).Party))
	}
	for _, c := range e.state().Party {
		if !rules1e.IsParty(c) || c.Order >= 8 || c.Control >= 0x80 {
			t.Errorf("%s: side %d order %d control %x, not joined", c.Name(), c.Side, c.Order, c.Control)
		}
	}
}
