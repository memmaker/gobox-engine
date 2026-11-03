// Package sprites is the shared pool of portraits, combat sprites and item icons.
package sprites

import (
	"bytes"
	"embed"
	"image"
	"image/png"
	"sort"
	"strings"
)

//go:embed all:portraits all:combat all:items
var fsys embed.FS

// Image decodes sprite name (no extension) from pool "portraits", "combat" or "items"; nil if absent.
func Image(pool, name string) image.Image {
	b, err := fsys.ReadFile(pool + "/" + name + ".png")
	if err != nil {
		return nil
	}
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		return nil
	}
	return img
}

// Names lists the sprite names of a pool, sorted.
func Names(pool string) []string {
	es, _ := fsys.ReadDir(pool)
	var out []string
	for _, e := range es {
		if n, ok := strings.CutSuffix(e.Name(), ".png"); ok {
			out = append(out, n)
		}
	}
	sort.Strings(out)
	return out
}
