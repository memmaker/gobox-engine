package engine

import (
	"strings"
	"testing"

	"gobox/game"
	"gobox/rules"
)

// br1AtTurn begins the fight of block 201's body and plays on until a party member has the turn.
func br1AtTurn(t *testing.T, body string, facing int) (*Engine, game.Fighting) {
	t.Helper()
	e := br1Begun(t)
	br1Join(t, e)
	e.facing = facing
	br1OwnBlock(t, e, 201, body)
	for range br1FightCap {
		s := e.build()
		f, ok := s.(game.Fighting)
		switch {
		case !ok:
			t.Fatalf("the fight ended before a party turn: %T", s)
		case myTurn(s):
			return e, f
		case e.waiting(PendFight):
			e.Step()
		default:
			press(e)
		}
	}
	t.Fatal("no party turn")
	return nil, game.Fighting{}
}

func TestBR1TurnUndeadNotOffered(t *testing.T) {
	e, f := br1AtTurn(t, br1AdjacentFight, 0)
	if playAllows(f, game.TurnUndead{}) {
		t.Fatal("BR offers Turn")
	}
	if _, err := e.doNow(game.TurnUndead{}); err == nil {
		t.Error("TurnUndead accepted")
	}
	// the front's raw turn choice too: turnChoice takes only the verbs the ruleset offered (Turn.Can), so no
	// rules.TurnUndead is resolved and nothing panics
	if err := e.TurnChoice(TurnAct{Verb: VerbTurn}, false); err != nil {
		t.Errorf("TurnChoice: %v", err)
	}
	if _, ok := e.build().(game.Fighting); !ok || !myTurn(e.build()) {
		t.Error("the refused Turn used the turn")
	}
}

// br1Turns plays on from a party turn with choose until stop says so, returning the (round, unit) of each party
// turn seen (the one it starts at first).
func br1Turns(t *testing.T, e *Engine, f game.Fighting, choose func(game.Fighting) game.Action, stop func(game.State) bool) [][2]int {
	t.Helper()
	var seen [][2]int
	for range br1FightCap {
		s := e.build()
		if stop(s) {
			return seen
		}
		ff, ok := s.(game.Fighting)
		switch {
		case ok && myTurn(s):
			seen = append(seen, [2]int{ff.Field.Round, ff.Field.Turn})
			if _, err := e.doNow(choose(ff)); err != nil {
				t.Fatalf("%T refused: %v", choose(ff), err)
			}
		case ok && e.waiting(PendFight):
			e.Step()
		default:
			press(e)
		}
	}
	t.Fatal("the play did not reach its stop")
	return nil
}

// TestBR1Guard: Guard (cbtacts.pas Guard: the actions end, the unit guards) uses the turn; play goes on to the
// next unit and round.
func TestBR1Guard(t *testing.T) {
	e, f := br1AtTurn(t, br1AdjacentFight, 0)
	round, unit := f.Field.Round, f.Field.Turn
	if !playAllows(f, game.Guard{}) {
		t.Fatal("Guard not offered")
	}
	if _, err := e.doNow(game.Guard{}); err != nil {
		t.Fatal(err)
	}
	if s := e.build(); myTurn(s) && s.(game.Fighting).Field.Turn == unit && s.(game.Fighting).Field.Round == round {
		t.Error("the unit still has its turn after Guard")
	}
}

// TestBR1Delay: Delay ('W': initiative 1) sends the unit to the end of the round's order: it acts again in the
// same round, after the others.
func TestBR1Delay(t *testing.T) {
	e, f := br1AtTurn(t, br1AdjacentFight, 0)
	round, unit := f.Field.Round, f.Field.Turn
	delayed := false
	seen := br1Turns(t, e, f, func(f game.Fighting) game.Action {
		if !delayed {
			delayed = true
			if !playAllows(f, game.Delay{}) {
				t.Fatal("Delay not offered")
			}
			return game.Delay{}
		}
		return game.EndTurn{}
	}, func(s game.State) bool {
		_, ok := s.(game.Fighting)
		return !ok
	})
	n := 0
	for _, x := range seen {
		if x == [2]int{round, unit} {
			n++
		}
	}
	if n < 2 {
		t.Errorf("unit %d had %d turns in round %d, want its delayed second: %v", unit, n, round, seen)
	}
}

// br1WalkWest is the move to the offered square nearest the west edge (X 0), the unit's own when none is nearer.
func br1WalkWest(f game.Fighting) (game.Action, bool) {
	var me rules.BoardPos
	for _, u := range f.Field.Units {
		if u.ID == f.Field.Turn {
			me = u.At
		}
	}
	best := me
	for _, m := range f.Allowed.Moves {
		if m.At.X < best.X {
			best = m.At
		}
	}
	return game.MoveTo{To: best}, best != me
}

// br1FleeChoice walks west to the arena's edge, then Flee answered Yes ("Flee:", combat.pas MoveMenu): the move
// off the map, AttemptFlee. A blocked escape ("Escape is blocked") uses the turn too; the next one tries again.
func br1FleeChoice(t *testing.T, e *Engine) func(game.Fighting) game.Action {
	return func(f game.Fighting) game.Action {
		if a, ok := br1WalkWest(f); ok {
			return a
		}
		if _, err := e.doNow(game.Flee{}); err != nil {
			t.Fatalf("Flee at the edge: %v", err)
		}
		ask := commonOf(e.build()).Ask
		if ask == nil || ask.Text != "Flee:" && !strings.Contains(paneText(commonOf(e.build())), "Flee:") {
			t.Fatalf("Flee asked %+v, want the question \"Flee:\"", ask)
		}
		return game.Reply{N: 1} // Yes
	}
}

// TestBR1Flee: every member walks to the edge and flees (cbtacts.pas AttemptFlee: out of play as fleeing, 'Got
// Away'; a blocked one 'Escape is blocked'); when all are gone the fight is over (PartyFled), no experience, and the
// party list is the members: nobody was lost.
func TestBR1Flee(t *testing.T) {
	e, f := br1AtTurn(t, br1FarFight, br1FarFacing)
	var members []rules.CharID
	for _, id := range e.partyIDs() {
		if e.fact(id, rules.FactSide) != rules.SideMonsters {
			members = append(members, id)
		}
	}
	br1Turns(t, e, f, br1FleeChoice(t, e), func(s game.State) bool { _, ok := s.(game.Fighting); return !ok })
	for _, id := range members {
		if !e.ext.log.out[id] {
			t.Errorf("member %d not out of the fight, log %q", id, e.ext.log.lines)
		}
	}
	if got := len(e.partyIDs()); got != len(members) {
		t.Errorf("party list %d, want the %d members (the monster left it)", got, len(members))
	}
	for _, id := range members {
		if st, _ := e.ruleset().Fact(id, rules.FactStatus); st != rules.StatusOkay {
			t.Errorf("member %d status %v after the flight: the Pascal sets a fled member back to okay (postcbt.pas)", id, st)
		}
	}
}

// TestBR1FleeOne: one member flees and leaves the board ("<name> Got Away", its unit gone from the state) while the
// others stay and the fight goes on.
func TestBR1FleeOne(t *testing.T) {
	e, f := br1AtTurn(t, br1FarFight, br1FarFacing)
	runner := f.Field.Turn
	runnerID := e.unitID(runner) // looked up while it is on the board
	flee := br1FleeChoice(t, e)
	onBoard := func(s game.State) bool {
		ff, ok := s.(game.Fighting)
		if !ok {
			return false
		}
		for _, u := range ff.Field.Units {
			if u.ID == runner {
				return true
			}
		}
		return false
	}
	br1Turns(t, e, f, func(f game.Fighting) game.Action {
		if f.Field.Turn == runner {
			return flee(f)
		}
		return game.EndTurn{}
	}, func(s game.State) bool { return !onBoard(s) })
	s, ok := e.build().(game.Fighting)
	if !ok {
		t.Fatalf("the fight ended with one member fled: %T", e.build())
	}
	if !e.ext.log.out[runnerID] {
		t.Errorf("the runner %d is not out of the fight, log %q", runnerID, e.ext.log.lines)
	}
	if len(s.Field.Units) < 2 {
		t.Errorf("%d units left on the board, want the other member and the monster", len(s.Field.Units))
	}
}

// TestBR1QuickOne: Quick for one member (combat.pas TurnOnAuto): its turn is played by the computer from then on
// (its FightCard is Auto, it is not asked again), the other member is still asked. The key back to manual (Space,
// CheckKeystroke) is not reachable: see the report (Manual is offered only while a computer turn waits).
func TestBR1QuickOne(t *testing.T) {
	e, f := br1AtTurn(t, br1AdjacentFight, 0)
	quick := f.Field.Turn
	if !playAllows(f, game.Quick{}) {
		t.Fatal("Quick not offered")
	}
	if _, err := e.doNow(game.Quick{}); err != nil {
		t.Fatal(err)
	}
	if !e.fact(e.unitID(quick), rules.FightCardFact{}).(rules.FightCard).Auto {
		t.Error("the member is not computer-steered after Quick")
	}
	asked := map[int]int{}
	br1Turns(t, e, game.Fighting{}, func(f game.Fighting) game.Action {
		asked[f.Field.Turn]++
		return game.EndTurn{}
	}, func(s game.State) bool { _, ok := s.(game.Fighting); return !ok })
	if asked[quick] != 0 || len(asked) == 0 {
		t.Errorf("asked per unit after Quick: %v (quick member %d)", asked, quick)
	}
}
