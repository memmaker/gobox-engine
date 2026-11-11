// Package sprites is the shared pool of portraits, combat sprites and item
// icons, keyed by asset series: series/<series>/<kind>/<name>.png.
package sprites

import (
	"bytes"
	"embed"
	"encoding/json"
	"image"
	"image/png"
	"io/fs"
	"sort"
	"strings"
	"sync"
)

//go:embed all:series
var fsys embed.FS

// Image decodes sprite name (no extension) of series from kind "portraits", "combat" or "items"; nil if absent.
// Decoded once and kept: callers ask every View, and keep GPU copies by image identity.
func Image(series, kind, name string) image.Image {
	path := "series/" + series + "/" + kind + "/" + name + ".png"
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
	f, _ := fs.Sub(fsys, "series/"+series+"/combat/meshes")
	return f
}

// Combat is the series' combat directory: props.json says how its tiles draw (see combat.Presenter.Props).
func Combat(series string) fs.FS {
	f, _ := fs.Sub(fsys, "series/"+series+"/combat")
	return f
}

// Names lists the sprite names of a series' kind, sorted.
func Names(series, kind string) []string {
	es, _ := fsys.ReadDir("series/" + series + "/" + kind)
	var out []string
	for _, e := range es {
		if n, ok := strings.CutSuffix(e.Name(), ".png"); ok {
			out = append(out, n)
		}
	}
	sort.Strings(out)
	return out
}

// Pose is the series' combat head or body name drawn in the action pose when name is the ready one ("": none).
func Pose(series, name string) string {
	b, err := fsys.ReadFile("series/" + series + "/combat/poses.json")
	if err != nil {
		return ""
	}
	var m map[string]string
	json.Unmarshal(b, &m)
	return m[name]
}

// ItemIcons is a series' item icon sheet (itemicons/atlas.png) and its
// records (itemicons/items.rec: item name -> cell), one for every game of the series.
func ItemIcons(series string) (atlas []byte, rec string) {
	atlas, _ = fsys.ReadFile("series/" + series + "/itemicons/atlas.png")
	b, _ := fsys.ReadFile("series/" + series + "/itemicons/items.rec")
	return atlas, string(b)
}
