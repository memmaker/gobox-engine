package importer

import (
	_ "embed"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"gobox/adventure"
	"gobox/asset"
	"gobox/game"
)

// soundNames name the sound numbers (sound.go) as game Events and
// Data.Sounds do.
var soundNames = map[int]string{sndStop: "stop", sndMute: "mute", sndUnmute: "unmute", sndSpell: "spell", sndHit: "hit",
	sndMagicHit: "magic hit", sndFall: "fall", sndMissile: "missile", sndBump: "bump", sndSpell51: "spell 51", sndMiss: "miss",
	sndStep: "step", sndSpell47: "spell 47", sndMissileEnd: "missile end", sndTitle: "title", sndCountdown: "countdown"}

// StaticData is cotab's data that needs no folder: the series and id, the
// games it adds characters from, the spell tips, the Quick Start party, the
// confirms and texts in the original's words, the party menu, the sounds by
// name, the font and frame, the rests, the combat bar's verbs and the palette.
// The import writes it as data/gamedata.json (pool.GameDataFile).
func CotabStaticData() game.Data {
	d := game.Data{Series: "pool", ID: "cotab", SpellRules: cotabSpellTips(), Imports: []string{"por"}, QuickParty: cotabQuickParty, // Add from Pool of Radiance through the one character format (no Hillsfar codec)
		Confirms: map[string]game.Confirm{ // the original's words (O0AF: dropItem, the scribe check); YesNo starts on No
			"Drop":       {Asks: []string{"Your %s\nwill be gone forever\nDrop It? "}},
			"DropScribe": {Asks: []string{"%s was going to scribe from that scroll\nis it Okay to lose it? "}},
			"Sell":       {Asks: []string{"I'll give you %p gold pieces for your %s\nIs It a Deal? "}}, // O0AF:1489
		},
		Texts: map[string]string{"ready": "Yes ", // the item list's readied column (ItemName)
			"Items": "Use", "End": "Done", "Wait": "Delay"}, // the combat bar's words as combatMenu's and otherMenu's
		PartyMenu: []game.MenuKey{{Label: "Load Saved Game", Key: 'L'}, {Label: "Begin Adventuring", Key: 'B'}, {Label: "Quick Start", Key: 'Q'}, // the core's order (PartyFunctions)
			{Label: "Add Character", Key: 'A'}, {Label: "Create Character", Key: 'C'}, {Label: "Forget Character", Key: 'F'},
			{Label: "Drop Character", Key: 'D'}, {Label: "Modify Character", Key: 'M'}, {Label: "Train Character", Key: 'T'},
			game.ChangeClassFunction, {Label: "View Character", Key: 'V'}, {Label: "Remove Character", Key: 'R'},
			{Label: "Exit to DOS", Key: 'E'}},
		Sounds: map[string]string{}}
	for n, name := range soundNames {
		d.Sounds[name] = strconv.Itoa(n)
	}
	d.Border = &border
	d.AmigaMusic = CotabAmigaMusic                                                // the Amiga's sampled effects, when the import found them (pool.ReadGameData); else the Tandy's
	d.DungeonTiles = adventure.DungeonTiles                                       // the stone wall pieces at the start of the combat set (cubes on the board)
	d.IconParts = []string{"Body", "Arm", "Leg", "Hair/Face", "Shield", "Weapon"} // as the original's icon editor names them
	// camp's Fix and Rest; Rest is a Duration rest: gobox asks the days and hours
	d.Rests = []game.RestKind{{Text: "Fix", Key: 'F', N: adventure.RestFix, Fix: true}, {Text: "Rest", Key: 'R', N: adventure.RestWait, Duration: true}}
	// the Party submenu's Add from Pool (Pool and Hillsfar imports are Imports' Add from where?)
	d.Commands = append(d.Commands, game.PartyCommand{Label: "Add from Pool", Key: 'P', From: "por"})
	// the combat bar's letters, as combatMenu's (Move View Aim Use Cast Turn Quick Done) and otherMenu's
	// (Guard Delay Quit Bandage Speed Exit) words; Attack is Aim's target, not a word of its own
	d.VerbKeys = map[string]rune{"Move": 'M', "View": 'V', "Aim": 'A', "Items": 'U', "Cast": 'C', "Turn": 'T', "Quick": 'Q', "End": 'D',
		"Guard": 'G', "Wait": 'D', "Bandage": 'B'}
	// The original's masked loads: combat icons and figures on colour 0
	// (CPIC: icon's ParseImage; CHEAD/CBODY/COMSPR: the icon loader, PoR's
	// ovr33:01C8), the symbol sets on 13 (WallSymbolSets).
	d.DaxMask = map[string]int{"CHEAD": 0, "CBODY": 0, "COMSPR": 0}
	for i := 1; i <= 6; i++ {
		d.DaxMask[fmt.Sprint("CPIC", i)] = 0
		d.DaxMask[fmt.Sprint("8X8D", i)] = 0xD
	}
	d.VerbOrder = []string{"Move", "View", "Aim", "Use", "Cast", "Turn", "Quick", "Done"} // combatMenu's words; Done's otherMenu is its own Question
	d.MusicPlaces = CotabMusicPlaces()
	for _, c := range asset.EGAPalette {
		d.Palette = append(d.Palette, c[:]...)
	}
	return d
}

// spellTipsText is the spell tips by id: "id<TAB>tip" lines, # comments.
// The original has none; cotab's are written from its own spell table and
// cast routines.
//
//go:embed cotab/spelltips.txt
var spellTipsText string

// spellTips is spellTipsText as Data.SpellRules.
func cotabSpellTips() map[int]string {
	m := map[int]string{}
	for _, l := range strings.Split(spellTipsText, "\n") {
		id, tip, ok := strings.Cut(l, "\t")
		if n, err := strconv.Atoi(id); ok && err == nil {
			m[n] = strings.TrimSpace(tip)
		}
	}
	return m
}

var cotabBorderMask uint8 = 0xD // the frame's transparent colour

// border is R133 draw8x8_01's frame: 8X8D1 block 202 from frame 30 (symbol
// 286), its corner and side pieces (8e7f039 cmd/cotab/frame.go).
var border = game.Border{File: "8X8D1", Block: 202, Mask: &cotabBorderMask,
	Pieces: [8]int{30, 32, 31, 33, 31, 31, 32, 32}, // top-left, top-right, bottom-left, bottom-right, top, bottom, left, right
	// the cracked edges, cell by cell: draw8x8_01's tables (8e7f039^ frame.go), frame 30+symbol
	Edges: map[string][]int{
		"top":    {30, 36, 31, 31, 31, 31, 31, 31, 36, 31, 31, 31, 31, 34, 31, 31, 31, 36, 31, 31, 31, 31, 31, 31, 31, 38, 31, 31, 31, 31, 31, 31, 31, 34, 31, 31, 31, 36, 31, 32},
		"rule":   {30, 38, 31, 31, 31, 31, 31, 31, 31, 31, 31, 36, 31, 31, 31, 38, 34, 31, 31, 31, 36, 31, 31, 31, 31, 31, 31, 31, 31, 31, 34, 31, 36, 31, 31, 31, 31, 31, 38, 32},
		"bottom": {31, 38, 36, 31, 31, 31, 31, 31, 31, 31, 31, 34, 31, 31, 31, 31, 31, 36, 38, 31, 31, 31, 34, 31, 31, 31, 31, 31, 31, 36, 31, 31, 31, 31, 31, 31, 31, 31, 34, 33},
		"left":   {30, 32, 39, 35, 32, 32, 32, 32, 32, 32, 35, 37, 32, 32, 32, 32, 32, 39, 37, 32, 32, 32, 37, 31},
		"right":  {32, 32, 39, 37, 32, 32, 32, 35, 32, 32, 32, 32, 32, 32, 32, 32, 32, 37, 32, 32, 32, 32, 35, 32},
		// the fight's board divider (d0C54, draw8x8_06), every vertical divider
		"divider": {30, 37, 35, 32, 32, 32, 32, 32, 32, 32, 32, 32, 37, 35, 32, 32, 32, 32, 32, 32, 35, 32, 34},
		// "edge@n": the original's own table for an edge n cells long
		"divider@17": {30, 37, 35, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 35, 32, 39, 34},                         // 3D view, column 16 (d0BFA, draw8x8_03)
		"left@23":    {30, 32, 39, 35, 32, 32, 32, 32, 32, 32, 35, 37, 32, 32, 32, 32, 32, 39, 37, 32, 32, 32, 31}, // fight (d0C3D, draw8x8_06)
		"right@23":   {32, 32, 39, 37, 32, 32, 32, 35, 32, 32, 32, 32, 32, 32, 32, 32, 32, 37, 32, 32, 32, 32, 32}, // fight (d0C6B)
		// the 3D view's inner box, cells 2..14 (d0BC6/d0BD3/d0C21/d0C2E at +col, symbols +276)
		"view.top":    {20, 26, 21, 21, 21, 21, 28, 21, 21, 24, 21, 21, 22},
		"view.bottom": {21, 24, 21, 21, 21, 21, 21, 21, 28, 24, 21, 21, 23},
		"view.left":   {20, 22, 27, 22, 22, 22, 22, 25, 22, 22, 22, 22, 21},
		"view.right":  {22, 25, 29, 22, 22, 22, 27, 25, 22, 22, 22, 22, 23},
	},
	Amiga:     cotabAmigaFrame(),
	MarkShift: -10, // the selected card: frames 20..29, the 3D view's yellow box (white 30..39 in yellow)
}

// The Amiga Curse's frame screens (Curse, code hunk 0): ALWAYS.TLB entries
// (amigaimport block = entry, one frame) blitted by F_3bcf8(y, x, lib 0,
// entry). The DOS game draws the same three: the full frame, the 3D view's
// and the one with the rule alone.
const (
	cotabFullFrame = iota // F_398b0 (0x398b0)
	cotabViewFrame        // 0x39990: F_398b0, the 3D view's box and the divider, the rule
	cotabRuleFrame        // 0x39afe: F_398b0, the rule (row 16)
)

// ALWAYS.TLB's frame pieces (their sizes and hot spots from the entries' headers).
const (
	cotabTopBar     = 71  // 320x8
	cotabLeftSide   = 72  // 8x16, repeated
	cotabLeftFoot   = 73  // 8x16
	cotabRightSide  = 74  // 8x16, repeated
	cotabBottomBar  = 75  // 320x16, hot spot y -184
	cotabDivider    = 76  // 8x16, repeated
	cotabDividerTop = 77  // 8x8
	cotabDividerEnd = 78  // 8x16, over the rule
	cotabRule       = 96  // 320x8
	cotabViewBox    = 98  // 104x104, hot spot -16, -16
	cotabRightFoot  = 106 // 8x16
)

// The rows of a side's two chain links (the side pieces' cells, clear of
// the joint at row 4 and of the feet).
const (
	cotabChainLinkA = 5
	cotabChainLinkB = 6
)

// cotabAmigaFrame is the Amiga Curse's frame, its cells where the DOS frame's are.
func cotabAmigaFrame() *game.AmigaFrame {
	at := func(entry, x, y int) game.FrameBlit { return game.FrameBlit{Block: entry, X: x, Y: y} }
	full := []game.FrameBlit{at(cotabTopBar, 0, 0)} // 0x398ce
	for y := 8; y <= 152; y += 16 {                 // 0x398e6..0x39920: d2 1, 3 .. 19, y 8*d2
		full = append(full, at(cotabLeftSide, 0, y), at(cotabRightSide, 312, y))
	}
	full = append(full, at(cotabLeftFoot, 0, 168), at(cotabRightFoot, 312, 168), at(cotabBottomBar, 0, 184)) // 0x39922, 0x39934, 0x39946
	view := append(slices.Clone(full), at(cotabViewBox, 16, 16), at(cotabDividerTop, 128, 0))                // 0x39998, 0x399aa
	for y := 8; y <= 104; y += 16 {                                                                          // 0x399c2..0x399e2: d2 1, 3 .. 13
		view = append(view, at(cotabDivider, 128, y))
	}
	view = append(view, at(cotabRule, 0, 128), at(cotabDividerEnd, 128, 120)) // 0x399e4, 0x399f6
	rule := append(slices.Clone(full), at(cotabRule, 0, 128))                 // 0x39b02
	edges := frameEdges(cotabFullFrame, cotabRuleFrame, cotabViewFrame)
	// the 3D view's box, cells 2..14 (the DOS "view." edges' cells)
	edges["view.top"] = frameRow(cotabViewFrame, 2, 2, 14)
	edges["view.bottom"] = frameRow(cotabViewFrame, 14, 2, 14)
	edges["view.left"] = frameCol(cotabViewFrame, 2, 2, 14)
	edges["view.right"] = frameCol(cotabViewFrame, 14, 2, 14)
	// every other box (the portrait cards, the menus, the text): the chain
	// alone, the top bar's corners and links, the bottom its links upside down
	at8 := func(col, row int, flip bool) game.FrameCell {
		return game.FrameCell{Screen: cotabFullFrame, Col: col, Row: row, FlipY: flip}
	}
	last := frameCols - 1
	for _, e := range []struct {
		key  string
		cell func(col, row int) game.FrameCell
		ends [2][2]int // the corner cells (col, row)
		link [2][2]int // the two links the edge repeats
	}{
		{"box.top", func(c, r int) game.FrameCell { return at8(c, r, false) }, [2][2]int{{0, 0}, {last, 0}}, [2][2]int{{1, 0}, {2, 0}}},
		{"box.bottom", func(c, r int) game.FrameCell { return at8(c, r, true) }, [2][2]int{{0, 0}, {last, 0}}, [2][2]int{{1, 0}, {2, 0}}},
		{"box.left", func(c, r int) game.FrameCell { return at8(c, r, false) }, [2][2]int{{0, 1}, {0, cotabChainLinkA}}, [2][2]int{{0, cotabChainLinkA}, {0, cotabChainLinkB}}},
		{"box.right", func(c, r int) game.FrameCell { return at8(c, r, false) }, [2][2]int{{last, 1}, {last, cotabChainLinkA}}, [2][2]int{{last, cotabChainLinkA}, {last, cotabChainLinkB}}},
	} {
		edges[e.key] = []game.FrameCell{e.cell(e.ends[0][0], e.ends[0][1]), e.cell(e.link[0][0], e.link[0][1]), e.cell(e.link[1][0], e.link[1][1]), e.cell(e.ends[1][0], e.ends[1][1])}
	}
	return &game.AmigaFrame{File: "ALWAYS", Screens: [][]game.FrameBlit{full, view, rule}, Pieces: framePieces(cotabFullFrame), Edges: edges, Mask: alwaysMask}
}

// AmigaMusic is the Amiga release's sampled effects (SOUNDS.GLB, 8-bit at
// 7418 Hz), played in place of the Tandy's sounds 2..12 (spell .. missile
// end, GLIB entries 0..10). The Amiga has no songs.
var CotabAmigaMusic = &game.AmigaMusic{Effects: "SOUNDS.GLB", EffectOf: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}}
