package engine

import (
	"io/fs"
	"slices"
	"testing"

	"gobox/adventure"
	"gobox/asset"
)

// TestBR1CombatArtNames: a BR1 fight names its tiles as the series' combat art does (<sheet>_tile_<frame>), so the
// remastered meshes and the plain floor are found, and names a floor the series has a sprite for. Detects the class: an
// engine-made tile name the series art does not know (every mesh then silently falls back to the flat DOS tile).
func TestBR1CombatArtNames(t *testing.T) {
	_, f := br1AtTurn(t, br1FarFight, br1FarFacing)
	meshes, used := asset.Meshes("buck-rogers"), map[string]bool{}
	for _, g := range f.Field.Squares {
		if g.Tile < len(f.Field.Tiles) {
			used[f.Field.Tiles[g.Tile]] = true
		}
	}
	models := 0
	for n := range used {
		if _, err := fs.Stat(meshes, n+".gltf"); err == nil {
			models++
		}
	}
	t.Logf("floor %q, %d tiles in use, %d with a model", f.Field.FloorTile, len(used), models)
	if f.Field.FloorTile == "" || asset.SpriteImage("buck-rogers", "combat", f.Field.FloorTile) == nil {
		t.Errorf("floor %q has no series sprite", f.Field.FloorTile)
	}
	if models == 0 {
		t.Error("no tile of the fight has its model")
	}
}

// TestBR1LotekOverrun: lotek's tile 52, past its 52-frame block, is hitek's frame 52 (the memory the original reads past
// the block) in pixels and in name, so its sprite and model are found.
func TestBR1LotekOverrun(t *testing.T) {
	e, _ := br1AtTurn(t, br1FarFight, br1FarFacing)
	i := slices.IndexFunc(e.artNames().PlaceTiles, func(l adventure.TileLoad) bool { return l.File == "LOTEK" })
	if i < 0 {
		t.Fatal("no LOTEK place sheet")
	}
	l := e.artNames().PlaceTiles[i]
	e.sqrpaci[0] = blankTiles()
	e.LayTiles(l)
	const slot = 52
	if !slices.Equal(e.sqrpaci[0].Frame(l.First+slot), e.LoadPic("HITEK", l.Block).Frame(slot)) {
		t.Error("lotek tile 52 is not hitek's frame 52")
	}
	if n := sheetTileName(l, l.First+slot); n != "hitek_tile_052" {
		t.Errorf("lotek tile 52 named %q, want hitek_tile_052", n)
	}
}
