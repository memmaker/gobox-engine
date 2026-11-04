package front

import (
	"image"
	"image/color"
)

// Views: the pages the game lays out for a screen (pictures, the grid
// picker, the combat icon editor, the save slots).

// Picture is one thing drawn on the page, bottom first: frame Frame of
// block Block of the DAX file File (a picture, an animation's frame, a
// symbol, a tile) decoded in the page's palette with its mask pixels
// transparent, its top left at At.Min and cut to At; or, Fill set (the
// presenter's own layouts; the engine draws none), a fill of Colour over At.
type Picture struct {
	File         string // "" for a fill (the presenter adds its own pictures, the minimap, so)
	Block, Frame int
	Img          image.Image // nil for a fill, or for a picture front/present decodes from File, Block, Frame
	Mask         *uint8      // with File: this colour index is transparent too, as the mask (nil: only dax.Clear)
	Pix          []uint8     // the game's own pixels, PixW wide, Img nil: front/present converts them in View.Ink (IndexedRGBA)
	PixW         int
	At           image.Rectangle
	Fill         bool
	Colour       color.RGBA
}

// PickHead .. PickForget are the pickers OpenPicker readies.
const (
	PickHead     = iota // the portrait's head (creation sheet)
	PickBody            // its body
	PickIconHead        // the combat icon's head (icon editor)
	PickWeapon          // its weapon
	PickAdd             // the saved characters to add to the party (party menu)
	PickForget          // the saved characters to forget (party menu)
)

// IconEditor is the combat icon editor's page (View.IconEditor), in cells:
// the presenter lays it out and takes its clicks. A click on Head or the
// poses' upper half opens ActPicker PickIconHead, on Weapon or their lower
// half PickWeapon, on Size ActIconSize; on Parts' row N, colour (X-X1)/3
// selects that swatch (ActIconColour N, X, -1); on the palette's cell the
// selected swatch (Part, Shade) takes its colour.
type IconEditor struct {
	Poses              [2]image.Image // ready and action, drawn over PoseAt at the window's resolution
	PoseAt             [2]CellRect
	Large              bool // Size: large (else small)
	Head, Weapon, Size CellRect
	Names              CellRect         // the parts' names, a row each
	Parts              CellRect         // their two swatches, 2 cells wide and 3 apart
	Palette            CellRect         // the colours, 8 a row, 2 cells each
	Colours            [16]int          // the colour number in each palette cell, row-major (-1: empty)
	Part, Shade        int              // the swatch the palette sets
	Colour             [6][2]int        // each part's two colour numbers (the swatches)
	Custom             bool             // the poses are a dropped PNG: no parts or colours, only Remove (ActIconRemove)
	Remove             CellRect         // the Remove word, when Custom
	Swatch             [6][2]color.RGBA // each part's two colours as they show
	PaletteInk         [16]color.RGBA   // each palette cell's colour as it shows
}

// Picker is a grid picker's page as the engine drew it.
type Picker struct {
	N, Page, Pages, PerPage, Cols int
	Cur                           int           // the picture it opened on
	Cells                         []CellRect    // the pictures on this page: Page*PerPage on
	Pics                          []Picture     // their pictures, as the page would show them
	Prev, Next                    CellRect      // the pager's arrows (zero: one page)
	Bottom                        int           // the box's inner bottom row (the box: from cell 1, 1 to the page's right)
	Labels                        []string      // the pictures' names (the saved characters)
	Classes                       []string      // ... their classes ("": the files are gone)
	Portraits                     []image.Image // ... their PNG portraits, fitted into the head's place (nil: none; the DAX head is among the pictures)
	Ask                           string        // the question before a pick (%s: the label), "" none
}

// Slots is the save slot screen: the saves, newest first.
type Slots struct {
	Names, Times []string
	Levels       []int  // the highest level of a member of each slot's party (0 unknown)
	CanSave      bool   // Save is taken (else loading only, or not now)
	LoadOnly     bool   // the party menu's Load
	Current      string // the last save ("" none): the list starts on it
}

// MainMenu is the party menu's page: the functions open now, the title's
// pictures and the party.
type MainMenu struct {
	Open     []ActKind   // ActCreate .. ActQuit, ActMember, ActDropMember
	Hall     bool        // the training hall's: over the 3D view, Begin leaves the hall
	Logo     *Picture    // the title picture (Pix) the lettering is cut from (Data.Logo); nil: none
	Backdrop image.Image // the painted title under it (nil: black)
	Party    int         // the members
	Selected int         // the selected member (-1: none)
	Name     string      // ... its name ("" none)
}
