package engine

import (
	"fmt"
	"strings"
	"testing"

	"gobox/game"
	"gobox/rules"
)

// spaceFightScript is BR1's own line (area34's L_8A4D) between a cell that marks the script as having begun and a
// cell that marks it as having gone on.
const spaceFightScript = "    Word4C61 = 1\n    Word4CE0 = 5\n    SpaceFight(Word4C61, 1, 50, 18)\n    Word4CE2 = 77\n    NewGameBytes = ShipOutcome"

const fightSteps = 400 // a fight that does not end in this many answers is stuck

// shipFightOf is the fight's state, or ok false when no ship fight is up.
func shipFightOn(e *Engine) (game.ShipFight, bool) {
	e.rebuild()
	f, ok := e.State().(game.ShipFight)
	return f, ok
}

// answerShip answers what the fight asks as a front would: pick names the reply to take (first match; Continue
// is always answered); the transcript is the fight's message window lines.
func answerShip(t *testing.T, e *Engine, pick func(q *game.Question) int) (steps int, transcript string) {
	t.Helper()
	var log strings.Builder
	for ; steps < fightSteps; steps++ {
		f, ok := shipFightOn(e)
		if !ok {
			return steps, log.String()
		}
		q := f.Ask
		if q == nil {
			t.Fatalf("the fight asks nothing: %+v", f.Fight)
		}
		for _, l := range f.Fight.Log {
			fmt.Fprintln(&log, l.Text)
		}
		var a game.Action = game.ContinueAction{}
		if !q.Continue {
			a = game.Reply{N: pick(q)}
		}
		if _, err := e.doNow(a); err != nil {
			t.Fatalf("answer %v (%v): %v", a, q.Replies, err)
		}
	}
	t.Fatalf("the fight is not over after %d answers", fightSteps)
	return
}

func indexOf(q *game.Question, name string) int {
	for i, r := range q.Replies {
		if r == name {
			return i
		}
	}
	return -1
}

// TestBR1SpaceFightPlayed: SpaceFight holds the script, the state is the ship fight with both ships, and a front's
// answers play it to its end; the script goes on and reads ShipOutcome.
func TestBR1SpaceFightPlayed(t *testing.T) {
	brSpaceFightPlayed(t, gameBR1, spaceFightScript, "Word4CE2", "NewGameBytes")
}

// brSpaceFightPlayed is TestBR1SpaceFightPlayed on game g with the block's script, the cell it sets once the fight is over
// (to 77) and the cell it reads ShipOutcome into.
func brSpaceFightPlayed(t *testing.T, g brGame, script, doneCell, stateCell string) {
	e := brBegun(t, g)
	br1OwnBlock(t, e, 201, script)
	f, ok := shipFightOn(e)
	if !ok {
		t.Fatalf("state %T", e.State())
	}
	if f.Fight.Own.Max[0] == 0 || f.Fight.Enemy.Max[0] == 0 || f.Fight.Range != 8 || !f.NoSave || f.Ask == nil || f.Ask.Text == "" {
		t.Fatalf("the fight's first state: %+v ask %+v", f.Fight, f.Ask)
	}
	if got := br1Cell(t, e, doneCell); got == 77 {
		t.Fatal("the script went on while the fight was up")
	}
	steps, _ := answerShip(t, e, func(q *game.Question) int {
		for _, n := range []string{"Fire", "Close", "Command", "Quit"} { // BR2's pilot reaches Close through Command
			if i := indexOf(q, n); i >= 0 {
				return i
			}
		}
		return 0
	})
	t.Logf("%d answers", steps)
	if got := br1Cell(t, e, doneCell); got != 77 {
		t.Fatalf("the script did not go on: %d", got)
	}
	if got := br1Cell(t, e, stateCell); got != 0 && got != 1 {
		t.Fatalf("ShipOutcome %d", got)
	}
}

// TestBR1SpaceFightSeed: the same seed plays the same fight.
func TestBR1SpaceFightSeed(t *testing.T) {
	play := func() string {
		e := br1Begun(t)
		br1OwnBlock(t, e, 201, spaceFightScript)
		steps, log := answerShip(t, e, func(q *game.Question) int { return max(indexOf(q, "Fire"), 0) })
		return fmt.Sprintf("%d %d\n%s", steps, br1Cell(t, e, "NewGameBytes"), log)
	}
	a, b := play(), play()
	if a != b {
		t.Fatalf("two fights differ:\n%s\n--\n%s", a, b)
	}
}

// TestBR1SpaceFightLost: a party that only quits against a hull of 1 loses its ship (or the foe leaves): ShipOutcome
// is 1 as the original's, and the script's winner branch is not taken.
func TestBR1SpaceFightLost(t *testing.T) {
	e := br1Begun(t)
	br1OwnBlock(t, e, 201, "    ShipHull = 1\n"+spaceFightScript)
	answerShip(t, e, func(q *game.Question) int { return max(indexOf(q, "Quit"), 0) })
	if got := br1Cell(t, e, "NewGameBytes"); got != 1 {
		t.Fatalf("ShipOutcome %d after a lost fight", got)
	}
	if got := br1Cell(t, e, "Word4CE2"); got != 77 {
		t.Fatalf("the script did not go on: %d", got)
	}
}

// TestBR1SpaceFightSheet: View shows the own ship's sheet over the fight until a key, the menu comes back; the
// sheet function builds the same record from the ruleset outside a fight; the foe's has no record outside one.
func TestBR1SpaceFightSheet(t *testing.T) {
	e := br1Begun(t)
	br1OwnBlock(t, e, 201, spaceFightScript)
	f, _ := shipFightOn(e)
	if _, err := e.doNow(game.Reply{N: indexOf(f.Ask, "View")}); err != nil {
		t.Fatal(err)
	}
	f, _ = shipFightOn(e)
	if f.Sheet == nil || !f.Sheet.Own || f.Sheet.Name == "" || f.Sheet.Tonnage == 0 || f.Ask == nil || !f.Ask.Continue {
		t.Fatalf("the sheet: %+v ask %+v", f.Sheet, f.Ask)
	}
	own, ok := e.ShipSheetState(true, true)
	if !ok || own.Name != f.Sheet.Name || own.Speed != f.Sheet.Speed {
		t.Fatalf("ShipSheetState %+v %v", own, ok)
	}
	if foe, ok := e.ShipSheetState(false, false); !ok || foe.Name == "" || foe.Known {
		t.Fatalf("the foe's sheet %+v %v", foe, ok)
	}
	if _, err := e.doNow(game.ContinueAction{}); err != nil {
		t.Fatal(err)
	}
	if f, _ = shipFightOn(e); f.Sheet != nil || f.Ask == nil || f.Ask.Continue {
		t.Fatalf("the menu did not come back: %+v", f.Ask)
	}
	answerShip(t, e, func(q *game.Question) int { return max(indexOf(q, "Quit"), 0) })
	if _, ok := e.ShipSheetState(false, false); ok {
		t.Fatal("a foe's sheet outside a fight")
	}
	var _ rules.ShipRules // the sheet is the ruleset's
}

// TestBR1SpaceFightScene: the scene shows the foe's picture (SHIPS, block = ship id mod 5: saveload.pas LoadShipData) and the
// manoeuvre's text (gdconsts.pas maneuverNames, 'Closing    ' with the foe not retreating); Target is one submenu of the
// seven systems (shipcbt.pas 'Target which system?') whose choice fires; the pilot's menu names the pilot first.
func TestBR1SpaceFightScene(t *testing.T) {
	e := br1Begun(t)
	br1OwnBlock(t, e, 201, spaceFightScript)
	f, ok := shipFightOn(e)
	if !ok {
		t.Fatalf("state %T", e.State())
	}
	const foeShip, shipPictures = 1, 5 // the script's SpaceFight(.., 1, ..): ship id 0; SHIPS block 0 % 5
	if sf := f.Fight; sf.Foe.File != "SHIPS" || sf.Foe.Block != (foeShip-1)%shipPictures || sf.Status != "Closing    " {
		t.Errorf("scene: foe %+v status %q", sf.Foe, sf.Status)
	}
	// walk the crew until a menu offers Target (a member with a weapon): its choices are the systems, Esc-less
	for range fightSteps {
		f, ok = shipFightOn(e)
		if !ok {
			t.Fatal("the fight ended before a menu offered Target")
		}
		if i := indexOf(f.Ask, "Target"); i >= 0 {
			if indexOf(f.Ask, "Hull") >= 0 {
				t.Fatalf("the systems are not in the top menu: %v", f.Ask.Replies)
			}
			if _, err := e.doNow(game.Reply{N: i}); err != nil {
				t.Fatal(err)
			}
			f, _ = shipFightOn(e)
			if f.Ask == nil || f.Ask.Text != "Target which system?" || len(f.Ask.Replies) != 7 || f.Ask.Replies[0] != "Hull" || f.Ask.Replies[6] != "Weap" {
				t.Fatalf("the submenu: %+v", f.Ask)
			}
			if _, err := e.doNow(game.Reply{N: 0}); err != nil {
				t.Fatal(err)
			}
			return
		}
		var a game.Action = game.ContinueAction{}
		if !f.Ask.Continue {
			a = game.Reply{N: max(indexOf(f.Ask, "Quit"), 0)}
		}
		if _, err := e.doNow(a); err != nil {
			t.Fatal(err)
		}
	}
	t.Fatal("no menu offered Target")
}

// TestBR1SpaceFightPilotNarrated: a menu of the pilot is preceded by the narrated line "X is your pilot." (shipcbt.pas BuildPlayerMenu).
func TestBR1SpaceFightPilotNarrated(t *testing.T) {
	e := br1Begun(t)
	br1OwnBlock(t, e, 201, spaceFightScript)
	pilotMenus := 0
	for range fightSteps {
		f, ok := shipFightOn(e)
		if !ok {
			break
		}
		if n := e.ruleset().Allowed(e.ships().ShipState()).Notice; n != "" && !f.Ask.Continue { // "X is your pilot." is narrated before the prompt
			if l := f.Fight.Log; len(l) == 0 || l[len(l)-1].Text != n || !strings.HasSuffix(n, " is your pilot.") {
				t.Errorf("the pilot's menu is not narrated: %q in %v", n, l)
			}
			pilotMenus++
		}
		var a game.Action = game.ContinueAction{}
		if !f.Ask.Continue {
			a = game.Reply{N: max(indexOf(f.Ask, "Command"), indexOf(f.Ask, "Quit"), 0)} // take over as pilot, else pass
		}
		if _, err := e.doNow(a); err != nil {
			t.Fatal(err)
		}
	}
	if pilotMenus == 0 {
		t.Error("no menu of a pilot came up")
	}
}
