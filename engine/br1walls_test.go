package engine

import (
	"testing"

	"gobox/game"
)

func br1Walls(t *testing.T, e *Engine) game.MapWalls {
	t.Helper()
	x, ok := e.build().(game.Exploring)
	if !ok {
		t.Fatalf("%T", e.build())
	}
	return x.Map.Walls
}

// TestBR1WallsBackdrop: the tarmac (block 16, set 1, walls 7 then 127: no BACK1 picture 255) shows the
// city's flats without a picture, as the old front did (buckrogers-go rules TestClassStreetOutside); in
// set 4 with BACK4 block 10 loaded last, outdoors (cell without bit 7: block 138) the picture and the mars
// clouds' flats show (rules TestBackdrop, TestBackFlats), indoors (bit 7: block 10) the haze's.
func TestBR1WallsBackdrop(t *testing.T) {
	e := br1Begun(t)
	if w := br1Walls(t, e); w.Floor != "br1_city" || w.Sky != "sky_br1_city" || w.Backdrop != nil {
		t.Errorf("tarmac: floor %q sky %q backdrop %v", w.Floor, w.Sky, w.Backdrop != nil)
	}
	e.set = 4
	e.LoadWalldef(4, 1)
	e.LoadWalldef(10, 2)
	w := br1Walls(t, e)
	if w.Floor != "br1_mars_clouds" || w.Sky != "sky_br1_mars_clouds" || w.Backdrop == nil || w.Backdrop.Block != 138 || w.Backdrop.File != "BACK4" {
		t.Errorf("set 4 outdoors: floor %q sky %q backdrop %+v", w.Floor, w.Sky, w.Backdrop)
	}
	e.gs.geo.Plane[0x200+e.posY<<4+e.posX] |= 0x80
	w = br1Walls(t, e)
	if w.Floor != "br1_mars_haze" || w.Backdrop == nil || w.Backdrop.Block != 10 {
		t.Errorf("set 4 indoors: floor %q backdrop %+v", w.Floor, w.Backdrop)
	}
}

// TestBR1PlaceMusic: Where.Music is the block's place (the old front's musicPlace): Neo-Chicago on the tarmac
// (block 16), Salvation in the base's menu (block 17), whose streets are a town.
func TestBR1PlaceMusic(t *testing.T) {
	e := br1Begun(t)
	if w := commonOf(e.build()).Where; w.Music != "neo-chicago" || w.Mode == "town" {
		t.Errorf("block 16: music %q mode %q", w.Music, w.Mode)
	}
	br1ToBaseMenu(t, e)
	if w := commonOf(e.build()).Where; w.Music != "salvation" || w.Mode != "town" {
		t.Errorf("block 17: music %q mode %q", w.Music, w.Mode)
	}
}

// TestBR1ScriptSound: the intro's Sound(5) (in a For loop) reaches the front as the cue "5", Sound(131) as none (the old engine's
// command 67 reports the numbers below 128).
func TestBR1ScriptSound(t *testing.T) {
	e := br1Engine(t)
	e.Open("")
	e.titleDone()
	e.Advance()
	var heard []string
	note := func(out game.Outcome) {
		for _, ev := range out.Events {
			if s, ok := ev.(game.Sounded); ok {
				heard = append(heard, s.Sound)
			}
		}
	}
	for _, a := range []game.Action{game.AddCharacter{Roster: 0}, game.Begin{}} {
		out, err := e.doNow(a)
		if err != nil {
			t.Fatalf("%T: %v", a, err)
		}
		note(out)
	}
	for n := 0; commonOf(e.build()).Ask != nil && n < 80; n++ {
		note(br1AnswerOut(t, e))
	}
	for _, h := range heard { // Sound(5) is in the intro's For loop, once a pass
		if h != "5" {
			t.Errorf("the intro sounded %q", h)
		}
	}
	if len(heard) == 0 {
		t.Error("the intro sounded nothing")
	}
}

// TestBR1DelayMs: a script's Delay pauses 100 ms a unit of the game speed (BR textrend.pas PauseOnly), none at speed 0.
func TestBR1DelayMs(t *testing.T) {
	for _, c := range []struct{ speed, ms int }{{3, 300}, {0, 0}} {
		e := br1Begun(t)
		e.msgDelay = byte(c.speed)
		br1OwnBlock(t, e, 201, "    Delay()")
		got := 0
		if e.eclWait != nil {
			got = e.eclWait.ms
		}
		if got != c.ms {
			t.Errorf("speed %d: pause %d ms, want %d", c.speed, got, c.ms)
		}
	}
}

// TestBR1ViewKeptAcrossBlocks: a block whose start does not set the view kind keeps the overland (BR's EclInit leaves the
// interface state alone; the Pool's InitAreaScript sets the 3D view, which a Pool game still gets).
func TestBR1ViewKeptAcrossBlocks(t *testing.T) {
	e := br1Begun(t)
	if e.world.In3D != 1 {
		t.Fatalf("a new game's view: In3D %d, want the 3D view", e.world.In3D)
	}
	e.world.In3D = 0 // the overland
	e.InitAreaScript()
	if e.world.In3D != 0 {
		t.Errorf("a block's start set In3D %d over the overland", e.world.In3D)
	}
	v := *e.eclVersion // a copy: the dialect is shared
	v.Behaviour.ViewKept = false // as the Pool's
	e.eclVersion = &v
	e.InitAreaScript()
	if e.world.In3D != 1 {
		t.Errorf("without ViewKept the start leaves In3D %d", e.world.In3D)
	}
}
