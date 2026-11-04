package blobber

import (
	"image"
	"strings"

	"gobox/gpu"
)

var wallCache = map[string]*gpu.Image{}

func decodeWall(name string) image.Image {
	if cur != nil {
		return cur.Textures[name]
	}
	return nil
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
	wallCache[name] = t
	return t
}

// paintedWalls is View.Painted: map character -> texture name.
func paintedWalls() map[byte]string {
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
	if kind == "floor" {
		return defFloor
	}
	return defCeil
}
