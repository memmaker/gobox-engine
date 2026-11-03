package ui

import (
	"image"
	"math"

	"gobox/gpu"
)

// AreaMap draws the overland's whole map over the window and keeps the
// uploaded pictures while their sources stay the same.
type AreaMap struct {
	src, markSrc image.Image
	img, mark    *gpu.Image
}

// AreaGeom is where the map goes in a window of bounds b: the pixels k a
// map pixel takes (whole, at least k0 and as much more as the map of
// bounds mb needs to cover the window), the pixels n a tile takes, and the
// offset x, y of the window in the scaled map, centred on the party's
// square as far as the map's edges let it.
func AreaGeom(b, mb image.Rectangle, k0 float64, tile int, party image.Point) (k, n, x, y float64) {
	k = max(k0, math.Ceil(float64(b.Dx())/float64(mb.Dx())), math.Ceil(float64(b.Dy())/float64(mb.Dy())))
	n = k * float64(tile)
	cx, cy := (float64(party.X)+0.5)*n, (float64(party.Y)+0.5)*n
	x = min(max(cx-float64(b.Dx())/2, 0), k*float64(mb.Dx())-float64(b.Dx()))
	y = min(max(cy-float64(b.Dy())/2, 0), k*float64(mb.Dy())-float64(b.Dy()))
	return
}

// Draw draws pic (tiles of tile pixels) with marker (nil: none) on the
// party's square, at least k0 times its pixels.
func (a *AreaMap) Draw(dst *gpu.Image, pic image.Image, tile int, party image.Point, marker image.Image, k0 float64) {
	a.draw(dst, image.Point{}, pic, tile, party, marker, k0)
}

// DrawIn is Draw into the rectangle r of dst only (clipped to it, centred
// on the party as in a window of r's size).
func (a *AreaMap) DrawIn(dst *gpu.Image, r image.Rectangle, pic image.Image, tile int, party image.Point, marker image.Image, k0 float64) {
	a.draw(dst.SubImage(r).(*gpu.Image), r.Min, pic, tile, party, marker, k0)
}

// draw draws into dst with its map origin at off.
func (a *AreaMap) draw(dst *gpu.Image, off image.Point, pic image.Image, tile int, party image.Point, marker image.Image, k0 float64) {
	if a.src != pic {
		if a.img != nil {
			a.img.Deallocate()
		}
		a.src, a.img = pic, gpu.NewImageFromImage(pic)
	}
	mb := pic.Bounds()
	k, n, x, y := AreaGeom(dst.Bounds(), mb, k0, tile, party)
	op := &gpu.DrawImageOptions{}
	op.GeoM.Scale(k, k)
	op.GeoM.Translate(float64(off.X)-x, float64(off.Y)-y)
	dst.DrawImage(a.img, op)
	if marker == nil {
		return
	}
	if a.markSrc != marker {
		if a.mark != nil {
			a.mark.Deallocate()
		}
		a.markSrc, a.mark = marker, gpu.NewImageFromImage(marker)
	}
	mk := marker.Bounds()
	op = &gpu.DrawImageOptions{}
	op.GeoM.Scale(k, k)
	op.GeoM.Translate(float64(party.X)*n+(n-k*float64(mk.Dx()))/2-x+float64(off.X), float64(party.Y)*n+(n-k*float64(mk.Dy()))/2-y+float64(off.Y))
	dst.DrawImage(a.mark, op)
}
