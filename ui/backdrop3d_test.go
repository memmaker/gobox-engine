package ui

import (
	"image/color"
	"testing"

	"gobox/game"
	"gobox/gpu"
	"gobox/ui/blobber"
)

// Walls.Backdrop fills the whole 3D view with its picture, stretched
// nearest-neighbour; without it the flats' fills show.
func TestWallsBackdrop(t *testing.T) {
	quad := []color.RGBA{{255, 0, 0, 255}, {0, 255, 0, 255}, {0, 0, 255, 255}, {255, 255, 0, 255}}
	m := game.Picture{W: 2, Pixels: []byte{0, 1, 2, 3}, Palette: quad, Clear: -1}
	fill := color.RGBA{200, 0, 200, 255}
	open := [][][4]int{{{0, 0, 0, 0}}}
	dst := gpu.NewImage(blobber.W, blobber.H)
	draw := func(w *game.Walls) {
		var p Presenter
		p.view.Show(p.wallScene(w, nil), blobber.Camera{})
		p.view.Draw(dst)
	}
	draw(&game.Walls{Key: "bd", Side: open, FloorRGBA: &fill, CeilRGBA: &fill, Backdrop: &m})
	for y := 0; y < blobber.H; y += 7 {
		for x := 0; x < blobber.W; x += 7 {
			want := quad[(y*2/blobber.H)*2+x*2/blobber.W]
			if got := color.RGBAModel.Convert(dst.At(x, y)); got != want {
				t.Fatalf("(%d,%d) %v, want %v", x, y, got, want)
			}
		}
	}
	draw(&game.Walls{Key: "nobd", Side: open, FloorRGBA: &fill, CeilRGBA: &fill})
	if r, g, b, _ := dst.At(blobber.W/2, 2).RGBA(); r>>8 < 100 || g != 0 || b>>8 < 100 { // the ceiling's fill, shaded
		t.Fatalf("nil backdrop: top %v, want the fill", dst.At(blobber.W/2, 2))
	}
}

// The remastered view drops Walls.Backdrop where the series paints either
// flat Walls names (the other takes the pool's), and keeps a real picture
// only where neither resolves.
func TestWallsBackdropFlats(t *testing.T) {
	bd := &game.Picture{W: 2, Pixels: []byte{0, 1, 1, 1}, Palette: []color.RGBA{{200, 0, 0, 255}, {0, 0, 200, 255}}, Clear: -1}
	open := [][][4]int{{{0, 0, 0, 0}}}
	var p Presenter
	if sc := p.wallScene(&game.Walls{Key: "flats", Side: open, Floor: "cobble", Sky: "sky_phlan", Backdrop: bd}, nil); sc.Backdrop != nil {
		t.Error("painted flats: backdrop kept")
	}
	if sc := p.wallScene(&game.Walls{Key: "oneflat", Side: open, Floor: "nosuch", Sky: "sky_phlan", Backdrop: bd}, nil); sc.Backdrop != nil {
		t.Error("the sky resolves alone: backdrop kept")
	}
	if sc := p.wallScene(&game.Walls{Key: "noflat", Side: open, Floor: "nosuch", Sky: "nosuch", Backdrop: bd}, nil); sc.Backdrop == nil {
		t.Error("no flat resolves: the picture dropped")
	}
}
