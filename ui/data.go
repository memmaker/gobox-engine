// data.go: the game's data the screens lay themselves out by — its page
// boxes, the paper doll, the credits, the rule texts, its art — handed in
// by the game (SetData) before the first frame.

package ui

import (
	"cmp"
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing/fstest"

	"gobox/adventure"
	"gobox/asset"
	"gobox/asset/orig"
	"gobox/audio"
	"gobox/game"
)

// Data is what the game hands the presenter.
type Data struct {
	Series string // the asset series whose shared pools the game uses ("pool", "buck-rogers"; "": "pool")

	PaneBox, ViewWindow, InnerWindow CellRect // the original's main screen boxes at their 40x25 places

	DungeonTiles int // the tiles at the start of the combat set

	Theme   *Theme            // the look (nil: DefaultTheme)
	Texts   map[string]string // the game's named texts and tables beyond Credits, SpellRules and IconParts (Text)
	Entries []Entry           // the game's own main menu entries, after the default list (nil: none)

	Rests []Rest // camp's rests, in the menu's order

	IconParts  []string
	SpellRules []string // what each spell does, by spell id (the camp's spell tips; "" none)

	AreaMaps map[AreaID]string // the 3D map (Art maps/NAME.map) of each area
	Art      fs.FS             // maps/NAME.map and assets/walls/walls.txt
	OpenAt   string            // map:NAME:X:Y:D shows map NAME from that square alone (drawOpen); "" none

	Bindings []Binding // the game's own help rows, after Any time's Ctrl-F row

	Sounds     []byte           // the sound driver's tables (gobox/audio NewSynth's data: songs, scripts, vibrato waveforms; nil: none)
	AdLib      *audio.AdLibData // the AdLib player's music and effects (nil: none); NewFront attaches it to the driver as the soundtrack
	AmigaSound []byte           // the Amiga's SOUND.BIN (amiga.Base..End; nil: the DOS driver), for the Driver NewFront builds
}

// readLogo is the lettering PNG at path; nil when the game has none. A file
// that is not a PNG is a broken import: it panics.
func readLogo(path string) image.Image {
	b, err := asset.ReadPNG(path)
	if err != nil {
		panic(fmt.Sprintf("%v (run the import again)", err))
	}
	if b == nil {
		return nil
	}
	im, err := asset.DecodePNG(b)
	if err != nil {
		panic(fmt.Sprintf("%s: %v (run the import again)", path, err))
	}
	return im
}

var (
	paneBox, viewWindow, innerWindow CellRect
	dungeonTiles                     int
	menuEntries                      []Entry
	texts                            map[string]string
	rests                            []Rest
	creditGaps                       []int // the credits' rows a blank row follows on a tall page, as many as fit from the first (nil: none)
	iconParts                        []string
	spellRules                       []string
	dollBoxes                        = defaultDollBoxes // the doll is every game's, whichever way its data came (SetData or NewFront)
	slotNames                        = defaultSlotNames
	areaMaps                         map[AreaID]string
	art                              fs.FS = fstest.MapFS{} // Data.Art; empty until Load (headless: no walls.txt, no maps)
	openAt                           string
	dollPNG                          = defaultDoll
	sounds                           []byte
	amigaSound                       []byte
	adLib                            *audio.AdLibData
	series                           = "pool"
	menuLogo                         image.Image // the imported lettering (adventure.LogoFile); nil: none
)

// The maps' and lists' looks, gobox's for every game: an area map
// square's pixels (mapCell) and the map's side (mapSide, 16 squares), a
// door's colour on the maps (light brown, Amiga 0xA84: it stands out
// against the white walls and the grey, black and blue floors), a map
// label's EGA colour by its kind (labelKinds), and the mark before a
// readied item's name. The adventure's look file (adventure.Look) replaces
// the door, the mark and the Amiga combat ground.
const (
	mapCell = 10
	mapSide = 16 * mapCell

	defaultReadyMark = "*"
)

var (
	defaultDoor = color.RGBA{0xaa, 0x88, 0x44, 0xff}
	doorRGBA    = defaultDoor
	readiedMark = defaultReadyMark
	amigaGround *color.RGBA                    // the combat tiles' bare ground in the Amiga art; nil: black, as in the DOS art
	labelInks   = [...]int{15, 10, 13, 14, 11} // the player's, shop, temple, inn, civic
)

// applyLook sets the door, ready mark and Amiga ground from the adventure
// folder adv's look file; a broken one is a broken import: it panics.
func applyLook(adv string) {
	l, err := adventure.ReadLook(adv)
	if err != nil {
		panic(fmt.Sprintf("%v (run the import again)", err))
	}
	doorRGBA, readiedMark, amigaGround = defaultDoor, cmp.Or(l.ReadyMark, defaultReadyMark), l.AmigaGround
	if l.Door != nil {
		doorRGBA = *l.Door
	}
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
	SIDMusic      bool    // the AdLib songs on the Innovation SSI-2001's SID (audio.AdLib.SetSID), where they have its instruments
	PixelWalls    bool    // the 3D view: the pixel-art walls instead of the hi-def ones (Shift-F12)
}

// DefaultOptions are the settings before the game's are loaded.
var DefaultOptions = Options{MapAlpha: 1, Zoom: 0.5, Cards: 1, Map: 1, Pic: 1, DamageNumbers: true, NoMapZones: true}

// Host is the presenter's settings in use: the shared engine settings' (HostSettings.Options).
var Host = &HostSettings.Options

// UISteps are the sizes the screen elements come in: whole steps of the one glyph cell size (K).
var UISteps = []float64{1, 2, 3, 4}

// Entry is a main menu entry of the game's own (Data.Entries): Text with
// its Key sends Act (N -1) to the game, on the main list or, Party, in the
// Party submenu.
type Entry struct {
	Text  string
	Key   byte
	Act   ActKind
	N     int    // ActCommand's: the index into game.Data.Commands
	Seq   string // a game.MenuItem's: the sequence it plays (playSeq), no Act sent
	Party bool
}

// amigaMX and amigaFx are Data.AmigaMusic's player and effect table (newDriver).
var (
	amigaMX *orig.MX
	amigaFx []int
)

// confirms is the game's Data.Confirms.
var confirms map[string]game.Confirm

// confirmOf is the game's confirm for act (Data.Confirms), else def.
func confirmOf(act string, def game.Confirm) game.Confirm {
	if c, ok := confirms[act]; ok && len(c.Asks) > 0 {
		return c
	}
	return def
}

// itemConfirm is the Data.Confirms key of act k (ActDrop, ActSell) on it:
// the most specific the game gives, else "Drop" or "Sell". Sell's are
// SellMany (a stack), SellFree (no price) and SellFreeMany; Drop's
// DropScribe (a scroll being scribed).
func itemConfirm(k ActKind, it game.Item) string {
	base, keys := "Drop", []string{}
	if k == ActSell {
		base = "Sell"
		free := salePrice(it) == (Money{})
		switch {
		case free && it.Count > 1:
			keys = []string{"SellFreeMany", "SellFree"}
		case free:
			keys = []string{"SellFree"}
		case it.Count > 1:
			keys = []string{"SellMany"}
		}
	} else if scribing(it) {
		keys = []string{"DropScribe"}
	}
	for _, c := range keys {
		if len(confirms[c].Asks) > 0 {
			return c
		}
	}
	return base
}

// confirmText is a Confirm's ask with %s the name and %p the price.
func confirmText(ask, name, price string) string {
	return strings.NewReplacer("%s", name, "%p", price).Replace(ask)
}

// Text is the game's named text (Data.Texts; "": none).
func Text(name string) string { return texts[name] }

// filesDir is where the animations' files are (Data.AnimDir; SequenceStep).
var filesDir string

// The adventure's scripts (SetScripts): the title and the credits (Start),
// and the end (End); nil: none.
var startSeq, endSeq *adventure.Sequence

// SetScripts hands gobox the adventure's Start and End sequences (its
// manifest's; nil: none) before NewFront: gobox composes the title, the
// credits, the menu's backdrop and the end from them.
func SetScripts(start, end *adventure.Sequence) { startSeq, endSeq = start, end }

// scriptOf is the script intro in plays: the End's or the Start's.
func scriptOf(in *game.Intro) *adventure.Sequence {
	if in != nil && in.End {
		return endSeq
	}
	return startSeq
}

// credits are the title's credits lines (the Start's; nil: none).
func credits() []adventure.Credit {
	if startSeq == nil {
		return nil
	}
	return startSeq.Credits
}

// creditBoxes are the original's credits boxes (the Start's; nil: one frame on the page's edges).
func creditBoxes() []adventure.Cells {
	if startSeq == nil {
		return nil
	}
	return startSeq.Boxes
}

// titleBackdrop is the title's picture under the main menu (the Start's Backdrop; nil: none).
func titleBackdrop() []adventure.Layer {
	if startSeq == nil {
		return nil
	}
	return startSeq.Backdrop
}

// advDir is the imported adventure folder (Data.Adventure; "": none).
var advDir string

var daxMask map[string]int // Data.DaxMask

// gameData overlays the game's declarative game.Data (NewFront's) on what
// SetData handed in: its series and its spell tips (by spell id).
func gameData(g game.Data) {
	gameID, daxMask = g.ID, g.DaxMask
	confirms = g.Confirms
	for k, v := range g.Texts { // NewFront's games name their texts here (SetData's Data.Texts too)
		if texts == nil {
			texts = map[string]string{}
		}
		texts[k] = v
	}
	filesDir, advDir = g.AnimDir(), g.Adventure
	GameID, Amiga = g.ID, amigaShown
	AmigaArt = Amiga // the members' pool portraits follow F12 too (SetData set it only for a game giving Data.DAX)
	if g.Series != "" {
		series = g.Series
	}
	Series = series
	Coin = cmp.Or(g.Coin, game.Gold)
	if b := g.Border; b != nil { // the game's look, as plain data
		t := theme
		t.Border = Border{File: b.File, Block: b.Block, Pieces: b.Pieces, Caps: b.Caps, Mask: b.Mask, Edges: b.Edges}
		applyTheme(t)
	}
	switch { // the game's imported art (its import: dax.ImportArt)
	case daxArt.art != nil:
	case g.Adventure != "":
		daxArt.art = asset.OpenArt(filepath.Join(g.Adventure, "art"))
	case g.Files != "":
		daxArt.art = asset.OpenArt(asset.Dir(g.Files))
	}
	if dir := g.SoundDir(); g.AdLib != nil && dir != "" && adLib == nil {
		music, _ := os.ReadFile(filePath(dir, g.AdLib.Music))
		snd, _ := os.ReadFile(filePath(dir, g.AdLib.Sounds))
		if music != nil {
			adLib = &audio.AdLibData{Music: music, Sounds: snd, Effects: g.AdLib.Effects, Rate: g.AdLib.Rate, Tempo: g.AdLib.Tempo, Named: map[string][]byte{}}
			for _, sq := range []*adventure.SequenceStep{startSeq.LeadingAnim(), endSeq.LeadingAnim()} { // the songs the animations cue (NAME.SMS beside them)
				if sq == nil || sq.Song == "" {
					continue
				}
				if b, err := os.ReadFile(filepath.Join(dir, sq.Song+".SMS")); err == nil {
					adLib.Named[sq.Song] = b
				}
			}
		}
	}
	if m := g.AmigaMusic; m != nil && g.SoundDir() != "" && amigaMX == nil {
		dir := filepath.Join(filepath.Dir(g.Files), "amiga")
		if g.Adventure != "" {
			dir = g.SoundDir()
		}
		mx, _ := os.ReadFile(filepath.Join(dir, m.Songs))
		glb, _ := os.ReadFile(filepath.Join(dir, m.Effects))
		if p, err := orig.NewMX(mx, glb); err == nil {
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
	if g.IconParts != nil {
		iconParts = g.IconParts
	}
	if g.Adventure != "" {
		applyLook(g.Adventure)
		menuLogo = readLogo(filepath.Join(g.Adventure, adventure.ArtDir, adventure.LogoFile))
	}
	if g.AreaMaps != nil {
		areaMaps = map[AreaID]string{}
		for id, n := range g.AreaMaps {
			areaMaps[AreaID(id)] = n
		}
	}
	if g.Remaster != nil {
		art = g.Remaster
	}
	if g.OpenAt != "" {
		openAt = g.OpenAt
	}
	if g.Bindings != nil {
		var own []Binding
		for _, b := range g.Bindings {
			own = append(own, Binding{Ctx: b.Ctx, Desc: b.Desc, Keys: b.Keys, Mouse: b.Mouse, Pad: b.Pad})
		}
		setBindings(own)
	}
	menuEntries = slices.DeleteFunc(menuEntries, func(e Entry) bool { return e.Act == ActCommand })
	menuEntries = slices.DeleteFunc(menuEntries, func(e Entry) bool { return e.Seq != "" })
	for _, m := range g.MenuItems {
		menuEntries = append(menuEntries, Entry{Text: m.Label, Key: byte(m.Key), Seq: m.Sequence})
	}
	for i, c := range g.Commands {
		menuEntries = append(menuEntries, Entry{Text: c.Label, Key: byte(c.Key), Act: ActCommand, N: i, Party: true})
	}
	if paneBox == (CellRect{}) { // the pool games' main screen: the message pane, the view's frame and its picture
		paneBox, viewWindow, innerWindow = CellRect{X1: 1, Y1: 17, X2: 38, Y2: 22}, CellRect{X1: 1, Y1: 1, X2: 15, Y2: 15}, CellRect{X1: 3, Y1: 3, X2: 13, Y2: 13}
	}
	if len(table.Rows) == len(bindings) { // no game rows: digits pick members
		setBindings([]Binding{MemberDigits})
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

// filePath is file name in dir, its name's case as on disk ("": none).
func filePath(dir, name string) string {
	ents, _ := os.ReadDir(dir)
	for _, e := range ents {
		if strings.EqualFold(e.Name(), name) {
			return filepath.Join(dir, e.Name())
		}
	}
	return ""
}
