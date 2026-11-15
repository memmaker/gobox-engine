package ui

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"testing"
	"testing/fstest"

	"gobox/asset"
	"gobox/game"
	"gobox/gpu"
	"gobox/ui/blobber"
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
	asset.Register("styletest", fstest.MapFS{
		"w.png":       {Data: solid(color.RGBA{255, 0, 0, 255}, 8)},
		"pixel/w.png": {Data: solid(color.RGBA{0, 0, 255, 255}, 4)},
	})
	old, oldHost := series, Host
	series = "styletest"
	h := DefaultOptions
	Host = &h
	defer func() { series, Host = old, oldHost; asset.SetStyle("hd") }()

	var p Presenter
	w := &game.Walls{Key: "styletest", Side: [][][4]int{{{1, 1, 1, 1}}}, Pool: map[int]string{1: "w"}}
	dst := gpu.NewImage(blobber.W, blobber.H)
	frame := func() (red, blue int) { // the wall's pixels of either colour
		sc := p.wallScene(w, nil)
		if p.shown != "s" {
			p.view.Show(sc, blobber.Camera{})
			p.shown = "s"
		}
		p.view.Draw(dst)
		px := make([]byte, 4*blobber.W*blobber.H) // one read of the frame: At reads the texture back per pixel (93 s of the suite)
		dst.ReadPixels(px)
		for y := 0; y < blobber.H; y += 10 {
			for x := 0; x < blobber.W; x += 10 {
				r, b := px[4*(y*blobber.W+x)], px[4*(y*blobber.W+x)+2]
				if r > 150 && b == 0 { // lit or in the torch's shade
					red++
				}
				if b > 150 && r == 0 {
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
	if r, b := frame(); b == 0 || r != 0 || asset.Style() != "pixel" || !Host.PixelWalls {
		t.Fatalf("pixel: %d red, %d blue pixels (style %s), want the blue wall", r, b, asset.Style())
	}
	p.keyAct(KeyShiftF12)
	if r, b := frame(); r == 0 || b != 0 || asset.Style() != "hd" {
		t.Fatalf("back to hi-def: %d red, %d blue pixels, want the red wall", r, b)
	}
}

// A walls scene's very first frame (textures still held in memory, no device
// yet) already draws its floor and ceiling over the fog.
func TestWallSceneFirstFrameFlats(t *testing.T) {
	old, oldHost := series, Host
	series = "buck-rogers"
	h := DefaultOptions
	Host = &h
	defer func() { series, Host = old, oldHost }()

	var p Presenter
	w := &game.Walls{Key: "firstframe", Side: openGrid(), Floor: "br1_venus", Sky: "br1_greendome"}
	p.view.Show(p.wallScene(w, nil), blobber.Camera{X: 2, Y: 4})
	dst := gpu.NewImage(blobber.W, blobber.H)
	p.view.Draw(dst)
	for _, y := range []int{blobber.H / 8, blobber.H * 7 / 8} { // ceiling, floor
		if r, g, _, _ := dst.At(blobber.W/2, y).RGBA(); r|g == 0 { // the fog is blue only
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

// wallScene carries the area's scenery: water, ponds, the sea, and a
// free-standing wall's far side (Scenery.TwoSided) where that side is open.
func TestWallSceneScenery(t *testing.T) {
	var p Presenter
	w := &game.Walls{Key: "scenery", Side: [][][4]int{{{0, 2, 0, 0}, {0, 0, 0, 0}}}, Fronts: map[int]string{2: "w"},
		Scenery: game.Scenery{Water: [][2]int{{1, 0}}, Pond: [][2]int{{0, 0}}, Sea: true, TwoSided: []int{2}}}
	art := map[string]game.Picture{"w": {W: 1, H: 1, Pixels: []byte{1}, Clear: -1}}
	sc := p.wallScene(w, art)
	if len(sc.Faces[[3]int{0, 0, 1}]) == 0 || len(sc.Faces[[3]int{1, 0, 3}]) == 0 {
		t.Errorf("faces %v: want the wall's both sides", sc.Faces)
	}
	if !sc.Water[[2]int{1, 0}] || !sc.Pond[[2]int{0, 0}] || !sc.Sea {
		t.Errorf("water %v pond %v sea %v", sc.Water, sc.Pond, sc.Sea)
	}
}
