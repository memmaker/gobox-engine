// Package walls is the shared pool of blobber wall, floor and ceiling
// textures, keyed by asset series: games of one series (e.g. "pool",
// "buck-rogers") share series/<series>/; a game names its series as data.
package walls

import (
	"bytes"
	"embed"
	"image"
	"image/png"
	"sort"
	"strings"
)

//go:embed all:series
var pool embed.FS

// Series lists the asset series that have textures, sorted.
func Series() []string {
	es, _ := pool.ReadDir("series")
	var out []string
	for _, e := range es {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out
}

// Image decodes texture name (no extension) of series; nil if absent.
func Image(series, name string) image.Image {
	b, err := pool.ReadFile("series/" + series + "/" + name + ".png")
	if err != nil {
		return nil
	}
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		return nil
	}
	return img
}

// Names lists every texture name of series, sorted.
func Names(series string) []string {
	es, _ := pool.ReadDir("series/" + series)
	var out []string
	for _, e := range es {
		if n, ok := strings.CutSuffix(e.Name(), ".png"); ok {
			out = append(out, n)
		}
	}
	sort.Strings(out)
	return out
}

// Textures maps each given name of series to its image, for blobber.View.Textures; absent names are skipped.
func Textures(series string, names ...string) map[string]image.Image {
	m := map[string]image.Image{}
	for _, n := range names {
		if img := Image(series, n); img != nil {
			m[n] = img
		}
	}
	return m
}
