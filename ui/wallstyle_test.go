package ui

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"testing"
	"testing/fstest"

	"gobox/game"
	"gobox/gpu"
	"gobox/walls"
)

// F11's style swap: the 3D view's next frame draws the pixel wall (and back),
// without a new map or a new view.
func TestWallStyleSwapsTextures(t *testing.T) {
	solid := func(c color.RGBA, w int) []byte {
		im := image.NewRGBA(image.Rect(0, 0, w, w))
		draw.Draw(im, im.Bounds(), image.NewUniform(c), image.Point{}, draw.Src)
		var b bytes.Buffer
		png.Encode(&b, im)
		return b.Bytes()
	}
	walls.Register("styletest", fstest.MapFS{
		"w.png":       {Data: solid(color.RGBA{255, 0, 0, 255}, 8)},
		"pixel/w.png": {Data: solid(color.RGBA{0, 0, 255, 255}, 4)},
	})
	old, oldHost := series, Host
	series, art = "styletest", fstest.MapFS{}
	h := DefaultOptions
	Host = &h
	defer func() { series, Host = old, oldHost; walls.SetStyle("hd") }()

	var p Presenter
	w := &game.Walls{Key: "styletest", Side: [][][4]int{{{1, 1, 1, 1}}}, Pool: map[int]string{1: "w"}}
	dst := gpu.NewImage(W, H)
	frame := func() (red, blue int) { // the wall's pixels of either colour
		sc := p.wallScene(w, nil)
		if p.shown != "s" {
			p.view.Show(sc, Camera{})
			p.shown = "s"
		}
		p.view.Draw(dst)
		for y := 0; y < H; y += 10 {
			for x := 0; x < W; x += 10 {
				r, _, b, _ := dst.At(x, y).RGBA()
				if r>>8 > 150 && b == 0 { // lit or in the torch's shade
					red++
				}
				if b>>8 > 150 && r == 0 {
					blue++
				}
			}
		}
		return
	}
	if r, b := frame(); r == 0 || b != 0 {
		t.Fatalf("hi-def: %d red, %d blue pixels, want the red wall", r, b)
	}
	p.keyAct(KeyShiftF12)
	if r, b := frame(); b == 0 || r != 0 || walls.Style() != "pixel" || !Host.PixelWalls {
		t.Fatalf("pixel: %d red, %d blue pixels (style %s), want the blue wall", r, b, walls.Style())
	}
	p.keyAct(KeyShiftF12)
	if r, b := frame(); r == 0 || b != 0 || walls.Style() != "hd" {
		t.Fatalf("back to hi-def: %d red, %d blue pixels, want the red wall", r, b)
	}
}

// A walls scene's very first frame (textures still held in memory, no device
// yet) already draws its floor and ceiling over the fog.
func TestWallSceneFirstFrameFlats(t *testing.T) {
	old, oldHost := series, Host
	series, art = "buck-rogers", fstest.MapFS{}
	h := DefaultOptions
	Host = &h
	defer func() { series, Host = old, oldHost }()

	var p Presenter
	w := &game.Walls{Key: "firstframe", Side: openGrid(), Floor: "br1_venus", Sky: "br1_greendome"}
	p.view.Show(p.wallScene(w, nil), Camera{X: 2, Y: 4})
	dst := gpu.NewImage(W, H)
	p.view.Draw(dst)
	for _, y := range []int{H / 8, H * 7 / 8} { // ceiling, floor
		if r, g, _, _ := dst.At(W/2, y).RGBA(); r|g == 0 { // the fog is blue only
			t.Errorf("first frame: row %d is the fog, not a flat", y)
		}
	}
}

func openGrid() [][][4]int {
	g := make([][][4]int, 5)
	for y := range g {
		g[y] = make([][4]int, 5)
	}
	return g
}
