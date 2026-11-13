package engine

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestAmigaArtIndex: what each game's art index (art/key.txt) names of the
// Amiga import is in data/amiga at the size read, every Amiga sheet's text
// has its picture and the other way round, so an import that did not
// produce a file the index names fails here.
func TestAmigaArtIndex(t *testing.T) {
	ran := 0
	for _, game := range []string{"por", "cotab", "sotsb", "pod"} {
		data := filepath.Join("..", "data", game, "sources")
		key, err := os.ReadFile(filepath.Join(data, AssetsDir, ArtDir, "key.txt"))
		if err != nil {
			t.Logf("%s: no imported art, skipped", game)
			continue
		}
		ran++
		for _, line := range strings.Split(string(key), "\n") {
			f := strings.Fields(line)
			if len(f) != 3 || !strings.HasPrefix(f[0], AmigaDir+"/") {
				continue
			}
			fi, err := os.Stat(filepath.Join(data, f[0]))
			if err != nil {
				t.Errorf("%s: the art index names %s: %v", game, f[0], err)
			} else if strings.TrimSpace(f[1]) != itoa(fi.Size()) {
				t.Errorf("%s: %s is %d bytes, the art index read %s (stale art: run the game once)", game, f[0], fi.Size(), f[1])
			}
		}
		sheets := filepath.Join(data, AssetsDir, ArtDir, AmigaDir)
		ents, err := os.ReadDir(sheets)
		if err != nil {
			t.Errorf("%s: %v", game, err)
			continue
		}
		have := map[string]bool{}
		for _, e := range ents {
			have[e.Name()] = true
		}
		for n := range have {
			if base, ok := strings.CutSuffix(n, ".txt"); ok && !have[base+".png"] {
				t.Errorf("%s: sheet %s has no %s.png", game, n, base)
			}
			if base, ok := strings.CutSuffix(n, ".png"); ok && !have[base+".txt"] {
				t.Errorf("%s: picture %s has no %s.txt", game, n, base)
			}
		}
	}
	if ran == 0 {
		t.Skip("no game files")
	}
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }
