package parity

import (
	"fmt"
	"path/filepath"
	"testing"

	"gobox/asset"
	"gobox/asset/orig"
	"gobox/ui"
	"gobox/ui/blobber"
)

// Every 3D area of the game (a GEO block) has its remastered map: the
// classic page never stands in for one.
func TestEveryAreaHasMap(t *testing.T) {
	l, err := orig.NewLoader(filepath.Join(porSources, "dos"))
	if err != nil {
		t.Skip("no game files")
	}
	if _, err := blobber.NewView(asset.Textures("pool", asset.WallNames("pool")...), &ui.Settings.Options); err != nil { // ParseScene checks the names against the current View.Textures
		t.Skip("no GPU:", err)
	}
	art, areaMaps := asset.Remaster("por")
	for disk := 1; disk <= 8; disk++ {
		f := l.File(fmt.Sprintf("GEO%d", disk))
		if f == nil {
			t.Skip("no GEO files")
		}
		for b := 0; b < 64; b++ {
			if f.BlockShared(b) == nil {
				continue
			}
			n, ok := areaMaps[b]
			if !ok {
				t.Errorf("GEO%d block %d: no remastered map (phlan-poc/areas.py, make_map.py)", disk, b)
				continue
			}
			m, err := art.Open("maps/" + n + ".map")
			if err != nil {
				t.Errorf("GEO%d block %d: %v", disk, b, err)
				continue
			}
			if _, _, err := blobber.ParseScene(m); err != nil {
				t.Errorf("%s.map: %v", n, err)
			}
			m.Close()
		}
	}
}
