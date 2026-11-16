package ui

import (
	"testing"

	"gobox/game"
	"gobox/gpu"
)

// TestBlankWaitShown: a key wait with nothing on the text pane (POR's
// Stojanow Gate arrows: Text(" "), then opDamage's any-key pause when every
// arrow missed) draws a prompt, over the 3D view and on a page; it was an
// invisible input block. The stuck check (keyWait) catches the class: the
// same wait drawn with nothing is stuck.
func TestBlankWaitShown(t *testing.T) {
	full := func(byte) [8]byte { return [8]byte{255, 255, 255, 255, 255, 255, 255, 255} }
	note := Note{Glyph: full, Font: func(byte, bool) [8]byte { return full(0) }}
	party := []game.Member{{Name: "ANNA", HP: 8, MaxHP: 8}}
	blank := []string{" ", "", "", "", "", ""}
	for name, q := range map[string]*game.Question{
		"press":    {Replies: []string{"Press"}, Keys: []rune{'P'}},
		"continue": {Continue: true},
	} {
		for _, walls := range []*game.MapWalls{{Side: [][][4]int{{{0, 0, 0, 0}}}}, nil} {
			st := game.Scene{Common: game.Common{Adventure: true, Party: party, Text: blank, Ask: q}, Walls: walls}
			dst := gpu.NewImage(1280, 800)
			p, _ := presenterOn(st)
			p.Draw(dst, note)
			if p.stuck() {
				t.Errorf("%s, 3D %v: a key wait with a blank pane on no screen (%s)", name, walls != nil, p.asksID())
			}
			dst.Deallocate()
		}
	}
}

// TestKeyWaitStuck: a key wait with a blank pane is stuck when the frame
// showed nothing of it (the check the headless runs make, OnStuck), and
// not once its message is in the pane.
func TestKeyWaitStuck(t *testing.T) {
	p, _ := presenterOn(game.Scene{Common: game.Common{Adventure: true, Text: []string{" "}, Ask: &game.Question{Continue: true}}})
	p.markSeen() // a frame that drew nothing
	if !p.stuck() {
		t.Error("an undrawn key wait is not stuck")
	}
	p.waitShown = true
	p.markSeen()
	if p.stuck() {
		t.Error("a key wait with its message shown is stuck")
	}
}
