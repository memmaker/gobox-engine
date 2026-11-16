package engine

import (
	"testing"

	"gobox/game"
	"gobox/rules"
	"gobox/rules1e"
)

// TestPoolTakeBack: a Pool fight with a member on Quick. game.Manual between turns is the typed rules.TakeBack, the one
// way for every ruleset: pressed, it waits for the next computer turn (por combatai.pas CheckKeystroke polls there), whose
// first poll puts every player character that is no NPC and not animated back under manual control (quickfight := 0), and
// the key is spent. Alt-M is the typed rules.QuickMagic and flips the computer's casting at that same poll.
func TestPoolTakeBack(t *testing.T) {
	e := devFirstTurn(t, "losefight", false)
	if !myTurn(e.build()) {
		t.Fatal("no first turn")
	}
	if _, err := e.doNow(game.Quick{}); err != nil {
		t.Fatal(err)
	}
	autos := func() (n int) {
		for _, c := range e.state().Party {
			if c.Auto != 0 && c.Control < 0x80 && c.Status != rules1e.StatusAnimated {
				n++
			}
		}
		return n
	}
	if autos() == 0 {
		t.Fatal("Quick put no member on the computer")
	}
	sent := false
	for range 300 {
		s := e.build()
		if _, ok := s.(game.Fighting); !ok {
			t.Fatalf("the fight ended (%T) before Manual was offered", s)
		}
		if !myTurn(s) && e.waiting(PendFight) && playAllows(s, game.Manual{}) {
			if _, err := e.doNow(game.Manual{}); err != nil {
				t.Fatal(err)
			}
			if _, ok := e.aiKey.(rules.TakeBack); !ok {
				t.Fatalf("Manual left the key %T, want rules.TakeBack", e.aiKey)
			}
			if autos() == 0 {
				t.Fatal("the members were taken back before the next computer turn polled")
			}
			sent = true
			break
		}
		if myTurn(s) { // the player's members end their turns; the Quick one and the monsters are the computer's
			if err := e.TurnChoice(TurnAct{Verb: VerbEnd}, false); err != nil {
				t.Fatal(err)
			}
		} else if e.waiting(PendFight) {
			e.Step()
		} else {
			press(e)
		}
	}
	if !sent {
		t.Fatal("Manual was never offered between turns")
	}
	for range 300 { // the next computer turn polls first (aiTurn), then the key is spent
		if e.aiKey == nil {
			break
		}
		s := e.build()
		if _, ok := s.(game.Fighting); !ok {
			t.Fatalf("the fight ended (%T) with the key unspent", s)
		}
		if myTurn(s) {
			if err := e.TurnChoice(TurnAct{Verb: VerbEnd}, false); err != nil {
				t.Fatal(err)
			}
		} else if e.waiting(PendFight) {
			e.Step()
		} else {
			press(e)
		}
	}
	if e.aiKey != nil || autos() != 0 || e.state().Fight.Key != rules1e.AiKeyNone {
		t.Fatalf("after the poll: key %v, %d quick members, rules' key %d; want none, 0, none", e.aiKey, autos(), e.state().Fight.Key)
	}
}
