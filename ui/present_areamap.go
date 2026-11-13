package ui

import (
	"image"
	"image/draw"
	"slices"
	"strconv"

	"gobox/game"
	"gobox/gpu"
)

// areaPic is the area map's picture, built from the squares' pictures and
// kept while none of them changes. The pictures are resolved from the art at
// draw time, once a terrain.
type areaPic struct {
	img     *image.RGBA
	squares [][]image.Image // what img was built from
	tile    int

	terrain []int           // the terrain the pictures below are of
	size    image.Point     // ... and its size
	pics    [][]image.Image // terrain's square pictures
	pixels  int             // the largest square's side
	markers map[string]image.Image
}

// resolve fills pics and pixels for terrain t of size sz from the art (square pictures "terrain/<id>").
func (a *areaPic) resolve(t []int, sz image.Point, art map[string]game.Picture) {
	if a.pics != nil && a.size == sz && slices.Equal(a.terrain, t) {
		return
	}
	a.terrain, a.size, a.pixels = slices.Clone(t), sz, 0
	a.pics = make([][]image.Image, sz.Y)
	for y := range a.pics {
		a.pics[y] = make([]image.Image, sz.X)
		for x := range a.pics[y] {
			pc := art["terrain/"+strconv.Itoa(t[y*sz.X+x])]
			a.pics[y][x] = wallPic(pc)
			a.pixels = max(a.pixels, pc.W)
		}
	}
}

// marker is the party's icon art[name], kept.
func (a *areaPic) marker(name string, art map[string]game.Picture) image.Image {
	if m, ok := a.markers[name]; ok {
		return m
	}
	if a.markers == nil {
		a.markers = map[string]image.Image{}
	}
	a.markers[name] = wallPic(art[name])
	return a.markers[name]
}

// build is the picture of the squares (a seen square's picture, nil: none drawn), tile pixels each.
func (a *areaPic) build(squares [][]image.Image, seen [][]bool, tile int) image.Image {
	if seen != nil { // unseen squares drawn as none
		c := make([][]image.Image, len(squares))
		for y, row := range squares {
			c[y] = slices.Clone(row)
			for x := range row {
				if y >= len(seen) || x >= len(seen[y]) || !seen[y][x] {
					c[y][x] = nil
				}
			}
		}
		squares = c
	}
	if a.img != nil && a.tile == tile && sameSquares(a.squares, squares) {
		return a.img
	}
	h, w, n := len(squares), 0, tile
	if h > 0 {
		w = len(squares[0])
	}
	a.img = image.NewRGBA(image.Rect(0, 0, w*n, h*n))
	a.squares = make([][]image.Image, h)
	for y, row := range squares {
		a.squares[y] = slices.Clone(row)
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
// it (no bars), the party's marker on its square. Nothing without the map's art.
func (p *Presenter) drawAreaMap(dst *gpu.Image) {
	if p.eng == nil {
		return
	}
	am, art, where := p.eng.r.derived().automap, p.eng.art(), p.where()
	if len(am.Terrain) == 0 || len(art) == 0 || len(am.Terrain) != am.Size.X*am.Size.Y {
		return
	}
	a := &p.areaPic
	a.resolve(am.Terrain, am.Size, art)
	mk := am.Marker
	if mk == "" {
		mk = "marker"
	}
	var seen [][]bool
	if len(am.Known) < len(am.Terrain) { // else every square is seen
		seen = make([][]bool, am.Size.Y)
		for y := range seen {
			seen[y] = make([]bool, am.Size.X)
		}
		for _, q := range am.Known {
			if q.In(image.Rectangle{Max: am.Size}) {
				seen[q.Y][q.X] = true
			}
		}
	}
	p.area.Draw(dst, a.build(a.pics, seen, a.pixels), a.pixels, image.Pt(where.X, where.Y), a.marker(mk+"/"+strconv.Itoa(where.Facing), art), scale(dst))
}
