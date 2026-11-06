package present

import (
	"image"
	"image/color"
	"testing"

	"gobox/blobber"
	"gobox/front"
	"gobox/gpu"
)

// Walls.Backdrop fills the whole 3D view with its picture, stretched
// nearest-neighbour; without it the flats' fills show.
func TestWallsBackdrop(t *testing.T) {
	quad := []color.RGBA{{255, 0, 0, 255}, {0, 255, 0, 255}, {0, 0, 255, 255}, {255, 255, 0, 255}}
	m := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for i, c := range quad {
		m.Set(i%2, i/2, c)
	}
	fill := color.RGBA{200, 0, 200, 255}
	open := [][][4]int{{{0, 0, 0, 0}}}
	dst := gpu.NewImage(blobber.W, blobber.H)
	draw := func(w *front.Walls) {
		var p Presenter
		p.view.Show(p.wallScene(w), blobber.Camera{})
		p.view.Draw(dst)
	}
	draw(&front.Walls{Key: "bd", Side: open, FloorRGBA: &fill, CeilRGBA: &fill, Backdrop: &front.Picture{Img: m}})
	for y := 0; y < blobber.H; y += 7 {
		for x := 0; x < blobber.W; x += 7 {
			want := quad[(y*2/blobber.H)*2+x*2/blobber.W]
			if got := color.RGBAModel.Convert(dst.At(x, y)); got != want {
				t.Fatalf("(%d,%d) %v, want %v", x, y, got, want)
			}
		}
	}
	draw(&front.Walls{Key: "nobd", Side: open, FloorRGBA: &fill, CeilRGBA: &fill})
	if r, g, b, _ := dst.At(blobber.W/2, 2).RGBA(); r>>8 < 100 || g != 0 || b>>8 < 100 { // the ceiling's fill, shaded
		t.Fatalf("nil backdrop: top %v, want the fill", dst.At(blobber.W/2, 2))
	}
}

// The remastered view drops Walls.Backdrop where the series paints either
// flat Walls names (the other takes the pool's), and keeps a real picture
// only where neither resolves.
func TestWallsBackdropFlats(t *testing.T) {
	pic := image.NewRGBA(image.Rect(0, 0, 2, 2))
	pic.Pix[0] = 200
	bd := &front.Picture{Img: pic}
	open := [][][4]int{{{0, 0, 0, 0}}}
	var p Presenter
	if sc := p.wallScene(&front.Walls{Key: "flats", Side: open, Floor: "cobble", Sky: "sky_phlan", Backdrop: bd}); sc.Backdrop != nil {
		t.Error("painted flats: backdrop kept")
	}
	if sc := p.wallScene(&front.Walls{Key: "oneflat", Side: open, Floor: "nosuch", Sky: "sky_phlan", Backdrop: bd}); sc.Backdrop != nil {
		t.Error("the sky resolves alone: backdrop kept")
	}
	if sc := p.wallScene(&front.Walls{Key: "noflat", Side: open, Floor: "nosuch", Sky: "nosuch", Backdrop: bd}); sc.Backdrop == nil {
		t.Error("no flat resolves: the picture dropped")
	}
}
