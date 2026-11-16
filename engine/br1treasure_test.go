package engine

import (
	"gobox/ecl"
	"strings"
	"testing"

	"gobox/game"
)

// TestBR1TreasureOf: block 33's line (area33.txt: ClearMonsters, TreasureOf(0, 5, 31 x5), Combat) with credits:
// the credits are the party's at once, the Combat with no monsters ends at the treasure screen, the items
// are offered and taken, and the script goes on after it.
func TestBR1TreasureOf(t *testing.T) {
	e := br1Begun(t)
	purse := func() uint64 { return e.purse(false).Copper }
	start := purse()
	br1OwnBlock(t, e, 203, "    ClearMonsters()\n    TreasureOf(300, 2, 31, 31)\n    Combat()\n    NewGameBytes = 77")
	if got := purse(); got != start+300 {
		t.Errorf("purse %d, want %d", got, start+300)
	}
	offered := 0
	for range br1FightCap {
		st := e.build()
		if l, ok := st.(game.Looting); ok && !playAllows(st, game.ContinueAction{}) {
			offered = len(l.Items)
		}
		if br1Cell(t, e, "NewGameBytes") == 77 {
			break
		}
		br1AfterFight(e, st) // takes every item, then leaves
	}
	if offered != 2 {
		t.Fatalf("%d items offered, want 2", offered)
	}
	if n := len(e.goods()); n != 0 {
		t.Errorf("%d items left", n)
	}
	if got := br1Cell(t, e, "NewGameBytes"); got != 77 {
		t.Errorf("the script stopped after the screen: %d", got)
	}
}

// TestBR1ExperienceText: BR's GainExperience prints BR's words (eclvm.pas TEAM_AWARD_TEXT, AWARD_PREFIX/SUFFIX),
// not the Pool's.
func TestBR1ExperienceText(t *testing.T) {
	e := br1Begun(t)
	br1OwnBlock(t, e, 204, "    GainExperience(1, 50)")
	if got := paneText(commonOf(e.build())); !strings.Contains(got, "Congrats. The team gains experience!") {
		t.Errorf("text %q", got)
	}
}

// TestBR1PrevSquare: Area[240], [241] are the square before the step (eclvm.pas: previousX/Y := the party's
// square after entry 1, before ResolveMovement): the one World.PrevX/PrevY, so a script reads them after a step.
func TestBR1PrevSquare(t *testing.T) {
	e := br1Begun(t)
	br1OwnBlock(t, e, 205, "    PartyX = 3\n    PartyY = 4")
	if br1Cell(t, e, "PrevX") != int64(e.world.PrevX) || br1Cell(t, e, "PrevY") != int64(e.world.PrevY) {
		t.Fatalf("the cells are not the world's: %d %d / %d %d", br1Cell(t, e, "PrevX"), br1Cell(t, e, "PrevY"), e.world.PrevX, e.world.PrevY)
	}
	e.world.PrevX, e.world.PrevY = 9, 8
	if br1Cell(t, e, "PrevX") != 9 || br1Cell(t, e, "PrevY") != 8 {
		t.Errorf("the cells read %d %d, want 9 8", br1Cell(t, e, "PrevX"), br1Cell(t, e, "PrevY"))
	}
}

// TestBR1NewEclInLoad: a start entry that loads another script (eclvm.pas NewEclDone repeats from entry 5 for
// each script loaded) hands over before the first one's move entry: the new script's own start entry runs.
func TestBR1NewEclInLoad(t *testing.T) {
	e := br1Begun(t)
	br1OwnBlock(t, e, 207, "    NewGameBytes = 5") // writes 207's file
	e.host.Set(func() ecl.Ref { r, _ := e.layoutField("NewGameBytes"); return r }(), 0)
	br1OwnBlock(t, e, 206, "    NewECL(207)")
	if got := br1Cell(t, e, "NewGameBytes"); got != 5 {
		t.Errorf("block 207's start entry did not run: %d", got)
	}
	if e.gs.areaKind != 207 {
		t.Errorf("in block %d", e.gs.areaKind)
	}
}
