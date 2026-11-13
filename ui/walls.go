package ui

import (
	"image"
	"image/color"
	"strings"

	"gobox/gpu"
)

var wallCache = map[string]*gpu.Image{}

func decodeWall(name string) image.Image {
	if cur == nil {
		return nil
	}
	if img := cur.Textures[name]; img != nil || cur.Load == nil {
		return img
	}
	return cur.Load(name)
}

// wallTexture loads View.Textures[name], (once), or nil when there is none; a painted texture
// replaces the generated one for its material character.
func wallTexture(name string) *gpu.Image {
	if t, ok := wallCache[name]; ok {
		return t
	}
	img := decodeWall(name)
	if img == nil {
		wallCache[name] = nil
		return nil
	}
	t := gpu.NewImageFromImage(img)
	gpu.Note("texture %s %dx%d", name, img.Bounds().Dx(), img.Bounds().Dy())
	wallCache[name] = t
	return t
}

// blobberPaintedWalls is View.Painted: map character -> texture name.
func blobberPaintedWalls() map[byte]string {
	if cur != nil {
		return cur.Painted
	}
	return nil
}

var flatNames [2]string // the current map's floor and ceiling ("" = default)

// openSky: the ceiling is a sky_* texture, drawn as a backdrop behind everything instead of as
// ceiling tiles, panning with the view angle.
var openSky bool

// TryCeiling, when set by the host (a dev flag), replaces every map's ceiling.
var TryCeiling string

// FlatColors are the current map's plain floor and ceiling colours (nil: none), used by
// FlatTexture where no flat is named; Scene.Show sets them.
var FlatColors [2]*color.RGBA

// backdrop is the current scene's Scene.Backdrop on the GPU (nil: none);
// backdropSrc the image it was made from.
var (
	backdrop    *gpu.Image
	backdropSrc image.Image
)

func setBackdrop(m image.Image) {
	if m == backdropSrc {
		return
	}
	if backdrop != nil {
		backdrop.Deallocate()
	}
	backdrop, backdropSrc = nil, m
	if m != nil {
		backdrop = gpu.NewImageFromImage(m)
	}
}

var solidCache = map[solidKey]*gpu.Image{}

// SetFlats makes floor and ceiling (names from Flats; "" = the default textures) the current
// map's. Hosts call it on every map change; it holds across GenTextures.
func SetFlats(floor, ceiling string) {
	if TryCeiling != "" { // try a ceiling (e.g. cave) on any map
		ceiling = TryCeiling
	}
	flatNames = [2]string{floor, ceiling}
	openSky = strings.HasPrefix(ceiling, "sky")
	worldChanged()
	if defFloor != nil {
		texFloor, texCeil = FlatTexture("floor", floor), FlatTexture("ceiling", ceiling)
	}
}

// FlatTexture returns the floor or ceiling texture NAME, or the default one for "" or an unknown name.
func FlatTexture(kind, name string) *gpu.Image {
	if name != "" {
		if t := wallTexture(kind + "_" + name); t != nil {
			return t
		}
	}
	if c := FlatColors[map[string]int{"floor": 0, "ceiling": 1}[kind]]; c != nil && name == "" { // a game's colour on the pool's flat, never a flat fill
		k := solidKey{kind, *c}
		t, ok := solidCache[k]
		if src := decodeWall(kind); !ok && src != nil {
			t = gpu.NewImageFromImage(Tint(src, *c))
			solidCache[k] = t
		}
		if t != nil {
			return t
		}
	}
	if kind == "floor" {
		return defFloor
	}
	return defCeil
}

type solidKey struct {
	kind string
	c    color.RGBA
}

// Tint is texture tex's light and shade on colour c: a flat colour drawn
// textured (never a flat fill).
func Tint(tex image.Image, c color.RGBA) *image.RGBA {
	lum := func(c color.Color) float64 {
		r, g, b, _ := c.RGBA()
		return (0.3*float64(r)+0.59*float64(g)+0.11*float64(b))/65535 + 1e-3
	}
	b := tex.Bounds()
	var sum float64
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			sum += lum(tex.At(x, y))
		}
	}
	mean := sum / float64(b.Dx()*b.Dy())
	m := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			k := lum(tex.At(x, y)) / mean
			f := func(v uint8) uint8 { return uint8(min(255, float64(v)*k)) }
			m.SetRGBA(x-b.Min.X, y-b.Min.Y, color.RGBA{f(c.R), f(c.G), f(c.B), 255})
		}
	}
	return m
}
