package ui

import (
	"image/color"
	"testing"

	"gobox/game"
)

// layPage carries PaneRows to the pane and a Layer's Frames and FrameMs to its picture.
func TestPageRowsFrames(t *testing.T) {
	l := layPage(&game.Page{
		Pane: [][]game.TextRun{{{Text: "x"}}}, PaneRows: 4,
		Layers: []game.Layer{{Frames: []int{1, 2}, FrameMs: 80}}}, nil, [16]color.RGBA{})
	if l.pane.Rows != 4 || len(l.pictures[0].Frames) != 2 || l.pictures[0].FrameMs != 80 {
		t.Fatalf("%+v %+v", l.pane, l.pictures)
	}
}

// A game event with a Page is a timed moment holding that page.
func TestCombatEventsPage(t *testing.T) {
	pg := &game.Page{Layers: []game.Layer{{FrameMs: 1}}}
	out := played([]game.Report{{Ms: 60, Page: pg}}, false, nil, 0)
	if len(out) != 1 || out[0].Ms != 60 || out[0].Page != pg {
		t.Fatalf("%+v", out)
	}
}
