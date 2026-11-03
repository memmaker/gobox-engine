// Package walls is the shared pool of blobber wall, floor and ceiling textures.
package walls

import (
	"bytes"
	"embed"
	"image"
	"image/png"
	"sort"
	"strings"
)

//go:embed pool/*.png
var pool embed.FS

// Image decodes texture name (no extension); nil if absent.
func Image(name string) image.Image {
	b, err := pool.ReadFile("pool/" + name + ".png")
	if err != nil {
		return nil
	}
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		return nil
	}
	return img
}

// Names lists every texture name, sorted.
func Names() []string {
	es, _ := pool.ReadDir("pool")
	var out []string
	for _, e := range es {
		out = append(out, strings.TrimSuffix(e.Name(), ".png"))
	}
	sort.Strings(out)
	return out
}

// Textures maps each given name to its image, for blobber.View.Textures; absent names are skipped.
func Textures(names ...string) map[string]image.Image {
	m := map[string]image.Image{}
	for _, n := range names {
		if img := Image(n); img != nil {
			m[n] = img
		}
	}
	return m
}
