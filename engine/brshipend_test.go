package engine

import (
	"path/filepath"
	"strings"
	"testing"

	"gobox/game"
	"gobox/rules"
)

// A ship fight leaves zone 0's shipCombatOutcome (ShipOutcome) and postCombatMode (PostCombatMode) as shipcbt.pas
// does: the foe destroyed 0/NORMAL (392, 1217), the team's hull lost 1/NORMAL (396), boarding 0/BOARDED (1105, 1106;
// C_PostCb 1). The block's script takes the Pascal's branch: BR1 area34 L_8A4D and BR2 area19 L_83E5 board
// (NewECL) on `if 1 == PostCombatMode`; BR2 reads ShipOutcome after it.

const (
	shipBranchBoards = 52 // the branch cell's value when `if 1 == PostCombatMode` is taken (stands for NewECL(52))
	shipBranchLost   = 9  // BR2's `if ShipOutcome != 1` taken
)

// shipPrefer answers with the first of the names a question lists (reply 0 when none).
func shipPrefer(names ...string) func(q *game.Question) int {
	return func(q *game.Question) int {
		for _, n := range names {
			if k := indexOf(q, n); k >= 0 {
				return k
			}
		}
		return 0
	}
}

var (
	shipQuitPolicy  = shipPrefer("Quit")
	shipBoardPolicy = shipPrefer("Board", "Close", "Engn", "Target", "Fire", "Command") // stop the foe's engines at range, close, board
)

// brShipEnd plays a ship fight in a block of the test's own, whose lines after the fight are the real block's
// condition lines (the file is checked to still have them), and checks the cells and the branch.
func brShipEnd(t *testing.T, g brGame, setup string, pick func(q *game.Question) int, wantOutcome, wantMode int64) {
	t.Helper()
	e := brBegun(t, g)
	cell, call, area := "Word4CE2", "    Word4C61 = 1\n    SpaceFight(Word4C61, 1, 50, 18)\n", "area34.txt"
	if g.ID == gameBR2.ID {
		cell, call, area = "Word4BE7", "    SpaceFight(2, 1, 50, 18)\n", "area19.txt" // foe 3 rams; 2 can be stopped
	}
	text := readFile(t, filepath.Join(e.scriptDir(), area))
	lines := "    if 1 == PostCombatMode {\n        " + cell + " = 52\n    }\n"
	if !strings.Contains(text, "if 1 == PostCombatMode {") {
		t.Fatalf("%s no longer reads PostCombatMode after its SpaceFight", area)
	}
	if g.ID == gameBR2.ID {
		if !strings.Contains(text, "if ShipOutcome != 1 {") {
			t.Fatalf("%s no longer reads ShipOutcome", area)
		}
		lines += "    if ShipOutcome != 1 {\n        NewGameBytes = 9\n    }\n"
	}
	br1OwnBlock(t, e, 201, setup+"    "+cell+" = 0\n    NewGameBytes = 0\n"+call+lines)
	if _, ok := shipFightOn(e); !ok {
		t.Fatalf("no ship fight up: %T", e.State())
	}
	if got := br1Cell(t, e, "PostCombatMode"); got != 0 {
		t.Fatalf("PostCombatMode %d as the fight begins, want NORMAL 0", got)
	}
	answerShip(t, e, pick)
	if got := br1Cell(t, e, "ShipOutcome"); got != wantOutcome {
		t.Errorf("ShipOutcome %d, want %d", got, wantOutcome)
	}
	if got := br1Cell(t, e, "PostCombatMode"); got != wantMode {
		t.Errorf("PostCombatMode %d, want %d", got, wantMode)
	}
	wantBranch := int64(0)
	if wantMode == 1 {
		wantBranch = shipBranchBoards
	}
	if got := br1Cell(t, e, cell); got != wantBranch {
		t.Errorf("the branch cell %d, want %d", got, wantBranch)
	}
	if g.ID == gameBR2.ID {
		wantLost := int64(0)
		if wantOutcome != 1 {
			wantLost = shipBranchLost
		}
		if got := br1Cell(t, e, "NewGameBytes"); got != wantLost {
			t.Errorf("BR2's ShipOutcome branch %d, want %d", got, wantLost)
		}
	}
}

func testShipEnds(t *testing.T, g brGame) {
	t.Run("ends", func(t *testing.T) { // every typed end, as the encodings turn it into the cells
		e := brBegun(t, g)
		for _, c := range []struct {
			end           rules.ShipEnd
			outcome, mode int64
		}{
			{rules.ShipFoeDestroyed, 0, 0}, {rules.ShipBoarded, 0, 1}, {rules.ShipHullLost, 1, 0}, {rules.ShipReceded, 1, 0},
		} {
			e.writeShipOutcome(c.end)
			e.writePostCombatMode(c.end)
			if o, m := br1Cell(t, e, "ShipOutcome"), br1Cell(t, e, "PostCombatMode"); o != c.outcome || m != c.mode {
				t.Errorf("ship end %d: ShipOutcome %d PostCombatMode %d, want %d %d", c.end, o, m, c.outcome, c.mode)
			}
		}
	})
	t.Run("lost", func(t *testing.T) { brShipEnd(t, g, "    ShipHull = 1\n", shipQuitPolicy, 1, 0) })
	t.Run("boarded", func(t *testing.T) { brShipEnd(t, g, "", shipBoardPolicy, 0, 1) })
}

func TestBR1ShipEnds(t *testing.T) { testShipEnds(t, gameBR1) }
func TestBR2ShipEnds(t *testing.T) { testShipEnds(t, gameBR2) }
