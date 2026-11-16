package engine

import "testing"

// br2Blocks are BR2's script blocks (the 33 area files of its import: ECL1.DAX).
var br2Blocks = []int{1, 2, 16, 17, 18, 19, 20, 21, 24, 32, 33, 34, 35, 36, 37, 38, 48, 49, 50, 64, 65, 66, 80, 81, 82, 84, 95, 96, 97, 98, 112, 113, 114}

// br2SweepKnown are BR2's findings still open, as br1SweepKnown.
var br2SweepKnown = []struct{ Pattern, Why string }{
	{`^script: Word#B#\[#\] is past the field's # \(layout too short\) @ gobox/engine.\(\*Engine\).inField`, "open, owner b2/layout: area1.txt's variable editor (`Word4B00[&NewGameBytes_2] = NewGameBytes_4` after `if &NewGameBytes_2 <= 256`, a typed address) writes element 256 of Word4B00, whose layout has 256 elements (the SaveTable past Word4B00 REFACTOR.md l2 lists); only the typed 999..256 reaches it. Replay: GOBOX_SWEEP_GAME=br2 GOBOX_SWEEP_REPLAY=\"1/0/0,0/0 : r0 r0 x0 x1 x2 x0 x1 x2 r0 x3 x3 r1 r0 r1 x4 x0 x1 x2\""},
	{`^LOG: script: area 21: T_\w+\[\d+\] is past the field's \d+ entries`, "none: the entry skips the story, so Word4C93 (set by block 19's encounters, area19.txt:149,641,843) holds 0, NewGameBytes = Word4C93 - 1 is 255 and the tables are read at 252..254 as the original does (engine log, TableFaults); as BR1's"},
}

func TestBR2Sweep(t *testing.T) { sweepGameRun(t, gameBR2, br2Blocks, br2SweepKnown) }
