package present

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"testing"
	"testing/fstest"

	"gobox/blobber"
	"gobox/front"
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
	w := &front.Walls{Key: "styletest", Side: [][][4]int{{{1, 1, 1, 1}}}, Pool: map[int]string{1: "w"}}
	dst := gpu.NewImage(blobber.W, blobber.H)
	frame := func() (red, blue int) { // the wall's pixels of either colour
		sc := p.wallScene(w)
		if p.shown != "s" {
			p.view.Show(sc, blobber.Camera{})
			p.shown = "s"
		}
		p.view.Draw(dst)
		for y := 0; y < blobber.H; y += 10 {
			for x := 0; x < blobber.W; x += 10 {
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
	p.keyAct(KeyF11, front.View{})
	if r, b := frame(); b == 0 || r != 0 || walls.Style() != "pixel" || !Host.PixelWalls {
		t.Fatalf("pixel: %d red, %d blue pixels (style %s), want the blue wall", r, b, walls.Style())
	}
	p.keyAct(KeyF11, front.View{})
	if r, b := frame(); r == 0 || b != 0 || walls.Style() != "hd" {
		t.Fatalf("back to hi-def: %d red, %d blue pixels, want the red wall", r, b)
	}
}
