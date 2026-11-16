package engine

import (
	"testing"

	"gobox/game"
)

// TestBR2WallsBackdrop: BR2 on the shared engine (BACK1 only, flats buckrogers2-go rules/gamewalls.go BackFlats, tested as
// BR1's TestBR1WallsBackdrop): the base's streets without a BACK picture show the city's flats; with BACK1 block 17
// loaded last a square with bit 7 (indoors) shows that picture and the haze's flats.
func TestBR2WallsBackdrop(t *testing.T) {
	e := brBegun(t, gameBR2)
	w := br1Walls(t, e)
	if w.Floor != "br1_city" || w.Sky != "sky_br1_city" || w.Backdrop != nil {
		t.Errorf("no picture: floor %q sky %q backdrop %v", w.Floor, w.Sky, w.Backdrop != nil)
	}
	e.LoadWalldef(1, 1)
	e.LoadWalldef(17, 2)
	e.gs.geo.Plane[0x200+e.posY<<4+e.posX] |= 0x80
	w = br1Walls(t, e)
	if w.Floor != "br1_mars_haze" || w.Sky != "sky_br1_mars_haze" || w.Backdrop == nil || w.Backdrop.Block != 17 || w.Backdrop.File != "BACK1" {
		t.Errorf("indoors: floor %q sky %q backdrop %+v", w.Floor, w.Sky, w.Backdrop)
	}
	e.gs.geo.Plane[0x200+e.posY<<4+e.posX] &^= 0x80
	w = br1Walls(t, e)
	if w.Floor != "br1_city" || w.Backdrop != nil { // 17+128 is no BACK1 block
		t.Errorf("outdoors: floor %q backdrop %v", w.Floor, w.Backdrop != nil)
	}
}

// TestBR2PlaceMusic: Where.Music is the block's place (buckrogers2-go blockMusic): Caloris in block 17, a town (the dock after the opening).
func TestBR2PlaceMusic(t *testing.T) {
	e := brBegun(t, gameBR2)
	if w := commonOf(e.build()).Where; w.Music != "caloris" || w.Mode != "town" {
		t.Errorf("block 17: music %q mode %q", w.Music, w.Mode)
	}
}

// TestBR2ScriptSound: block 19's Sound(130) is above 127: buckrogers2-re eclvm.pas EclPlaySound starts music sequence 2 there, not a
// sound command, so no cue reaches the front (as BR1's limit brSoundLimit; the old BR2 engine's op 67 reports every byte, a
// deviation from the Pascal); Sound(5) is the cue "5".
func TestBR2ScriptSound(t *testing.T) {
	e := brEngine(t, gameBR2)
	sounds := e.game().ScriptSounds
	if c := sounds[5].Cue; c != "5" {
		t.Errorf("Sound(5): cue %q", c)
	}
	if _, ok := sounds[130]; ok {
		t.Error("Sound(130) has a cue")
	}
	e.Open("")
	e.titleDone()
	e.Advance()
	for _, a := range []game.Action{game.AddCharacter{Roster: 0}, game.Begin{}} {
		out, err := e.doNow(a)
		if err != nil {
			t.Fatalf("%T: %v", a, err)
		}
		for _, ev := range out.Events {
			if s, ok := ev.(game.Sounded); ok {
				t.Errorf("the opening sounded %q", s.Sound)
			}
		}
	}
}
