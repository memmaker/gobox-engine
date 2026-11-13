// Package walls is the shared pool of blobber wall, floor and ceiling
// textures, keyed by asset series: games of one series (e.g. "pool",
// "buck-rogers") share series/<series>/; a game names its series as data.
// The textures are read from disk when first asked for, not embedded: a
// game's binary carries none, and a map loads only the ones it draws.
package walls

import (
	"bytes"
	"image"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
)

// Dir is the directory holding series/<series>/<name>.png: $GOBOX_WALLS, else
// this package's own directory in the gobox checkout the game was built from.
var Dir = func() string {
	if d := os.Getenv("GOBOX_WALLS"); d != "" {
		return d
	}
	_, f, _, _ := runtime.Caller(0)
	return filepath.Dir(f)
}()

var (
	mu     sync.Mutex
	series = map[string]fs.FS{} // registered (tests) or opened from Dir
)

// Register sets series' texture files (in place of Dir's).
func Register(name string, files fs.FS) {
	mu.Lock()
	defer mu.Unlock()
	series[name] = files
	for k := range images {
		if k.series == name {
			delete(images, k)
		}
	}
}

func open(s string) fs.FS {
	mu.Lock()
	defer mu.Unlock()
	if f, ok := series[s]; ok {
		return f
	}
	if s == "" || strings.ContainsAny(s, `/\.`) {
		return nil
	}
	d := filepath.Join(Dir, "series", s)
	if st, err := os.Stat(d); err != nil || !st.IsDir() {
		return nil
	}
	series[s] = os.DirFS(d)
	return series[s]
}

// Series lists the series under Dir, sorted.
func Series() []string {
	es, _ := os.ReadDir(filepath.Join(Dir, "series"))
	var out []string
	for _, e := range es {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}

var pixel bool // the style: pixel/<name>.png over <name>.png where there is one

// SetStyle picks the textures' style: "pixel" (series/<series>/pixel/<name>.png,
// each falling back to the hi-def one where it is missing) or anything else hi-def.
func SetStyle(style string) {
	mu.Lock()
	defer mu.Unlock()
	pixel = style == "pixel"
}

// Style is the textures' style, "hd" or "pixel".
func Style() string {
	mu.Lock()
	defer mu.Unlock()
	if pixel {
		return "pixel"
	}
	return "hd"
}

// path is name's file in f in the current style.
func path(f fs.FS, name string) string {
	mu.Lock()
	px := pixel
	mu.Unlock()
	if px {
		if _, err := fs.Stat(f, "pixel/"+name+".png"); err == nil {
			return "pixel/" + name + ".png"
		}
	}
	return name + ".png"
}

func file(s, name string) ([]byte, error) {
	f := open(s)
	if f == nil {
		return nil, fs.ErrNotExist
	}
	return fs.ReadFile(f, path(f, name))
}

// Image decodes texture name (no extension) of series; nil if absent.
// Each file is read and decoded once (per style): callers draw it every
// frame and must not change it.
func Image(series, name string) image.Image {
	f := open(series)
	if f == nil {
		return nil
	}
	mu.Lock()
	k := imageKey{series, name, pixel}
	img, ok := images[k]
	mu.Unlock()
	if ok {
		return img
	}
	if b, err := fs.ReadFile(f, path(f, name)); err == nil {
		if m, err := png.Decode(bytes.NewReader(b)); err == nil {
			img = m
		}
	}
	mu.Lock()
	images[k] = img // nil too: an absent texture is looked for once
	mu.Unlock()
	return img
}

type imageKey struct {
	series, name string
	pixel        bool
}

var images = map[imageKey]image.Image{} // Image's decoded textures (guarded by mu)

// Has says whether series has texture name, without decoding it.
func Has(s, name string) bool {
	f := open(s)
	if f == nil {
		return false
	}
	_, err := fs.Stat(f, name+".png")
	return err == nil
}

// Names lists every texture name of series, sorted.
func Names(s string) []string {
	f := open(s)
	if f == nil {
		return nil
	}
	es, _ := fs.ReadDir(f, ".")
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
