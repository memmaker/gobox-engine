// data.go: the game's data the screens lay themselves out by — its page
// boxes, the paper doll, the credits, the rule texts, its art — handed in
// by the game (SetData) before the first frame.

package present

import (
	"gobox/front/classic"
	"image"
	"image/color"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing/fstest"

	"gobox/amiga"
	"gobox/audio"
	"gobox/dax"
	"gobox/front"
	"gobox/game"
	"gobox/ui"
)

// Credit is one line of the title's credits page, at its cell X, Y in the
// colour Colour.
type Credit struct {
	X, Y, Colour int
	Text         string
	Page         int
}

// Data is what the game hands the presenter.
type Data struct {
	Series string // the asset series whose shared pools the game uses ("pool", "buck-rogers"; "": "pool")

	PaneBox, ViewWindow, InnerWindow front.CellRect // the original's main screen boxes at their 40x25 places

	MapCell      int // pixels a square of View.Map
	DungeonTiles int // the tiles at the start of the combat set

	Theme   *Theme            // the look (nil: DefaultTheme)
	Texts   map[string]string // the game's named texts and tables beyond Credits, SpellRules and IconParts (Text)
	Entries []Entry           // the game's own main menu entries, after the default list (nil: none)

	DoorRGBA  color.RGBA // a door's colour on the maps
	LabelInks []int      // a map label's palette index (View.Ink) by its kind
	SelectInk color.RGBA // the frame round the selected box (selectOutline)

	Rests []Rest // camp's rests, in the menu's order

	Logo LogoCut // how the main menu's lettering is cut from View.Logo

	Credits        []Credit
	RemasterCredit [2]string
	CreditGaps     []int // the credits' rows a blank row follows on a tall page, as many as fit from the first (nil: 4, 5, 6)
	IconParts      []string
	SpellRules     []string // what each spell does, by spell id (the camp's spell tips; "" none)

	ReadyMark string // shown before a readied item's name (InvItem.Readied), for a game without a doll ("": none)

	AreaMaps map[front.AreaID]string // the 3D map (Art maps/NAME.map) of each area
	Art      fs.FS                   // maps/NAME.map and assets/walls/walls.txt
	OpenAt   string                  // map:NAME:X:Y:D shows map NAME from that square alone (drawOpen); "" none

	DAX *dax.Loader // the game's own DAX loader, over its own files: it decodes the pictures named by File and Block with Img nil and Theme.Border (nil: they stay undrawn)

	Bindings []ui.Binding // the game's own help rows, after Any time's Ctrl-F row

	Sounds     []byte           // the sound driver's tables (gobox/audio NewSynth's data: songs, scripts, vibrato waveforms; nil: none)
	AdLib      *audio.AdLibData // the AdLib player's music and effects (nil: none); NewFront attaches it to the driver as the soundtrack
	AmigaSound []byte           // the Amiga's SOUND.BIN (amiga.Base..End; nil: the DOS driver), for the Driver NewFront builds
}

var (
	paneBox, viewWindow, innerWindow front.CellRect
	mapCell, dungeonTiles            int
	doorRGBA                         color.RGBA
	menuEntries                      []Entry
	texts                            map[string]string
	labelInks                        []int
	selectInk                        color.RGBA
	rests                            []Rest
	credits                          []Credit
	remasterCredit                   [2]string
	creditGaps                       []int
	creditFrame                      string // game.Data.CreditFrame
	titleBackdrop                    int    // game.Data.TitleBackdrop
	iconParts                        []string
	spellRules                       []string
	dollBoxes                        []image.Rectangle
	readiedMark                      string
	slotNames                        []string
	areaMaps                         map[front.AreaID]string
	art                              fs.FS = fstest.MapFS{} // Data.Art; empty until Load (headless: no walls.txt, no maps)
	openAt                           string
	dollPNG                          []byte
	daxLoader                        *dax.Loader
	mapSide                          int // View.Map's side in pixels
	sounds                           []byte
	amigaSound                       []byte
	adLib                            *audio.AdLibData
	series                           = "pool"
	logoCut                          LogoCut // Data.Logo
)

// SetData hands the presenter the game's data.
func SetData(d Data) {
	paneBox, viewWindow, innerWindow = d.PaneBox, d.ViewWindow, d.InnerWindow
	mapCell, dungeonTiles = d.MapCell, d.DungeonTiles
	if series = d.Series; series == "" {
		series = "pool"
	}
	menuEntries, texts, logoCut = d.Entries, d.Texts, d.Logo
	if d.Theme != nil {
		applyTheme(*d.Theme)
	} else {
		applyTheme(DefaultTheme)
	}
	doorRGBA, labelInks, selectInk = d.DoorRGBA, d.LabelInks, d.SelectInk
	rests = d.Rests
	credits, remasterCredit, iconParts, spellRules = d.Credits, d.RemasterCredit, d.IconParts, d.SpellRules
	if creditGaps = d.CreditGaps; creditGaps == nil {
		creditGaps = []int{4, 5, 6}
	}
	dollBoxes, slotNames, readiedMark = defaultDollBoxes, defaultSlotNames, d.ReadyMark
	areaMaps, art, openAt = d.AreaMaps, d.Art, d.OpenAt
	dollPNG = defaultDoll
	daxLoader = d.DAX
	if daxLoader != nil && !daxLoader.HasAmiga() { // the Amiga art beside the DOS files (data/amiga), shown on F12
		daxLoader.UseAmiga(filepath.Join(filepath.Dir(daxLoader.Dir), "amiga"))
		daxLoader.SetDOS(true)
	}
	classic.AmigaArt = daxLoader.Amiga // the members' pool portraits and icons follow F12 too
	if art == nil {
		art = fstest.MapFS{} // no remastered maps: every area falls back to its original view
	}
	mapSide = 16 * mapCell
	setBindings(d.Bindings)
	sounds, amigaSound, adLib = d.Sounds, d.AmigaSound, d.AdLib
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
	NoMapZones    bool    // a game may take the minimap and full map away in areas it marks (game.Automap.NoMap)
	PixelWalls    bool    // the 3D view: the pixel-art walls instead of the hi-def ones (Shift-F12)
}

// DefaultOptions are the settings before the game's are loaded.
var DefaultOptions = Options{MapAlpha: 1, Zoom: 0.5, Cards: 1, Map: 1, Pic: 1, DamageNumbers: true, NoMapZones: true}

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
	N     int    // ActCommand's: the index into game.Data.Commands
	Seq   string // a game.MenuItem's: the sequence it plays (playSeq), no Act sent
	Party bool
}

// amigaMX and amigaFx are Data.AmigaMusic's player and effect table (newDriver).
var (
	amigaMX *amiga.MX
	amigaFx []int
)

// Sounds is the game's sound tables (Data.Sounds), for audio.NewSynth.
func Sounds() []byte { return sounds }

// Text is the game's named text (Data.Texts; "": none).
func Text(name string) string { return texts[name] }

// sequences and filesDir are the game's Data.Sequences and Data.Files (View.Sequence).
var (
	sequences []game.Sequence
	filesDir  string
)

// gameArt is the game's Data.Art (the overland's terrain and marker).
var gameArt map[string]game.Picture

var daxMask map[string]int // Data.DaxMask

// gameData overlays the game's declarative game.Data (NewFront's) on what
// SetData handed in: its series and its spell tips (by spell id).
func gameData(g game.Data) {
	gameArt, gameID, daxMask = g.Art, g.ID, g.DaxMask
	sequences, filesDir = g.Sequences, g.Files
	classic.GameID, classic.Amiga = g.ID, func() bool { return daxLoader.Amiga() }
	if g.Series != "" {
		series = g.Series
	}
	classic.Series = series
	classic.Flow, classic.PointBuy = g.CreationFlow, g.PointBuy
	if g.Font != "" || g.Border != nil { // the game's look, as plain data
		t := theme
		if g.Font != "" {
			t.Font = g.Font
		}
		if b := g.Border; b != nil {
			t.Border = Border{File: b.File, Block: b.Block, Pieces: b.Pieces, Caps: b.Caps, Mask: b.Mask, Edges: b.Edges}
		}
		applyTheme(t)
	}
	if g.Files != "" && daxLoader == nil { // the game's DAX files, decoded here
		if l, err := dax.NewLoader(g.Files); err == nil {
			daxLoader = l
			daxLoader.UseAmiga(filepath.Join(filepath.Dir(l.Dir), "amiga"))
			daxLoader.SetDOS(true)
		}
	}
	if g.AdLib != nil && daxLoader != nil && adLib == nil {
		music, _ := os.ReadFile(daxLoader.Path(g.AdLib.Music))
		snd, _ := os.ReadFile(daxLoader.Path(g.AdLib.Sounds))
		if music != nil {
			adLib = &audio.AdLibData{Music: music, Sounds: snd, Effects: g.AdLib.Effects, Rate: g.AdLib.Rate, Tempo: g.AdLib.Tempo, Named: map[string][]byte{}}
			for _, sq := range g.Sequences { // the songs the animations cue (NAME.SMS beside them)
				for _, m := range sq.Music {
					if b, err := os.ReadFile(filepath.Join(g.Files, m+".SMS")); err == nil {
						adLib.Named[m] = b
					}
				}
			}
		}
	}
	if m := g.AmigaMusic; m != nil && daxLoader != nil && amigaMX == nil {
		dir := filepath.Join(filepath.Dir(daxLoader.Dir), "amiga")
		mx, _ := os.ReadFile(filepath.Join(dir, m.Songs))
		glb, _ := os.ReadFile(filepath.Join(dir, m.Effects))
		if p, err := amiga.NewMX(mx, glb); err == nil {
			p.Songs, amigaMX, amigaFx = m.SongOf, p, m.EffectOf
		}
	}
	if g.SoundTables != nil {
		sounds = g.SoundTables
	}
	if g.DungeonTiles != 0 {
		dungeonTiles = g.DungeonTiles
	}
	if g.Rests != nil {
		rests = nil
		for _, r := range g.Rests {
			rests = append(rests, Rest(r))
		}
	}
	if g.CreditLines != nil {
		credits = nil
		for _, c := range g.CreditLines {
			credits = append(credits, Credit(c))
		}
	}
	creditFrame, titleBackdrop = g.CreditFrame, g.TitleBackdrop
	if g.CreditGaps != nil {
		creditGaps = g.CreditGaps
	}
	if g.IconParts != nil {
		iconParts = g.IconParts
	}
	if g.LabelInks != nil {
		labelInks = g.LabelInks
	}
	if g.MapCell != 0 {
		mapCell = g.MapCell
	} else if mapCell == 0 {
		mapCell = 10 // the pool games' area map square
	}
	mapSide = 16 * mapCell
	if g.ReadyMark != "" {
		readiedMark = g.ReadyMark
	}
	if g.DoorInk != (color.RGBA{}) {
		doorRGBA = g.DoorInk
	}
	if g.SelectInk != (color.RGBA{}) {
		selectInk = g.SelectInk
	}
	if l := g.Logo; l != nil {
		logoCut = LogoCut(*l)
	}
	if g.RemasterCredit != ([2]string{}) {
		remasterCredit = g.RemasterCredit
	}
	if g.AreaMaps != nil {
		areaMaps = map[front.AreaID]string{}
		for id, n := range g.AreaMaps {
			areaMaps[front.AreaID(id)] = n
		}
	}
	if g.Remaster != nil {
		art = g.Remaster
	}
	if g.OpenAt != "" {
		openAt = g.OpenAt
	}
	if g.Bindings != nil {
		var own []ui.Binding
		for _, b := range g.Bindings {
			own = append(own, ui.Binding{Ctx: b.Ctx, Desc: b.Desc, Keys: b.Keys, Mouse: b.Mouse, Pad: b.Pad})
		}
		setBindings(own)
	}
	menuEntries = slices.DeleteFunc(menuEntries, func(e Entry) bool { return e.Act == front.ActCommand })
	menuEntries = slices.DeleteFunc(menuEntries, func(e Entry) bool { return e.Seq != "" })
	for _, m := range g.MenuItems {
		menuEntries = append(menuEntries, Entry{Text: m.Label, Key: byte(m.Key), Seq: m.Sequence})
	}
	for i, c := range g.Commands {
		menuEntries = append(menuEntries, Entry{Text: c.Label, Key: byte(c.Key), Act: front.ActCommand, N: i, Party: true})
	}
	if paneBox == (front.CellRect{}) { // the pool games' main screen: the message pane, the view's frame and its picture
		paneBox, viewWindow, innerWindow = front.CellRect{X1: 1, Y1: 17, X2: 38, Y2: 22}, front.CellRect{X1: 1, Y1: 1, X2: 15, Y2: 15}, front.CellRect{X1: 3, Y1: 3, X2: 13, Y2: 13}
	}
	if len(table.Rows) == len(bindings) { // no game rows: digits pick members
		setBindings([]ui.Binding{MemberDigits})
	}
	if g.SpellRules != nil {
		n := 0
		for id := range g.SpellRules {
			n = max(n, id+1)
		}
		spellRules = make([]string, n)
		for id, t := range g.SpellRules {
			if id >= 0 {
				spellRules[id] = t
			}
		}
	}
}

// UseGameData hands the presenter the game's data without building a Front (headless callers).
func UseGameData(g game.Data) { gameData(g) }
