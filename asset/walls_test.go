package asset

import (
	"bytes"
	"image"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"testing/fstest"
)

// The lookups over a registered series (the real series test their files in their own modules).
func TestLookups(t *testing.T) {
	var b bytes.Buffer
	png.Encode(&b, image.NewGray(image.Rect(0, 0, 1, 1)))
	Register("test", fstest.MapFS{"floor_water.png": {Data: b.Bytes()}})
	if WallNames("test")[0] != "floor_water" || WallImage("test", "nope") != nil || WallImage("none", "floor_water") != nil ||
		!Has("test", "floor_water") || Has("test", "nope") {
		t.Error("lookups wrong")
	}
}

// Every series under Dir reads from disk, each texture decoding under a unique name.
func TestSeriesOnDisk(t *testing.T) {
	if len(Series()) < 2 || len(WallNames("pool")) == 0 || WallImage("buck-rogers", "floor_water") != nil || Has("../walls", "walls") {
		t.Fatalf("series %v", Series())
	}
	for _, s := range Series() {
		for _, n := range WallNames(s) {
			if WallImage(s, n) == nil {
				t.Errorf("%s/%s does not decode", s, n)
			}
		}
	}
}

func TestStyle(t *testing.T) {
	png1 := func(w int) []byte {
		var b bytes.Buffer
		png.Encode(&b, image.NewRGBA(image.Rect(0, 0, w, w)))
		return b.Bytes()
	}
	Register("styled", fstest.MapFS{"a.png": {Data: png1(4)}, "b.png": {Data: png1(4)}, "pixel/a.png": {Data: png1(2)}})
	defer SetStyle("hd")
	size := func(n string) int { return WallImage("styled", n).Bounds().Dx() }
	if Style() != "hd" || size("a") != 4 {
		t.Fatal("hd: want the hi-def a")
	}
	SetStyle("pixel")
	if Style() != "pixel" || size("a") != 2 || size("b") != 4 || WallNames("styled")[0] != "a" || len(WallNames("styled")) != 2 {
		t.Fatal("pixel: want pixel/a, hi-def b (no pixel one), names unchanged")
	}
}

// openCount counts the files opened in it.
type openCount struct {
	fs.FS
	n *int
}

func (o openCount) Open(name string) (fs.File, error) { *o.n++; return o.FS.Open(name) }

// Image reads a texture once: the combat board asks for the floor every
// frame, and a per-frame read and decode was half its CPU (TestImageOnce
// guards that class: no file is opened for a texture already read).
func TestImageOnce(t *testing.T) {
	var b bytes.Buffer
	png.Encode(&b, image.NewGray(image.Rect(0, 0, 1, 1)))
	n := 0
	Register("once", openCount{fstest.MapFS{"floor.png": {Data: b.Bytes()}}, &n})
	first := WallImage("once", "floor")
	opened := n
	if first == nil || WallImage("once", "floor") != first || n != opened {
		t.Errorf("second Image opened %d files", n-opened)
	}
}

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

// Series lists the series under Dir, sorted.
func Series() []string {
	es, _ := os.ReadDir(filepath.Join(WallDir, "series"))
	var out []string
	for _, e := range es {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}

// WallNames lists every texture name of series, sorted.
func WallNames(s string) []string {
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
