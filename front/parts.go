package front

import (
	"gobox/game"
	"gobox/rules"
	"image"
	"image/color"
	"strings"
	"time"
)

// The shapes the presenter's screens are laid out in and that games and
// tools outside it name: pictures, panes, maps, the creation sheet and
// the pickers. Nothing in them points back into the game's state.

// FrontAreaMap (extra) is the whole area map for a screen of its own: every
// square's picture (Squares[y][x], nil: none drawn), Tile pixels square,
// and the party's square and marker. The presenter builds the map's picture
// and keeps it while no square's picture changes.
type FrontAreaMap struct {
	Squares [][]image.Image
	Tile    int
	Party   image.Point // the party's map square
	Marker  image.Image // the party's icon, facing its way
	Seen    [][]bool    // Seen[y][x]: the party has seen that square (nil: all seen); unseen squares are drawn as none

	SquarePix [][]Picture // Squares as the game's pixels (Picture.Pix), the presenter's to colour into Squares
	MarkerPix Picture     // Marker so, or a DAX ref
}

// Read is journal text the game shows for the player to read:
// the presenter pages it (JournalPage) and answers ActOK, N the page it was
// left on; the script waits until then. It is the game's own (State.Read), the
// same pointer while the same text is up.
type Read = game.Read

// Money is core.Money: an amount the front end shows (as "12g 50c").
type Money = rules.Money

// Power is a member's action of the View menu the inventory offers too.
type Power struct {
	Name string
	Kind ActKind
}

// InvItem is one item of the inventory.
type InvItem struct {
	Item    rules.ItemHandle // names it (Action.Item)
	Name    string           // as the list shows it
	Type    byte             // base item type
	Words   []string         // the item's name words, known or not (for its icon)
	Count   int
	Readied bool     // it is readied (shown with present.Data.ReadyMark)
	Slot    bool     // it has a doll slot (it can be readied)
	Wrong   bool     // the member's class cannot ready it
	Use     bool     // it has a use (a scroll, a spell of its own)
	Look    []string // what it does (the shop's Look text)
	Price   Money    // what selling it pays (a ware: what it costs; loot: none)
	Scribe  bool     // a scroll someone was going to scribe from (the front end asks before it goes)
	Kind    ItemKind // its party inventory tab
	Rules   []string // extra: how it attacks in combat (the bar's tooltip), set on the Turn's Weapon and Missile
}

// ItemKind sorts the party inventory into tabs (extra).
type ItemKind int

const (
	KindMisc ItemKind = iota
	KindWeapon
	KindArmor
)

// Policy (extra) is a member's standing answer to one of the rule set's
// reactions, chosen beforehand on the sheet: nothing asks mid-turn.
type Policy uint8

const (
	Always Policy = iota // the zero value: an older save's, a foe's
	Never
	OnHit // only when it turns a hit (the rule set says which way)
)

func (p Policy) String() string { return [...]string{"Always", "Never", "On hit"}[p%3] }

// AreaID is the GEO block a 3D area is drawn from (the same area in every
// disk's GEO file that has it).
type AreaID int

// Pane is the text under the 3D view (TextPane): what DisplayTextBlock,
// the messages and the debug values put there, a line per row.
type Pane struct {
	Lines [][]PaneRun
	Rows  int // >0: the presenter shows Rows lines at a time and a key brings the next lot (0: all)
}

// PaneRun is text in one colour on a pane line from column X of the pane
// (0: its left edge); Upcase picks the glyphs as DisplayChar does.
type PaneRun struct {
	X      int
	Text   string
	Fg, Bg color.RGBA
	Upcase bool
	Blink  bool // the presenter alternates Fg with the highlight colour a few times (blink.go)
}

// FrontMapSquares is a map as the maps show it (minimap, the area map, the
// full map): squares [y][x], the party's square (off the map: not shown) and
// the step towards its facing.
type FrontMapSquares struct {
	Squares [16][16]FrontMapSquare
	Party   image.Point
	Ahead   image.Point
}

// FrontMapSquare is a square of a map: how far the party has seen it, its
// walls (N E S W) and its label (Letter 0: none).
type FrontMapSquare struct {
	Seen   FrontSeen
	Walls  [4]MapWall
	Kind   int  // the label's kind: its colour is Data.LabelInks[Kind]
	Letter byte // the label's legend letter
}

// FrontSeen is how much of a square the party has seen.
type FrontSeen uint8

const (
	FrontUnseen  FrontSeen = iota
	FrontSeenFar      // seen from afar
	FrontVisited      // walked on
)

// MapWall is a wall on a map.
type MapWall uint8

const (
	FrontNoWall MapWall = iota
	FrontWall
	FrontDoor
	FrontLockedDoor
)

// Shown reports whether the pane holds visible text (runs of blanks, as a
// cleared pane leaves, show nothing).
func (p *Pane) Shown() bool {
	for _, l := range p.Lines {
		for _, r := range l {
			if strings.TrimSpace(r.Text) != "" {
				return true
			}
		}
	}
	return false
}

// FrontNote is a toast for the status line: Text, put up at At
// (zero: none).
type FrontNote struct {
	Text string
	At   time.Time
}

// FrontTraining is the training hall's offer to the selected member.
type FrontTraining struct {
	Name    string
	Why     string      // the hall refuses: why
	Becomes []string    // "a level 3 Fighter" (Train trains)
	Price   int         // gold
	Gold    int         // the party's gold
	Learn   []BookSpell // a mage level was gained: the spell for the book (Learn), chosen before leaving
}

// BookSpell is a spell with a count, and the slots of its level.
type BookSpell struct {
	ID      int // the spell (ActPlan, ActScribe X)
	Name    string
	Level   string // its group, as "Cleric Level 1"
	Count   int
	Planned int // spells planned of its level
	Slots   int // slots of its level
}

// Pages is help's or the journal's page, in a box round the
// whole page: ActPage -1 and +1 page back and on (the journal: scroll a
// tall picture first), ActClose closes, any other key the next page.
type Pages struct {
	Title       PaneRun     // the box's title, on row 1
	Top         int         // the row of Lines[0]
	Lines       [][]PaneRun // the page's text, a line a row, X the page's column
	Scans       []Scan      // the journal's pictures on the page
	Hint        PaneRun     // on the message row: the page count or the keys
	Page, Pages int
}

// Scan is a journal picture (the file Name in journal/): the part Src of
// Img scaled into the page rectangle At, on paper of colour Paper.
type Scan struct {
	Name    string
	Img     image.Image
	Src, At image.Rectangle
	Paper   color.RGBA
}

// FrontFullMap is the full-screen map (FullMapView), the presenter's: its
// cursor cell's context menu has Walk (WalkTo), Mark and Remove.
type FrontFullMap struct {
	Name     string      // the map's name, top left
	Map      *FrontMapSquares // the squares, the legend's letter on each labelled square
	At       image.Point // the map's top left on the page, in pixels (At, Cell, LegendAt: the presenter lays them out)
	Cell     int         // pixels a square
	Cursor   image.Point // the cursor's square
	Legend   []MapLabel  // the legend's lines, from cell LegendAt down (the game gives them all, the presenter keeps those that fit)
	LegendAt image.Point
	Info     string  // the cursor's square and its label, on the message row
	Label    *string // the label being typed (nil: none), on the prompt row
	Labelled bool    // the cursor's square has a label (Mark is not offered)
	Mine     bool    // ... the player's (Remove is offered)
}

// MapLabel is a line of the full map's legend: the letter on its squares
// (in EGA colour Colour, by kind), its text and its first square.
type MapLabel struct {
	Letter byte
	Text   string
	Kind   int // its colour is Data.LabelInks[Kind]
	First  image.Point
}

// The pages the game lays out for a screen (pictures, the grid
// picker, the combat icon editor, the save slots).

// Picture is one thing drawn on the page, bottom first: frame Frame of
// block Block of the DAX file File (a picture, an animation's frame, a
// symbol, a tile) decoded in the page's palette with its mask pixels
// transparent, its top left at At.Min and cut to At; or, Fill set (the
// presenter's own layouts; the engine draws none), a fill of Colour over At.
type Picture struct {
	File         string // "" for a fill (the presenter adds its own pictures, the minimap, so)
	Block, Frame int
	Name         string      // a pool portrait name ("head_003"); with File, drawn from the pool when it resolves
	Img          image.Image // nil for a fill, or for a picture front/present decodes from File, Block, Frame
	Mask         *uint8      // with File: this colour index is transparent too, as the mask (nil: only dax.Clear)
	Pix          []uint8     // the game's own pixels, PixW wide, Img nil: front/present converts them in the page's ink (IndexedRGBA)
	PixW         int
	Palette      []color.RGBA // with Pix: Pix are indices into it (a VGA picture; alpha 0 transparent), not EGA; nil: EGA
	At           image.Rectangle
	Fill         bool
	Colour       color.RGBA
	Frames       []int // with FrameMs: the presenter cycles Frame through these (empty: static)
	FrameMs      int
}

// FrontPickHead .. PickForget are the pickers OpenPicker readies (kept in game,
// so a game names them without importing the front).
const (
	FrontPickHead     = game.PickHead
	FrontPickBody     = game.PickBody
	FrontPickIconHead = game.PickIconHead
	FrontPickWeapon   = game.PickWeapon
	PickAdd      = game.PickAdd
	PickForget   = game.PickForget
)

// FrontIconEditor is the combat icon editor's page, in cells:
// the presenter lays it out and takes its clicks. A click on Head or the
// poses' upper half opens ActPicker PickIconHead, on Weapon or their lower
// half PickWeapon, on Size ActIconSize; on Parts' row N, colour (X-X1)/3
// selects that swatch (ActIconColour N, X, -1); on the palette's cell the
// selected swatch (Part, Shade) takes its colour.
type FrontIconEditor struct {
	Poses              [2]image.Image // ready and action, drawn over PoseAt at the window's resolution
	PoseAt             [2]FrontCellRect
	Large              bool // Size: large (else small)
	Head, Weapon, Size FrontCellRect
	Names              FrontCellRect         // the parts' names, a row each
	Parts              FrontCellRect         // their two swatches, 2 cells wide and 3 apart
	Palette            FrontCellRect         // the colours, 8 a row, 2 cells each
	Colours            [16]int          // the colour number in each palette cell, row-major (-1: empty)
	Part, Shade        int              // the swatch the palette sets
	Colour             [6][2]int        // each part's two colour numbers (the swatches)
	Custom             bool             // the poses are a dropped PNG: no parts or colours, only Remove (ActIconRemove)
	Remove             FrontCellRect         // the Remove word, when Custom
	Swatch             [6][2]color.RGBA // each part's two colours as they show
	PaletteInk         [16]color.RGBA   // each palette cell's colour as it shows
	// Indexed alternative to Poses, Swatch and PaletteInk: with Indexed set the
	// presenter draws PosePix (as the game's pixels in the page's ink, or a DAX ref)
	// and takes SwatchIx and PaletteIx as indices of the page's ink.
	Indexed   bool
	PosePix   [2]Picture
	SwatchIx  [6][2]int
	PaletteIx [16]int
}

// FrontPicker is a grid picker's page as the engine drew it.
type FrontPicker struct {
	N, Page, Pages, PerPage, Cols int
	Cur                           int           // the picture it opened on
	Cells                         []FrontCellRect    // the pictures on this page: Page*PerPage on
	Pics                          []Picture     // their pictures, as the page would show them
	Prev, Next                    FrontCellRect      // the pager's arrows (zero: one page)
	Bottom                        int           // the box's inner bottom row (the box: from cell 1, 1 to the page's right)
	Labels                        []string      // the pictures' names (the saved characters)
	Classes                       []string      // ... their classes ("": the files are gone)
	Portraits                     []image.Image // ... their PNG portraits, fitted into the head's place (nil: none; the DAX head is among the pictures)
	Ask                           string        // the question before a pick (%s: the label), "" none
	Title                         string        // what is being picked, on the box's top edge ("" none)
}

// Slots is the save slot screen: the saves, newest first.
type Slots struct {
	Names, Times []string
	Overwrites   []string // the game's words for saving over each ("": "Overwrite NAME?")
	Levels       []int    // the highest level of a member of each slot's party (0 unknown)
	CanSave      bool     // Save is taken (else loading only, or not now)
	LoadOnly     bool     // the party menu's Load
	NoLoad       bool     // loading is not allowed here (State.Flags.NoLoad): no L=load
	Current      string   // the last save ("" none): the list starts on it
}

// MainMenu is the party menu's page: the functions open now, the title's
// pictures and the party.
type MainMenu struct {
	Open     []ActKind // ActCreate .. ActQuit, ActMember, ActDropMember
	Hall     bool      // the training hall's: over the 3D view, Begin leaves the hall
	Logo     *Picture  // the title picture (Pix) the lettering is cut from (Data.Logo); nil: none
	Party    int       // the members
	Selected int       // the selected member (-1: none)
	Name     string    // ... its name ("" none)
}

// FrontCellRect is a rectangle of 8x8 text cells, both corners included. For a
// box it is the inside, as DrawBorder takes it: the border is the ring of
// cells around it (Framed).
type FrontCellRect struct{ X1, Y1, X2, Y2 int }

// Pixels is the rectangle in screen pixels.
func (r FrontCellRect) Pixels() image.Rectangle {
	return image.Rect(r.X1*8, r.Y1*8, r.X2*8+8, r.Y2*8+8)
}

// Framed is the box with its border.
func (r FrontCellRect) Framed() FrontCellRect { return FrontCellRect{r.X1 - 1, r.Y1 - 1, r.X2 + 1, r.Y2 + 1} }

// FrontCreation is the creation sheet's parts: the drop-down fields, and the
// cells a click on the name, an ability, the portrait or the icon takes.
type FrontCreation struct {
	Fields    []Field // the sheet's fields in the game's order (ActSetField N): adnd1 FieldSex, FieldRace, FieldAlign, FieldClass
	Powers    []Field // the rule set's power picks (ActSetField N len(Fields)+i), off the page: the bar's Powers; Cur -1: none picked yet
	Name      FrontCellRect
	Stats     []FrontCellRect // the abilities' cells (ActStat N), STR .. CHA or the game's own
	StatNames []string   // their labels
	Head      FrontCellRect   // the portrait's head (PickHead) ...
	Body      FrontCellRect   // ... and body (PickBody)
	Icon      FrontCellRect   // the combat icon (ActIconEdit)
	Scale     int        // the portrait's and the icon's whole scale: the engine draws them at 1x from Head's and Icon's corners
	Abandon   string     // what Esc asks before ActClose
	Minimums  string     // the minimums Minimal last rolled to (ActRoll N 1)
}

// StatCol is the column an ability's value starts at after its label
// ("STR 18", "TECH 12").
func StatCol(label string) int { return max(5, len(label)+2) }

// The adnd1 creation's fields (ActSetField N).
const (
	FieldSex = iota
	FieldRace
	FieldAlign
	FieldClass
)

// Field is a drop-down of the creation sheet.
type Field struct {
	Name    string
	At      FrontCellRect // its text on the sheet
	Options []Option
	Cur     int // the option it holds
}

// Option is one of a Field's values; one that does not Fit changes the
// fields depending on it.
type Option struct {
	Text string
	Fits bool
}

// Waits: the options of what a suspended script waits on that a screen
// puts up (the treasure after a fight, a locked door).

// FrontLoot is the treasure screen's options besides View and Exit.
type FrontLoot struct {
	Take   bool     // there is treasure to take (LootTake)
	Detect bool     // the selected member has Detect Magic memorized (LootDetect)
	Share  bool     // there is money to divide among the party (LootShare)
	Pool   bool     // the party can pool its money (LootPool)
	Status bool     // the game offers Status (LootStatus)
	Drop   bool     // the party's items can be left in the treasure (ActDrop: the vault's Deposit)
	Money  []string // the treasure's money, "120 gold" (the loot inventory's border; ActTakeMoney)
}
