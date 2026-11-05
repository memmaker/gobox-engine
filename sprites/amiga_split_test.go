package sprites

import (
	"strings"
	"testing"
)

// TestAmigaSetApart: the <game>-amiga origins name only amiga_ sprites and the DOS ones never do, so F12 never mixes them.
func TestAmigaSetApart(t *testing.T) {
	o := Origins()
	if len(o["cotab-amiga"]["HEAD.DAX"]) == 0 {
		t.Fatal("no cotab Amiga heads")
	}
	for g, files := range o {
		for f, refs := range files {
			for r, n := range refs {
				if strings.HasSuffix(g, "-amiga") != strings.HasPrefix(n, "amiga_") {
					t.Errorf("%s %s %s -> %s", g, f, r, n)
				}
			}
		}
	}
}
