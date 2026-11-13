package pool

import (
	"gobox/core"
	"gobox/core/goldbox"
	"image"
	"slices"
)

// The read-only seam for presenters other than the classic screen: a View
// is the game as the player sees it, published by the engine each time it
// waits for a key. Nothing in it points back into the engine's state.

// The insides of the original's main screen boxes, at their
// 40x25 places whatever the page's size: the remaster lays that screen
// out itself (the text pane as a pop-up, the party as cards).
var (
	OuterBox    = CellRect{X1: 1, Y1: 1, X2: paneRight, Y2: paneBottom}
	ViewWindow  = CellRect{X1: 1, Y1: 1, X2: 0xF, Y2: 0xF}
	InnerWindow = CellRect{X1: 3, Y1: 3, X2: 0xD, Y2: 0xD} // the picture frame
	TextPane    = CellRect{X1: 1, Y1: msgTop, X2: paneRight, Y2: paneBottom}
)

// TextPaneRows is the text pane's rows.
const TextPaneRows = paneBottom - msgTop + 1

// MapCell is the size of a square in Map, in pixels.
const MapCell = mapCell

const mapCell = 10

// snapState is the engine's side of the seam.
type snapState struct {
	quiet      bool                        // a play action is running: no scene is published for the events
	invOn      bool                        // ItemsMenu waits: Inventory is up
	shop       bool                        // ... as the shop's Browse: Inventory.Wares
	identified *core.ItemCard              // the item the shop's Id last revealed (Shop.Identified)
	identN     int                         // ... counted: a new reveal is a new number
	identIt    core.ItemHandle             // ... the item itself (State's Shop.Identified)
	saves      bool                        // there are saved games (the party menu's Load)
	logo       *Picture                    // titleLogo, once read
	win        Window                      // set by whatever drew into the view window last
	heads      map[int]Picture             // HEAD blocks' pixels by id
	tiles      [3]Picture                  // DrawBorder style 0's corner, side and top cells, once taken (BorderPix)
	cursor     image.Point                 // drawCursor's last square
	cursorOn   bool                        // the cursor icon is on the page (drawn, not yet erased or drawn over)
	cmapOn     bool                        // the combat map is set up (cmapShown) and shown
	sheetOn    bool                        // a character sheet is up over the combat map
	portrait   CellRect                    // the sheet's portrait box while a sheet is up
	sheet      *SheetPage                  // the character sheet's page while it is up (SheetPage)
	results    *Results                    // the after-combat page while it is up (Results)
	full       bool                        // a whole-screen picture is up (Full)
	iconDrop   func([]byte)                // the icon editor, or the creation sheet, takes a dropped PNG while it is up (nil elsewhere)
	iconBox    CellRect                    // the creation sheet's icon: the cells that take the drop (zero: the whole screen)
	dungeon    bool                        // the combat tiles are DungCom's stone walls (else WildCom's outdoors)
	marks      []image.Point               // squares markReach framed
	reach      []image.Point               // squares a click on the main combat menu walks to
	paths      map[image.Point][]core.Step // the walk to each of them
	hits       map[int]int                 // each of targets' chance to be hit, in percent
	area       []image.Point               // squares aoeMarks covers
	skipped    []image.Point               // squares of allies in the area the spell passes over (Bless: an enemy next to them)
	aimWhy     string                      // aiming: why the aimed square cannot be taken ("" when it can)
	turn       *Turn                       // the player's turn waits in act
	targets    []int                       // combatants verbAttack takes; aiming: the foes to cycle through
	strike     Strike                      // the last weapon attack (From/To filled in by battle)
	striker    [2]core.CharID              // its attacker and target
}

// charHead is the head the player chose for c, as the sheet shows it
// (SheetPortrait), as View pixels in the EGA's own colours (Member.HeadPix;
// none without one).
func (e *Engine) charHead(cID core.CharID) Picture {
	h := e.fact(cID, core.PortraitFact{}).(core.Portrait).Head
	if h == 0 || h >= len(e.game().PortraitHead) {
		return Picture{}
	}
	return e.head(e.game().PortraitHead[h])
}

// head is portrait head p's pixels (cached by id, one picture in every
// set; none if missing).
func (e *Engine) head(p goldbox.PortraitPic) Picture {
	id := int(p.Id)
	if DOSArt.Load() {
		id |= 1 << 16 // the DOS twin's
	}
	if pc, ok := e.snap.heads[id]; ok {
		return pc
	}
	if e.snap.heads == nil {
		e.snap.heads = map[int]Picture{}
	}
	var out Picture
	im := e.art(e.LoadPic(p.Set.File("HEAD"), int(p.Id)), 0)
	if im != nil && im.Count > 0 {
		out = Picture{Pix: slices.Clone(im.Frame(0)), PixW: im.Width * 8, Palette: im.Palette}
	}
	e.snap.heads[id] = out
	return out
}

// menuShown records the menu DisplayInput waits on, and returns the one
// borderTiles is DrawBorder's style 0 cells (0x114 corner, 0x115 side,
// 0x116 top), taken once the symbols are loaded, as the border's pieces by
// place (ui.TopLeft ..): View pixels in the EGA's own colours.
func (e *Engine) borderTiles() [8]Picture {
	if cells := e.artNames().BorderCells; e.snap.tiles[0].Pix == nil && e.sym[4] != nil && len(cells) == len(e.snap.tiles) {
		im := e.sym[4]
		for i, c := range cells {
			e.snap.tiles[i] = Picture{Pix: slices.Clone(im.Frame(c - symFix[4])), PixW: im.Width * 8, Palette: im.Palette}
		}
	}
	c, side, top := e.snap.tiles[0], e.snap.tiles[1], e.snap.tiles[2]
	return [8]Picture{c, c, c, c, top, top, side, side}
}
