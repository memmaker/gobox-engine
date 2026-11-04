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
)

//go:embed all:series
var fsys embed.FS

// Image decodes sprite name (no extension) of series from kind "portraits", "combat" or "items"; nil if absent.
func Image(series, kind, name string) image.Image {
	b, err := fsys.ReadFile("series/" + series + "/" + kind + "/" + name + ".png")
	if err != nil {
		return nil
	}
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		return nil
	}
	return img
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

// Origins is every series' origins.json merged (game ids are unique across
// series): game id, DAX file, "block" or "block.frame" -> its series' sprite name.
func Origins() map[string]map[string]map[string]string {
	o := map[string]map[string]map[string]string{}
	files, _ := fs.Glob(fsys, "series/*/origins.json")
	for _, f := range files {
		var m map[string]map[string]map[string]string
		if b, err := fsys.ReadFile(f); err == nil && json.Unmarshal(b, &m) == nil {
			for k, v := range m {
				o[k] = v
			}
		}
	}
	return o
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
