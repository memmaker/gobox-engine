package engine

import (
	"slices"
	"testing"

	"gobox/rules"
)

// brIconCell is a combat icon's edge in pixels: one 24x24 cell (maptiles.pas DrawIconAnimSprite draws a 3 by 3 character cell).
const (
	brIconCell   = 24
	brMaskColour = 8 // maptiles.pas LoadIconAnim: LoadPic(.., True, 8, ..)
)

// TestBR1FightPictures: every unit of a BR fight publishes the picture the original shows (Unit.Art), from the imported
// art: a member's CHARS block (the record's icon) and a monster's CPIC block, 24x24, masked on colour 8 (the pixels of
// colour 8 are Clear), in the art's own palette; and the ground has the sheet LoadCombatTileSet lays for the place, each
// square's tile a picture of it.
func TestBR1FightPictures(t *testing.T) {
	e, f := br1AtTurn(t, br1FarFight, br1FarFacing)
	for _, u := range f.Field.Units {
		id := e.unitID(u.ID)
		slot := int(e.fact(id, rules.IconFact{}).(rules.Number))
		im := e.icons[slot][0]
		if im == nil {
			t.Errorf("unit %d %s: icon slot %d is empty", u.ID, u.Name, slot)
			continue
		}
		wantFile, wantBlock := "CPIC1", im.Block
		if !u.Foe {
			wantFile, wantBlock = "CHARS", e.fact(id, rules.IconPicsFact{}).(rules.IconPics).Body
			if slot >= 8 {
				t.Errorf("member %s holds icon slot %d, want one below 8 (AddCharToParty)", u.Name, slot)
			}
		}
		if im.File != wantFile || im.Block != wantBlock {
			t.Errorf("unit %d %s: icon is %s block %d, want %s block %d", u.ID, u.Name, im.File, im.Block, wantFile, wantBlock)
		}
		src := e.LoadPic(wantFile, wantBlock) // the imported art's own block, unmasked
		if src == nil {
			t.Errorf("unit %d %s: no %s block %d in the imported art", u.ID, u.Name, wantFile, wantBlock)
			continue
		}
		if u.Art.W != src.Width*8 || u.Art.H != src.Height || u.Art.W != brIconCell || u.Art.H != brIconCell {
			t.Errorf("unit %d %s: art %dx%d, the block is %dx%d (a cell is %d)", u.ID, u.Name, u.Art.W, u.Art.H, src.Width*8, src.Height, brIconCell)
		}
		if u.Art.Clear < 0 || len(u.Art.Palette) == 0 {
			t.Errorf("unit %d %s: art has no mask or palette (%d, %d colours)", u.ID, u.Name, u.Art.Clear, len(u.Art.Palette))
		}
		for i, p := range src.Frame(0) { // masked on 8: Clear where the block has colour 8, else its own index
			x := i % u.Art.W
			if u.Facing > 3 {
				x = u.Art.W - 1 - x
			}
			got := u.Art.Pixels[i-i%u.Art.W+x]
			if (p == brMaskColour) != (int(got) == u.Art.Clear) || p != brMaskColour && got != p && !(p > 16 && p < 32 && got == p&15) {
				t.Fatalf("unit %d %s: pixel %d is %d in the block, %d in the art (clear %d)", u.ID, u.Name, i, p, got, u.Art.Clear)
			}
		}
		if !u.Foe && u.Member < 0 {
			t.Errorf("unit %d %s: a member without its party index", u.ID, u.Name)
		}
	}
	// gdconsts.pas combatTerrain[t].tile: terrain 8 draws tile 1, 21 tile 14; a value of 0 or past 154 draws as 21 (DrawArenaCell)
	art := e.artNames()
	if art.TileOf(8) != 1 || art.TileOf(21) != 14 || art.TileOf(0) != 14 || art.TileOf(155) != 14 {
		t.Errorf("terrain tiles: 8 -> %d, 21 -> %d, 0 -> %d, 155 -> %d", art.TileOf(8), art.TileOf(21), art.TileOf(0), art.TileOf(155))
	}
	// the ground: the sheet of the place, laid from its tile; every square's tile is a picture of it
	traits, _ := e.placeTraits()
	var sheet = -1
	for i, l := range e.artNames().PlaceTiles {
		if l.Fits(traits) {
			sheet = i
			break
		}
	}
	if sheet < 0 {
		t.Fatalf("no tile sheet fits the place %b", traits)
	}
	l := e.artNames().PlaceTiles[sheet]
	src := e.LoadPic(l.File, l.Block)
	tiles := e.combatTiles()
	for n := range l.Last + 1 {
		if !slices.Equal(e.sqrpaci[0].Frame(l.First+n), src.Frame(n)) {
			t.Fatalf("tile %d is not %s block %d frame %d", l.First+n, l.File, l.Block, n)
		}
	}
	seen := map[int]bool{}
	for _, s := range f.Field.Squares {
		if seen[s.Tile] {
			continue
		}
		seen[s.Tile] = true
		if s.Tile < l.First || s.Tile > l.First+l.Last {
			t.Errorf("square %v: tile %d is outside the sheet %s (%d..%d)", s.At, s.Tile, l.File, l.First, l.First+l.Last)
			continue
		}
		p := tiles[s.Tile]
		colours := map[byte]bool{}
		for _, c := range p.Pix {
			colours[c] = true
		}
		if p.Pix == nil || len(p.Palette) == 0 || f.Field.Tiles[s.Tile] == "" || len(colours) < 2 {
			t.Errorf("square %v: tile %d has no picture (%d pixels, %d colours)", s.At, s.Tile, len(p.Pix), len(colours))
		}
	}
	if len(seen) < 2 {
		t.Errorf("the ground has %d kinds of tile", len(seen))
	}
}
