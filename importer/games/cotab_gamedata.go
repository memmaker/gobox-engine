package games

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"

	"gobox/asset"
	"gobox/game"
	"gobox/game/play"
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
			{Label: "Human Change Class", Key: 'H'}, {Label: "View Character", Key: 'V'}, {Label: "Remove Character", Key: 'R'},
			{Label: "Exit to DOS", Key: 'E'}},
		Sounds: map[string]string{}}
	for n, name := range soundNames {
		d.Sounds[name] = strconv.Itoa(n)
	}
	d.Font = "Gold_Box_8x8D1" // the gobox/ui/fonts pool font of cotab's pages (8X8D1 block 201 re-laid out)
	d.Border = &border
	d.AmigaMusic = CotabAmigaMusic                                                // the Amiga's sampled effects, when the import found them (pool.ReadGameData); else the Tandy's
	d.DungeonTiles = CotabDungeonTiles                                            // the stone wall pieces at the start of the combat set (cubes on the board)
	d.ReadyMark = "Yes "                                                          // the item list's readied column (ItemName)
	d.IconParts = []string{"Body", "Arm", "Leg", "Hair/Face", "Shield", "Weapon"} // as the original's icon editor names them
	// camp's Fix and Rest; Rest is a Duration rest: gobox asks the days and hours
	d.Rests = []game.RestKind{{Text: "Fix", Key: 'F', N: RestFix, Fix: true}, {Text: "Rest", Key: 'R', N: RestWait, Duration: true}}
	// the Party submenu's human class change (Pool and Hillsfar imports are Imports' Add from where?)
	d.Commands = append(play.CommandEntries(cotabPartyCommands), game.PartyCommand{Label: "Add from Pool", Key: 'P', From: "por"})
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
}

// AmigaMusic is the Amiga release's sampled effects (SOUNDS.GLB, 8-bit at
// 7418 Hz), played in place of the Tandy's sounds 2..12 (spell .. missile
// end, GLIB entries 0..10). The Amiga has no songs.
var CotabAmigaMusic = &game.AmigaMusic{Effects: "SOUNDS.GLB", EffectOf: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}}

// partyCommands are the game's own party commands.
var cotabPartyCommands = []play.Option{{Label: "New Class", Key: 'N', Do: play.ChangeClass{Member: -1, Class: -1}}}
