package engine

import (
	"testing"

	"gobox/game"
)

// br2Ambush is area17's first encounter (data/br2/scripts/area17.txt:171): four monsters of record 46 on the party's
// square, a real BR2 set-up.
const br2Ambush = "    ClearMonsters()\n    LoadMonster(46, count=4, icon=46)\n    Combat()"

// TestBR2Shop: Shop(1) with BR2's templates (br1shop_test.go brShop).
func TestBR2Shop(t *testing.T) { brShop(t, gameBR2) }

// TestBR2Fight: a ground fight from BR2's own script set-up, played to its end as a front would (br1Fought): the
// party fights four monsters of record 46, every unit has its picture and every square its tile.
func TestBR2Fight(t *testing.T) {
	e, log, stop := brFoughtIn(t, gameBR2, br2Ambush, 0, br1TurnAction)
	if stop != "" {
		t.Fatalf("the fight did not play to its end: %s\nbattle log so far: %q", stop, log)
	}
	if _, ok := e.build().(game.Fighting); ok {
		t.Error("still Fighting after the play")
	}
	if len(log) < 2 {
		t.Errorf("the fight left a trace of %d turns", len(log))
	}
}

// TestBR2RestMenu: Program(9) is the rest menu in BR2 too (br1walls_test.go brRestMenu).
func TestBR2RestMenu(t *testing.T) { brRestMenu(t, gameBR2) }

// TestBR2SaveLoadExploring: a BR2 game saved after its opening loads back at the same square with the same party, purse, cells
// and clock (br1walls_test.go brSaveLoadExploring).
func TestBR2SaveLoadExploring(t *testing.T) { brSaveLoadExploring(t, gameBR2) }

// TestBR2RocketStart and TestBR2RocketStep: BR2's rocket view (script view kind 2) starts and steps as BR1's; the old BR2
// engine's TestRocketStep holds the same values (start 9,8 heading 87; east into the sun undone, counter 30 to 1, step 223).
func TestBR2RocketStart(t *testing.T) { brRocketStart(t, gameBR2) }
func TestBR2RocketStep(t *testing.T)  { brRocketStep(t, gameBR2) }

// br2SpaceFightScript is area19's own line (L_83E5, data/br2/scripts/area19.txt:154) SpaceFight(3, 1, V_7F7D, V_7F7E), with the
// two figures BR1's line holds (50, 18) in place of the reads of T_A3ED and T_A3F2, which are tables of block 19's own memory;
// then ShipOutcome is read as BR1's script does and a cell the test's block owns marks the script as having gone on.
const br2SpaceFightScript = "    SpaceFight(3, 1, 50, 18)\n    NewGameBytes = ShipOutcome\n    Word4BE7 = 77"

// TestBR2SpaceFightPlayed: BR2's ship fight from its own script line is played to its end by a front's answers; the script goes on.
func TestBR2SpaceFightPlayed(t *testing.T) {
	brSpaceFightPlayed(t, gameBR2, br2SpaceFightScript, "Word4BE7", "NewGameBytes")
}
