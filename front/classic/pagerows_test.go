package classic

import (
	"testing"

	"gobox/front"
	"gobox/game"
)

// Page carries PaneRows to the pane and a Layer's Frames and FrameMs to its picture.
func TestPageRowsFrames(t *testing.T) {
	v := Page(front.View{}, game.State{Page: &game.Page{
		Pane: [][]game.TextRun{{{Text: "x"}}}, PaneRows: 4,
		Layers: []game.Layer{{Frames: []int{1, 2}, FrameMs: 80}}}})
	if v.Pane.Rows != 4 || len(v.Pictures[0].Frames) != 2 || v.Pictures[0].FrameMs != 80 {
		t.Fatalf("%+v %+v", v.Pane, v.Pictures)
	}
}
