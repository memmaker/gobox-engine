package ui

import (
	"testing"

	"gobox/game"
	"gobox/gpu"
)

// TestBlankWaitShown: a key wait with nothing on the text pane (POR's
// Stojanow Gate arrows: Text(" "), then opDamage's any-key pause when every
// arrow missed) draws a prompt; it was an invisible input block.
func TestBlankWaitShown(t *testing.T) {
	full := func(byte) [8]byte { return [8]byte{255, 255, 255, 255, 255, 255, 255, 255} }
	note := Note{Glyph: full, Font: func(byte, bool) [8]byte { return full(0) }}
	press := &game.Question{Replies: []string{"Press"}, Keys: []rune{'P'}}
	party := []game.Member{{Name: "ANNA", HP: 8, MaxHP: 8}}
	walls := &game.MapWalls{Side: [][][4]int{{{0, 0, 0, 0}}}} // over the 3D view, as at the gate
	st := game.Scene{Common: game.Common{Adventure: true, Party: party, Text: []string{" ", "", "", "", "", ""}, Ask: press}, Walls: walls}
	dst := gpu.NewImage(1280, 800)
	defer dst.Deallocate()
	n := 0
	drawnAt = func(what drawnKind, on *gpu.Image, _ float64) {
		if on == dst && what == drawnMenu {
			n++
		}
	}
	defer func() { drawnAt = nil }()
	p, _ := presenterOn(st)
	p.Draw(dst, note)
	if n == 0 {
		t.Fatal("a key wait with a blank pane drew nothing")
	}
}
