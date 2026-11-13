package ui

import (
	"testing"

	"gobox/game"
)

// layPage carries a picture's Frames and FrameMs to its picture, and lays
// a line's runs out one after the other in their inks.
func TestPageRowsFrames(t *testing.T) {
	ink := egaInk
	l := layPage(&TextPage{
		View:   [][]game.TextRun{{{Text: "ab", Fg: game.InkYellow}, {Text: "c"}}},
		Layers: []Picture{{Frames: []int{1, 2}, FrameMs: 80}}}, nil, ink)
	if len(l.pictures[0].Frames) != 2 || l.pictures[0].FrameMs != 80 {
		t.Fatalf("%+v", l.pictures)
	}
	if v := l.viewText[0]; v[0].Fg != ink[14] || v[1].X != 2 || v[1].Fg != col(inkGreen) {
		t.Fatalf("%+v", v)
	}
}

// A game event with a typed screen is a timed moment holding it.
func TestCombatEventsScreen(t *testing.T) {
	sc := &game.Screen{}
	out := played([]game.Report{{Ms: 60, Screen: sc}}, false, nil, 0)
	if len(out) != 1 || out[0].Ms != 60 || out[0].Screen != sc {
		t.Fatalf("%+v", out)
	}
}
