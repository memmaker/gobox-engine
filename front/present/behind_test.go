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
// own) draws over the last frame of the location, never over black.
func TestMenuOverBackdrop(t *testing.T) {
	dst := gpu.NewImage(1280, 800)
	defer dst.Deallocate()
	full := func(byte) [8]byte { return [8]byte{255, 255, 255, 255, 255, 255, 255, 255} }
	note := Note{Glyph: full, Font: func(byte, bool) [8]byte { return full(0) }}
	red := color.RGBA{200, 0, 0, 255}
	for _, v := range []front.View{{Temple: &front.Temple{}}, {Shop: &front.Shop{}}} {
		p := &Presenter{backdrop: gpu.NewImage(1280, 800)}
		p.backdrop.Fill(red)
		p.Draw(dst, note, v)
		if r, g, b, _ := dst.At(2, 2).RGBA(); r>>8 != 200 || g|b != 0 {
			t.Fatalf("%+v: the menu's corner is %v, not the location behind it", v, dst.At(2, 2))
		}
	}
}
