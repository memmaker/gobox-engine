package pool

import (
	"strings"
	"testing"

	"gobox/dax"
)

// TestPicturesMasked (BUG-CLASSES 32b): por hands gobox its masked
// pictures (foes' figures, effects, the party marker) as pixels with
// their own transparent index (Picture.Clear, the loader's dax.Clear), so
// none leans on Data.DaxMask: the only picture named by file is the
// title's backdrop, a whole unmasked screen.
func TestPicturesMasked(t *testing.T) {
	e := partyTurnScene(t)
	foes := 0
	for _, u := range fighting(e).Field.Units {
		if u.Foe && u.Art.Pixels != nil {
			foes++
			if u.Art.Clear != int(dax.Clear) {
				t.Errorf("foe %s: Clear %d, want %d", u.Name, u.Art.Clear, dax.Clear)
			}
		}
	}
	if foes == 0 {
		t.Fatal("no foe figure")
	}
	for n, p := range e.overlandArt() {
		if p.File != "" {
			t.Errorf("%s: by file %s, with no DaxMask", n, p.File)
		}
		if strings.HasPrefix(n, "marker/") && p.Clear != int(dax.Clear) {
			t.Errorf("%s: Clear %d", n, p.Clear)
		}
	}
	if len(e.gameData.DaxMask) != 0 {
		t.Errorf("DaxMask %v: por names no masked file", e.gameData.DaxMask)
	}
}
