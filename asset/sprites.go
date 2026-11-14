// sprites.go is the shared pool of portraits, combat sprites and item
// icons, keyed by asset series: sprites/series/<series>/<kind>/<name>.png.

package asset

import (
	"bytes"
	"embed"
	"image"
	"image/png"
	"io/fs"
	"sort"
	"strings"
	"sync"
)

//go:embed all:sprites/series
var fsys embed.FS

// SpriteImage decodes sprite name (no extension) of series from kind "portraits", "combat" or "items"; nil if absent.
// Decoded once and kept: callers ask every View, and keep GPU copies by image identity.
func SpriteImage(series, kind, name string) image.Image {
	path := "sprites/series/" + series + "/" + kind + "/" + name + ".png"
	if m, ok := decoded.Load(path); ok {
		im, _ := m.(image.Image)
		return im
	}
	m := decode(path)
	decoded.Store(path, m)
	return m
}

var decoded sync.Map // path -> image.Image (nil too: absent)

func decode(path string) image.Image {
	b, err := fsys.ReadFile(path)
	if err != nil {
		return nil
	}
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		return nil
	}
	return img
}

// Meshes is the series' combat tile models: <name>.gltf with its PNG (see combat.Presenter.Meshes).
func Meshes(series string) fs.FS {
	f, _ := fs.Sub(fsys, "sprites/series/"+series+"/combat/meshes")
	return f
}

// Combat is the series' combat directory: props.json says how its tiles draw (see combat.Presenter.Props).
func Combat(series string) fs.FS {
	f, _ := fs.Sub(fsys, "sprites/series/"+series+"/combat")
	return f
}

// names is Names' lists, made once each: the embedded files never change
// (the portrait pick lists them every frame).
var names sync.Map // series/kind -> []string

// SpriteNames lists the sprite names of a series' kind, sorted. The list is
// shared: callers must not change it.
func SpriteNames(series, kind string) []string {
	dir := "sprites/series/" + series + "/" + kind
	if out, ok := names.Load(dir); ok {
		return out.([]string)
	}
	es, _ := fsys.ReadDir(dir)
	var out []string
	for _, e := range es {
		if n, ok := strings.CutSuffix(e.Name(), ".png"); ok {
			out = append(out, n)
		}
	}
	sort.Strings(out)
	names.Store(dir, out)
	return out
}

// ItemIcons is a series' item icon sheet (itemicons/atlas.png) and its
// records (itemicons/items.rec: base item kind -> cell), one for every game of the series.
func ItemIcons(series string) (atlas []byte, rec string, err error) {
	if atlas, err = fsys.ReadFile("sprites/series/" + series + "/itemicons/atlas.png"); err != nil {
		return nil, "", err
	}
	b, err := fsys.ReadFile("sprites/series/" + series + "/itemicons/items.rec")
	return atlas, string(b), err
}
