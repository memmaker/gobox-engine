package present

import (
	"image"
	"image/draw"

	"gobox/front"
	"gobox/gpu"
)

// areaPic is the area map's picture, built from the squares' pictures and
// kept while none of them changes.
type areaPic struct {
	img     *image.RGBA
	squares [][]image.Image
	tile    int
}

func (a *areaPic) build(m *front.AreaMap) image.Image {
	if m.Seen != nil { // unseen squares drawn as none
		c := *m
		c.Squares = make([][]image.Image, len(m.Squares))
		for y, row := range m.Squares {
			c.Squares[y] = append([]image.Image(nil), row...)
			for x := range row {
				if y >= len(m.Seen) || x >= len(m.Seen[y]) || !m.Seen[y][x] {
					c.Squares[y][x] = nil
				}
			}
		}
		m = &c
	}
	if a.img != nil && a.tile == m.Tile && sameSquares(a.squares, m.Squares) {
		return a.img
	}
	h, w, n := len(m.Squares), 0, m.Tile
	if h > 0 {
		w = len(m.Squares[0])
	}
	a.img = image.NewRGBA(image.Rect(0, 0, w*n, h*n))
	a.squares = make([][]image.Image, h)
	for y, row := range m.Squares {
		a.squares[y] = append([]image.Image(nil), row...)
		for x, pic := range row {
			if pic != nil {
				draw.Draw(a.img, image.Rect(x*n, y*n, x*n+n, y*n+n), pic, pic.Bounds().Min, draw.Src)
			}
		}
	}
	a.tile = n
	return a.img
}

func sameSquares(a, b [][]image.Image) bool {
	if len(a) != len(b) {
		return false
	}
	for y := range a {
		if len(a[y]) != len(b[y]) {
			return false
		}
		for x := range a[y] {
			if a[y][x] != b[y][x] {
				return false
			}
		}
	}
	return true
}

// drawAreaMap draws the overland's whole map over the window: whole times
// its pixels, at least the page's scale and as much more as the map needs
// to cover the window, centred on the party as far as the map's edges let
// it (no bars), the party's marker on its square.
func (p *Presenter) drawAreaMap(dst *gpu.Image, m *front.AreaMap) {
	p.area.Draw(dst, p.areaPic.build(m), m.Tile, m.Party, m.Marker, scale(dst))
}
