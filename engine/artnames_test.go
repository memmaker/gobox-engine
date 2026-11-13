package engine

import (
	"slices"
	"testing"

	"gobox/adventure"
	"gobox/asset"
)

// TestArtNamesImported (bug class: pool loaded POR's picture names in every game,
// so a cotab fight had no SQRPACI tiles and never showed): every file the game's
// ArtNames name, and every fixed sheet pool loads by name, is in its imported art.
// Detect the class: grep -n 'LoadPic("\|LoadIconPic("\|LayTiles(' pool/*.go names no file.
func TestArtNamesImported(t *testing.T) { artNamesImported(t, testEngine(t)) }

func TestArtNamesImportedCOTAB(t *testing.T) { artNamesImported(t, advEngine(t, advEnvCOTAB)) }
func TestArtNamesImportedSOTSB(t *testing.T) { artNamesImported(t, advEngine(t, advEnvSOTSB)) }
func TestArtNamesImportedPOD(t *testing.T)   { artNamesImported(t, advEngine(t, advEnvPOD)) }

func artNamesImported(t *testing.T, e *Engine) {
	t.Helper()
	if e.game().Art == nil {
		t.Fatal("GameTables.Art unset: the game's import writes its ArtNames")
	}
	a := e.artNames()
	for _, l := range slices.Concat(a.TileSheets, a.Tiles, a.DungeonTiles, a.WildTiles, a.AreaTiles) {
		if l.From != adventure.TileBlockFixed {
			continue // a block the game's state picks: Files checks the sheet
		}
		if e.artFolder().Block(asset.DOS, l.File, l.Block) == nil {
			t.Errorf("%s block %d: not imported", l.File, l.Block)
		}
	}
	// the sheets still loaded by a fixed name (set 1's of the numbered ones)
	for _, f := range append(a.Files(), "CBODY", "CHEAD", "CPIC1", "PIC1", "SPRIT1") {
		if _, ok := e.artFolder().FirstID(asset.DOS, f); !ok {
			t.Errorf("%s: not imported", f)
		}
	}
	if len(a.DungeonTiles) == 0 {
		t.Error("no combat tiles: a fight never shows (battleShown)")
	}
}
