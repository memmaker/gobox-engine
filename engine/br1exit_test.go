package engine

import (
	"strings"
	"testing"

	"gobox/game"
)

// TestBR1ClassExitOneStep (old exitstep_test.go TestClassExitOneStep, bug class 72): one step off the open
// east edge of Venus's Lowlander town (disk 5, block 81, (15,0)) asks to leave at once, and Yes takes the
// party out to the jungle (block 80) with no second step. The step flags the layout's LeftMap (Party[725],
// adv.go:357 "off the map flagged before the step entry") before the move entry reads it; the data set
// (the layout's SymbolSet, Party[786], the old e.disk) is what picks block 81's GEO.
func TestBR1ClassExitOneStep(t *testing.T) {
	const disk, block, jungle = 5, 81, 80
	e := brBegun(t, gameBR1)
	r, ok := e.layoutField("SymbolSet") // the script that sends the party here writes it first (eclmenu.pas:584 PushState)
	if !ok {
		t.Fatal("no SymbolSet in the layout")
	}
	e.host.Set(r, disk)
	if int(e.set) != disk {
		t.Fatalf("SymbolSet = %d left the data set at %d", disk, e.set)
	}
	e.push(cont{kind: kStepEnd}) // as after a step whose script changed the block
	e.DevAt(block, 8, 8)         // inland: arriving on the edge would run the exit at once
	br1Pages(t, e)               // the block's first-visit text
	e.posX, e.posY = 15, 0       // the load entry placed the party at the town gate (8,15)
	if e.DoorAt(15, 0, 2) == 0 {
		t.Fatal("the east edge at (15,0) is shut")
	}
	e.facing = 2
	if _, err := e.doNow(game.RunEvent{Before: true, Side: 2}); err != nil { // the step's move entry, before the step
		t.Fatal(err)
	}
	if e.world.LeftMap == 0 {
		t.Fatal("the step off the edge did not flag LeftMap before the move entry")
	}
	asked := false
	for n := 0; n < 12 && e.gs.areaKind == block; n++ {
		c := commonOf(e.build())
		var a game.Action = game.Reply{}
		if c.Ask != nil && strings.Contains(paneText(c), "LEAVE") {
			asked, a = true, game.Reply{N: 1}
		} else if c.Ask == nil {
			break
		}
		if _, err := e.doNow(a); err != nil {
			t.Fatalf("%T: %v", a, err)
		}
	}
	if !asked {
		t.Fatalf("the step off the edge did not ask to leave: %T %q", e.build(), paneText(commonOf(e.build())))
	}
	if e.gs.areaKind != jungle {
		t.Fatalf("after Yes in block %d, want %d", e.gs.areaKind, jungle)
	}
}

// TestBR1ClassEdgeFlagBeforeStep (old exit_test.go TestClassEdgeFlagBeforeStep, bug class 72): the step off the
// map flags the layout's LeftMap (Party[725], pO09E_0000 0FE4) before the step entry runs, so the move entry
// reads 1 on that same step: block 81's entry asks to leave, a script's own flag read (V_7ED5 == 1).
func TestBR1ClassEdgeFlagBeforeStep(t *testing.T) {
	e := brBegun(t, gameBR1)
	r, _ := e.layoutField("SymbolSet")
	e.host.Set(r, 5)
	e.push(cont{kind: kStepEnd})
	e.DevAt(81, 8, 8)
	br1Pages(t, e)
	left, ok := e.layoutField("LeftMap")
	if !ok {
		t.Fatal("no LeftMap in the layout")
	}
	e.posX, e.posY, e.facing = 15, 0, 2
	e.world.LeftMap = 0
	e.checkEdge()
	if got := e.host.Get(left); got != 1 {
		t.Fatalf("the layout's LeftMap reads %d after a step off the edge, want 1", got)
	}
	e.posX, e.posY = 14, 0 // inside: the flag goes down again
	e.checkEdge()
	if got := e.host.Get(left); got != 0 {
		t.Fatalf("LeftMap reads %d inside the map, want 0", got)
	}
}
