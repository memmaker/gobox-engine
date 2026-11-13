package ui

import (
	"gobox/importer/pictures"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Every portrait the pool maps from a game's DAX is pixel-identical to the decoded DAX
// (games whose data/dos is not installed are skipped).
func TestPoolMatchesDAX(t *testing.T) {
	home, _ := os.UserHomeDir()
	defer func(id string) { gameID = id }(gameID)
	n := 0
	defer func() { t.Logf("compared %d", n) }()
	for id, files := range pictures.Origins() {
		dir := filepath.Join(home, "Projects", id+"-go", "data", "dos")
		if !useArt(t, dir) {
			t.Logf("skip %s: no data", id)
			continue
		}
		gameID = id
		for file, refs := range files {
			for ref, name := range refs {
				if !strings.HasPrefix(name, "head_") && !strings.HasPrefix(name, "body_") {
					continue
				}
				b, f, _ := strings.Cut(ref, ".")
				bi, _ := strconv.Atoi(b)
				fi, _ := strconv.Atoi(f)
				clear(daxPics)
				want := daxImage(file, bi, fi)
				if want == nil {
					continue
				}
				n++
				got := poolImage(frontPic(file, bi, fi))
				if got == nil || got.Bounds() != want.Bounds() {
					t.Errorf("%s %s %s: pool %v dax %v", id, file, ref, got != nil, want.Bounds())
					continue
				}
				bd := want.Bounds()
				for y := bd.Min.Y; y < bd.Max.Y; y++ {
					for x := bd.Min.X; x < bd.Max.X; x++ {
						gr, gg, gb, ga := got.At(x, y).RGBA()
						wr, wg, wb, wa := want.At(x, y).RGBA()
						if gr != wr || gg != wg || gb != wb || ga != wa {
							t.Fatalf("%s %s %s (%s): pixel %d,%d differs", id, file, ref, name, x, y)
						}
					}
				}
			}
		}
	}
}

func frontPic(file string, b, f int) Picture {
	return Picture{File: file, Block: b, Frame: f}
}
