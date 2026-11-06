package present

import (
	"testing"

	"image/color"

	"gobox/front"
	"gobox/gpu"
)

// TestMenuOverLastScene: a shop's or temple's page that publishes no map
// keeps the 3D scene last shown to draw over, never black.
func TestMenuOverLastScene(t *testing.T) {
	var p Presenter
	w := &front.Walls{Key: "town", Side: [][][4]int{{{0, 0, 0, 0}}}}
	shown := p.scene(front.View{Walls: w, Exploring: true})
	if p.scene(front.View{Temple: &front.Temple{}}) != nil || p.behind != shown {
		t.Fatal("the temple's page lost the town's scene")
	}
}

// TestMenuOverBackdrop: a shop's or temple's menu (no map, no picture of its
// own) draws over the location as last shown, never over black.
func TestMenuOverBackdrop(t *testing.T) {
	dst := gpu.NewImage(1280, 800)
	defer dst.Deallocate()
	full := func(byte) [8]byte { return [8]byte{255, 255, 255, 255, 255, 255, 255, 255} }
	note := Note{Glyph: full, Font: func(byte, bool) [8]byte { return full(0) }}
	red := color.RGBA{200, 0, 0, 255}
	loc := &front.View{Pane: &front.Pane{Lines: [][]front.PaneRun{{{Text: "THE TOWN SQUARE.", Fg: red, Bg: color.RGBA{A: 255}}}}}} // a drawn location (the front leaves its text out; any lit frame will do)
	for _, v := range []front.View{{Temple: &front.Temple{}}, {Shop: &front.Shop{}}} {
		p := &Presenter{}
		p.Draw(dst, note, v)
		alone := lit(dst)
		p = &Presenter{backView: loc}
		p.Draw(dst, note, v)
		if over := lit(dst); over <= alone {
			t.Fatalf("%+v: %d lit over the location, %d alone", v, over, alone)
		}
	}
}
