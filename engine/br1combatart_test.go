package engine

import (
	"io/fs"
	"testing"

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
