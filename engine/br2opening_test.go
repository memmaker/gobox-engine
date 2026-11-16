package engine

import (
	"strings"
	"testing"

	"gobox/game"
)

// br2OpeningTexts are the pages block 19's first entry shows (area19.txt L_8508, GAME.OVR text of the originals; the old
// engine's TestOpening starts on the same block 19, geo 17) and the dock's line after the treasure screen.
var br2OpeningTexts = []string{
	"YOU EASE INTO ORBIT AROUND MERCURY",
	"'I KNOW YOU'RE ITCHING TO PULL MORE COMBAT DUTY",
	"'BERKELEY IS CALLING FOR A BROTHERHOOD",
	"'IF BERKELEY SUCCEEDS, WE WILL HAVE A UNITED FRONT AGAINST RAM.",
	"YOU STEP OUT ONTO THE EXPANSIVE DOCK",
}

// TestBR2Opening: BR2 on the shared engine from the title through the party menu: block 19's briefing pages (picture 57), the
// treasure screen of its no-monster Combat (TreasureOf(8000, ...): the credits are the party's at once, the pile is
// offered), then the dock in block 17 (area19.txt; buckrogers2-go adv_test.go TestOpening: ends in the explore state
// after the opening's loot is left).
func TestBR2Opening(t *testing.T) {
	e := brEngine(t, gameBR2)
	e.Open("")
	e.titleDone()
	e.Advance()
	for _, a := range []game.Action{game.AddCharacter{Roster: 0}, game.Begin{}} {
		if _, err := e.doNow(a); err != nil {
			t.Fatalf("%T: %v", a, err)
		}
	}
	start := e.purse(false).Copper
	var seen []string
	offered := 0
	for n := 0; n < 40; n++ {
		st := e.build()
		c := commonOf(st)
		if tx := paneText(c); tx != "" {
			seen = append(seen, tx)
		}
		if l, ok := st.(game.Looting); ok {
			if !playAllows(st, game.ContinueAction{}) {
				offered = len(l.Items)
			}
			if got := e.purse(false).Copper; got != start+8000 {
				t.Errorf("purse %d, want %d", got, start+8000)
			}
			br1AfterFight(e, st) // the pile taken and left
		} else if c.Ask != nil {
			br1Answer(t, e)
		} else {
			break
		}
	}
	if e.gs.areaKind != 17 {
		t.Fatalf("in block %d, want 17; saw %q; %T", e.gs.areaKind, seen, e.build())
	}
	if offered == 0 {
		t.Error("no treasure offered")
	}
	j := strings.Join(seen, "\n")
	for _, want := range br2OpeningTexts {
		if !strings.Contains(j, want) {
			t.Errorf("never showed %q; saw %q", want, seen)
		}
	}
}
