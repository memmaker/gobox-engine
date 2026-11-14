package ui

import (
	"gobox/game"
	"gobox/gpu"
	"image/color"
	"testing"
)

// TestMenuOverLastScene: a shop's or temple's page that publishes no map
// keeps the 3D view it was entered from (wallsOf), never black.
func TestMenuOverLastScene(t *testing.T) {
	var p Presenter
	g, f := playGame(game.Exploring{Map: game.ExploreMap{Walls: game.MapWalls{Side: [][][4]int{{{0, 0, 0, 0}}}}}})
	p.SetEngine(g)
	if p.scene() == nil {
		t.Fatal("the town showed no scene")
	}
	f.st = game.Trading{}
	if p.scene() == nil {
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
	red := []color.RGBA{{200, 0, 0, 255}}
	loc := game.Scene{Pictures: []game.Picture{{File: "LOC", W: 64, H: 64, Pixels: make([]byte, 64*64), Palette: red, Clear: -1}}} // a drawn location
	for _, menu := range []game.State{templeState(game.Service{}), shopState()} {
		g, _ := playGame(menu)
		p := &Presenter{}
		p.SetEngine(g)
		p.Draw(dst, note)
		alone := lit(dst)
		g, f := playGame(loc)
		p = &Presenter{}
		p.SetEngine(g)
		p.Draw(dst, note)
		f.st = menu
		p.Draw(dst, note)
		if over := lit(dst); over <= alone {
			t.Fatalf("%T: %d lit over the location, %d alone", menu, over, alone)
		}
	}
}
