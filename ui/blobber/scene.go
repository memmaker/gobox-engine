package blobber

import (
	"image"
	"image/color"
)

// Scene is what a map shows: its grid, faces, thin walls, water, ponds,
// props, floor and ceiling. A host keeps one per map and Shows it on entering.
type Scene struct {
	Rows                []string
	Faces               map[[3]int][]FacePart
	Edges, NoMap        map[[3]int]bool
	Water, Pond         map[[2]int]bool
	Sea                 bool
	FloorRGBA, CeilRGBA *color.RGBA // plain flat colours, where Floor or Ceiling is ""
	Floor, Ceiling      string      // painted assets/walls/floor_NAME.png / ceiling_NAME.png ("" = default)
	Props               []*Object   // billboards and wells, fixed in place
	Backdrop            image.Image // over the whole view behind the walls, opaque and stretched nearest-neighbour, in place of sky and flats (nil: none)
}

// NewProp is a prop in the painted texture tex at floor position at: a
// billboard size squares high, or with ring a hole in the floor (see
// Object); nil when there is no such texture.
func NewProp(tex string, at [2]float64, size, front, sink float64, ring *[3]float64) *Object {
	img := wallTexture(tex)
	if img == nil {
		return nil
	}
	return &Object{X: int(at[0]), Y: int(at[1]), Img: img, Size: size, At: &at, Front: front, Sink: sink, Ring: ring}
}

// Show makes s the scene drawn, with the host's own objects (named props)
// among its props. It does not move the party: SetMap does.
func (s *Scene) Show(objs ...*Object) {
	Faces, Edges, Water, NoMap, Pond, Sea = s.Faces, s.Edges, s.Water, s.NoMap, s.Pond, s.Sea
	worldChanged()
	FlatColors = [2]*color.RGBA{s.FloorRGBA, s.CeilRGBA}
	SetFlats(s.Floor, s.Ceiling)
	setBackdrop(s.Backdrop)
	Objects = append(append(Objects[:0], s.Props...), objs...)
}
