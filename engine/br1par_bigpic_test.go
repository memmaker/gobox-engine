package engine

import (
	"strings"
	"testing"

	"gobox/game"
)

// TestBR1BigPicture ports rules TestBigPicture: the opening's attack scene is a BIGPIC shown whole, framed over the
// screen (BR PICTURE blocks 112 and up): a Scene that is Full with the picture's own 304x120 pixels.
func TestBR1BigPicture(t *testing.T) {
	e := brEngine(t, gameBR1)
	e.Open("")
	e.titleDone()
	e.Advance()
	if _, err := e.doNow(game.AddCharacter{Roster: 0}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.doNow(game.Begin{}); err != nil {
		t.Fatal(err)
	}
	for range 40 {
		if sc, ok := e.build().(game.Scene); ok && sc.Picture != nil && strings.HasPrefix(sc.Picture.File, "BIGPIC") {
			if !sc.Full || sc.Picture.W != 304 || sc.Picture.H != 120 || len(sc.Picture.Pixels) != 304*120 {
				t.Fatalf("big picture %s %d: full %v, %dx%d, %d pixels", sc.Picture.File, sc.Picture.Block, sc.Full, sc.Picture.W, sc.Picture.H, len(sc.Picture.Pixels))
			}
			return
		}
		if commonOf(e.build()).Ask == nil {
			break
		}
		br1Answer(t, e)
	}
	t.Fatal("no big picture in the opening")
}
