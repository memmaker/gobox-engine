package pool

import (
	"gobox/asset"
	"gobox/core"
	"gobox/core/goldbox"
	"gobox/ecl/script"
	"testing"
)

// runOp runs op as the dialect v's instruction with these immediate operands.
func runOp(e *Engine, v *script.Version, op func(), args ...int) {
	e.eclVersion = v
	in := script.Instr{}
	for _, a := range args {
		in.Args = append(in.Args, script.Arg{N: a})
	}
	e.vm.Script = &script.Script{Code: []script.Instr{in}, Layout: e.vm.Script.Layout}
	e.vm.PC = 0
	e.vm.Fetch()
	op()
}

// TestLoadAreaOp: pod's LoadFiles (map, walls) loads the 3D map, and none from 127 on.
func TestLoadAreaOp(t *testing.T) {
	e := quickStarted(t)
	e.setWorldWord(core.WordIn3D, 1)
	e.setWorldWord(core.WordGeo, 9)
	runOp(e, script.V3, e.opLoadArea, geoNone, wallsNone)
	if g := e.worldWord(core.WordGeo); g != 9 {
		t.Fatalf("map 127 loaded: geo %d", g)
	}
	runOp(e, script.V3, e.opLoadArea, 1, wallsNone)
	if g := e.worldWord(core.WordGeo); g != 1 {
		t.Fatalf("map 1: geo %d", g)
	}
}

// TestSetupMonsterSides: pod's 4th operand is the sides; outdoors the asked distance becomes 2.
func TestSetupMonsterSides(t *testing.T) {
	e := quickStarted(t)
	e.setWorldWord(core.WordIn3D, 0)
	runOp(e, script.V3, e.opSetupMonster, 1, 1, 1, 5)
	if e.run.monsterSides != 5 || e.vm.PC != 1 {
		t.Fatalf("sides %d, PC %d", e.run.monsterSides, e.vm.PC)
	}
	if d := e.worldWord(core.WordMaxDist); d != maxEncounterDist {
		t.Fatalf("max distance %d", d)
	}
}

// TestNewECLSavesLastBlock: NewECL keeps the block it leaves in the dialect's slot (WordLastBlock).
func TestNewECLSavesLastBlock(t *testing.T) {
	e := quickStarted(t)
	e.gs.areaKind = 7
	runOp(e, script.V1, e.opNewECL, 1)
	if b := e.worldWord(core.WordLastBlock); b != 7 {
		t.Fatalf("last block %d", b)
	}
}

// TestNoMonsterCombat: Combat with no monsters opens the dialect's first screen whose word is set, clearing it.
func TestNoMonsterCombat(t *testing.T) {
	e := quickStarted(t)
	e.eclVersion = script.V1
	e.setWorldWord(core.WordTemple, 1)
	if e.noMonsterScreen() == nil || e.worldWord(core.WordTemple) != 0 {
		t.Fatal("temple not opened")
	}
	if e.noMonsterScreen() != nil {
		t.Fatal("opened with no word set")
	}
	if got := script.V3.Behaviour.NoMonsterCombat; len(got) != 2 || got[0] != script.CombatVault || got[1] != script.CombatTemple {
		t.Fatalf("pod: %v", got)
	}
}

// TestAreaMonsterFor: the area screen picks the monster; past the table none.
func TestAreaMonsterFor(t *testing.T) {
	a := AreaMonsters{ByScreen: []AreaMonster{{59, 59}, {6, 6}}}
	if m, ok := a.areaMonsterFor(1); !ok || m.Monster != 6 {
		t.Fatalf("screen 1: %v %v", m, ok)
	}
	if _, ok := a.areaMonsterFor(2); ok {
		t.Fatal("screen 2 has a monster")
	}
}

// TestInputStringMax: the line is as long as the text window, or pod's operand.
func TestInputStringMax(t *testing.T) {
	e := quickStarted(t)
	runOp(e, script.V1, e.opInputString, 0, 0)
	if m := e.eclWait.maxlen; m != textCols {
		t.Fatalf("POR max %d", m)
	}
	runOp(e, script.V3, e.opInputString, 12, 0)
	if m := e.eclWait.maxlen; m != 12 {
		t.Fatalf("pod max %d", m)
	}
}

// TestSelectMemberNoDrop: pod masks bit 7 and drops no one.
func TestSelectMemberNoDrop(t *testing.T) {
	e := quickStarted(t)
	n := e.partySize()
	e.es.memberDied, e.es.memberGone = 1, 1
	runOp(e, script.V3, e.opLoadCharacter, 0x81)
	if e.sel != 1 || e.partySize() != n || e.es.memberDied != 1 {
		t.Fatalf("sel %d, party %d", e.sel, e.partySize())
	}
}

// TestAddNPCIcon: the dialect flag loads the NPC's combat icon.
func TestAddNPCIcon(t *testing.T) {
	e := quickStarted(t)
	g := e.game()
	g.AddNPCIcon = true
	e.given = &g
	for d := goldbox.Set(1); d <= 8 && refDAX(e).BlockShared(e.set.File("MON")+"CHA", 5) == nil; d++ {
		e.set = d // a set that has monster 5
	}
	runOp(e, script.V3, e.opAddNPC, 5, 100)
	if e.icons[e.partyIndex(e.selected())] == [2]*asset.Image{} {
		t.Fatal("no icon")
	}
}

// TestPodDelayAndNextFrame: pod's byte 58 is Delay (the game's pause), and its CALL 3 steps the
// view picture's frame, wrapping at the frame count.
func TestPodDelayAndNextFrame(t *testing.T) {
	if c := script.V3.Cmds[58]; c.Op != script.OpDelay || c.Fixed != 0 {
		t.Fatalf("pod byte 58: %+v, want Delay with no operands", c)
	}
	e := quickStarted(t)
	im := func() asset.Frame { return asset.Frame{Image: &asset.Image{Width: 1, Height: 1}} }
	e.pic.frames, e.pic.cur = []asset.Frame{im(), im(), im()}, 2
	for _, want := range []int{3, 1, 2} {
		runOp(e, script.V3, e.opCall, 3)
		if e.pic.cur != want {
			t.Fatalf("frame %d, want %d", e.pic.cur, want)
		}
	}
	e.pic.frames = nil
	runOp(e, script.V3, e.opCall, 3) // no picture: nothing to step
}

// TestGameTablesWords: the layout fields GameTables.Words names are the game's last block (NewECL)
// and DIVIDE's remainder; a game without them keeps the ruleset's word and no remainder; a name
// the layout lacks is a loud failure.
func TestGameTablesWords(t *testing.T) {
	e := quickStarted(t)
	lay := &script.Layout{Fields: []script.Field{
		{Name: "SaveBlock", Owner: script.OwnerGame, Width: script.WidthByte, From: 21, Len: 1},
		{Name: "Rest", Owner: script.OwnerGame, Width: script.WidthByte, From: 40, Len: 1},
	}}
	e.vm.Script.Layout = lay
	e.bindWords()
	if e.vm.Remainder != (script.Ref{}) {
		t.Fatal("remainder bound without Words")
	}
	e.given = &GameTables{Words: ScriptWords{LastBlock: "SaveBlock", Remainder: "Rest"}}
	e.bindWords()
	rest, _ := lay.Named("Rest")
	if e.vm.Remainder != rest {
		t.Fatalf("remainder %+v, want %+v", e.vm.Remainder, rest)
	}
	e.setLastBlock(7) // opNewECL's
	save, _ := lay.Named("SaveBlock")
	if b := e.getVar(save); b != 7 {
		t.Fatalf("last block field %d, want 7", b)
	}
	e.given = &GameTables{Words: ScriptWords{Remainder: "Nope"}}
	defer func() {
		if recover() == nil {
			t.Error("a Words name the layout lacks did not fail")
		}
	}()
	e.bindWords()
}
