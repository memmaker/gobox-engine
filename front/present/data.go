// data.go: the game's data the screens lay themselves out by — its page
// boxes, the paper doll, the credits, the rule texts, its art — handed in
// by the game (SetData) before the first frame.

package present

import (
	"image"
	"image/color"
	"io/fs"

	"gobox/dax"
	"gobox/front"
	"gobox/ui"
)

// Credit is one line of the title's credits page, at its cell X, Y in the
// colour Colour.
type Credit struct {
	X, Y, Colour int
	Text         string
}

// Data is what the game hands the presenter.
type Data struct {
	PaneBox, ViewWindow, InnerWindow front.CellRect // the original's main screen boxes at their 40x25 places

	MapCell      int // pixels a square of View.Map
	DungeonTiles int // the tiles at the start of the combat set

	Theme   *Theme            // the look (nil: DefaultTheme)
	Texts   map[string]string // the game's named texts and tables beyond Credits, SpellRules and IconParts (Text)
	Entries []Entry           // the game's own main menu entries, after the default list (nil: none)

	DoorRGBA      color.RGBA                                                    // a door's colour on the maps
	LabelInks     []int                                                         // a map label's palette index (View.Ink) by its kind
	SelectOutline func(image.Rectangle, int, func(image.Rectangle, color.RGBA)) // draws the outline round the selected box

	Rests []Rest // camp's rests, in the menu's order

	Credits        []Credit
	RemasterCredit [2]string
	IconParts      []string
	SpellRules     func(id, w int) []string

	DollBoxes []image.Rectangle // the doll's slot boxes in doll pixels, by doll slot
	SlotNames []string          // their names

	AreaMaps map[front.AreaID]string // the 3D map (Art maps/NAME.map) of each area
	Art      fs.FS                   // maps/NAME.map and assets/walls/walls.txt
	OpenAt   string                  // map:NAME:X:Y:D shows map NAME from that square alone (drawOpen); "" none

	BarsPNG, DollPNG, AtlasPNG []byte // the cards' bars, the paper doll, the item icons
	ItemsRec                   string // the icons' records

	DAX *dax.Loader // the game's own DAX loader, over its own files: it decodes the pictures named by File and Block with Img nil and Theme.Border (nil: they stay undrawn)

	Bindings []ui.Binding // the game's own help rows, after Any time's Ctrl-F row

	Sounds []byte // the sound driver's tables (gobox/audio NewSynth's data: songs, scripts, vibrato waveforms; nil: none)
}

var (
	paneBox, viewWindow, innerWindow front.CellRect
	mapCell, dungeonTiles            int
	doorRGBA                         color.RGBA
	menuEntries                      []Entry
	texts                            map[string]string
	labelInks                        []int
	selectOutline                    func(image.Rectangle, int, func(image.Rectangle, color.RGBA))
	rests                            []Rest
	credits                          []Credit
	remasterCredit                   [2]string
	iconParts                        []string
	spellRules                       func(id, w int) []string
	dollBoxes                        []image.Rectangle
	slotNames                        []string
	areaMaps                         map[front.AreaID]string
	art                              fs.FS
	openAt                           string
	barsPNG, dollPNG, atlasPNG       []byte
	itemsRec                         string
	daxLoader                        *dax.Loader
	mapSide                          int // View.Map's side in pixels
	sounds                           []byte
)

// SetData hands the presenter the game's data.
func SetData(d Data) {
	paneBox, viewWindow, innerWindow = d.PaneBox, d.ViewWindow, d.InnerWindow
	mapCell, dungeonTiles = d.MapCell, d.DungeonTiles
	menuEntries, texts = d.Entries, d.Texts
	if d.Theme != nil {
		applyTheme(*d.Theme)
	} else {
		applyTheme(DefaultTheme)
	}
	doorRGBA, labelInks, selectOutline = d.DoorRGBA, d.LabelInks, d.SelectOutline
	rests = d.Rests
	credits, remasterCredit, iconParts, spellRules = d.Credits, d.RemasterCredit, d.IconParts, d.SpellRules
	dollBoxes, slotNames = d.DollBoxes, d.SlotNames
	areaMaps, art, openAt = d.AreaMaps, d.Art, d.OpenAt
	barsPNG, dollPNG, atlasPNG, itemsRec = d.BarsPNG, d.DollPNG, d.AtlasPNG, d.ItemsRec
	daxLoader = d.DAX
	mapSide = 16 * mapCell
	setBindings(d.Bindings)
	sounds = d.Sounds
}

// Rest is one of camp's rests (Data.Rests): Text with its Key sends
// ActRest N. Fix: offered only while someone resting heals or studies
// (Camp.Healing, Studying). Night, NightKey: its text and key at night
// ("": Text's). Duration: the player sets how long first, in days and
// hours (ActRest N, X days, Y hours).
type Rest struct {
	Text     string
	Key      byte
	N        int
	Fix      bool
	Night    string
	NightKey byte
	Duration bool
}

// Options are the presenter's settings the game keeps (and saves).
type Options struct {
	MapOff        bool    // the minimap: hidden
	PartyRow      bool    // the portraits in one row along the bottom instead of down the sides
	MapAlpha      float64 // the minimap's opacity, 0.2 to 1
	Cards, Map    float64 // the portraits and minimap: one of UISteps times their normal size
	Pic           float64 // the picture box: one of UISteps times its normal size
	Zoom          float64 // the battle camera: how far one wheel step zooms, 0.25 to 2 times 15 percent
	DamageNumbers bool    // the battle: hit points lost float over the heads
}

// DefaultOptions are the settings before the game's are loaded.
var DefaultOptions = Options{MapAlpha: 1, Zoom: 0.5, Cards: 1, Map: 1, Pic: 1, DamageNumbers: true}

// Host is the settings in use: the game points it at its own.
var Host = &DefaultOptions

// UISteps are the sizes the screen elements come in: whole steps of the one glyph cell size (K).
var UISteps = []float64{1, 2, 3, 4}

// Entry is a main menu entry of the game's own (Data.Entries): Text with
// its Key sends Act (N -1) to the game, on the main list or, Party, in the
// Party submenu.
type Entry struct {
	Text  string
	Key   byte
	Act   front.ActKind
	Party bool
}

// Sounds is the game's sound tables (Data.Sounds), for audio.NewSynth.
func Sounds() []byte { return sounds }

// Text is the game's named text (Data.Texts; "": none).
func Text(name string) string { return texts[name] }
