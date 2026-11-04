// Package sprites is the shared pool of portraits, combat sprites and item
// icons, keyed by asset series: series/<series>/<kind>/<name>.png.
package sprites

import (
	"bytes"
	"embed"
	"image"
	"image/png"
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
