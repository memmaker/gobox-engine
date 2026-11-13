package importer

import (
	"strconv"
	"strings"

	"gobox/adventure"
	"gobox/game"
)

// gameData is POR's declarative data for gobox (game.Data), written as
// data/gamedata.json: the series, the game id and the spell tips (SpellRules by spell id).
func porGameData() game.Data {
	d := game.Data{Series: "pool", ID: "por", SpellRules: map[int]string{}}
	d.VerbChords = map[string]string{"Manual": "space", "Magic": "alt+m"}  // the computer turn's poll (ovr10:0FC1): Space takes the party back, Alt-M switches its casting
	d.VerbOrder = []string{"Move", "Use", "Cast", "Turn", "Quick", "Done"} // combatMenu's bar (ovr09:069E-07D4) less View and Aim (Felix 2026-11-13)
	// Quick Start: two fighters in front, two clerics, two magic-users (Sleep is
	// the spell that wins the early fights) and a thief for locks and traps.
	d.QuickParty = []game.QuickPick{
		{Race: "Dwarf", Gender: "Male", Class: "Fighter", Align: "Lawful Good", Need: map[string]int{"str": 18, "con": 17}, Name: "Thorgrim", Head: "head_009", Body: "body_004"},
		{Race: "Human", Gender: "Female", Class: "Cleric", Align: "Lawful Good", Need: map[string]int{"wis": 17, "con": 15, "str": 15}, Name: "Mirelle", Head: "head_011", Body: "body_009"},
		{Race: "Half-Elf", Gender: "Male", Class: "Cleric/Fighter", Align: "Neutral Good", Need: map[string]int{"str": 17, "wis": 16, "con": 15}, Name: "Aldric", Head: "head_018", Body: "body_011"},
		{Race: "Elf", Gender: "Female", Class: "Fighter/Magic-User", Align: "Chaotic Good", Need: map[string]int{"str": 16, "int": 17, "dex": 16}, Name: "Sylvaine", Head: "head_022", Body: "body_012"},
		{Race: "Human", Gender: "Male", Class: "Magic-User", Align: "Neutral Good", Need: map[string]int{"int": 18, "dex": 16, "con": 14}, Name: "Corwin", Head: "head_008", Body: "body_016"},
		{Race: "Halfling", Gender: "Male", Class: "Fighter/Thief", Align: "Neutral Good", Need: map[string]int{"str": 16, "dex": 17, "con": 15}, Name: "Pip", Head: "head_025", Body: "body_001"},
	}

	d.IconParts, d.DungeonTiles = adventure.IconParts[:], adventure.DungeonTiles
	d.Rests = []game.RestKind{ // camp's rests: until fixed while someone heals or studies, an hour, until night or day
		{Text: "Until fixed", Key: 'F', N: adventure.RestFix, Fix: true},
		{Text: "Wait an hour", Key: 'W', N: adventure.RestWait},
		{Text: "Until night", Key: 'N', N: adventure.RestUntil, Night: "Until day", NightKey: 'D'},
	}
	d.Border = &game.Border{Amiga: porAmigaFrame()} // the DOS border is the engine's symbols (ArtNames.BorderCells)
	d.Bindings = []game.Binding{{Ctx: "Any time", Keys: "F12", Desc: "art: Amiga, DOS", Mouse: "Esc>adventure.Graphics", Pad: "B, then pick"}}
	for id, t := range SpellRules {
		if t != "" {
			d.SpellRules[id] = t
		}
	}
	return d
}

// The Amiga Pool's window frame (porFrameBitMaps): archive FRAME, a block a
// piece, one frame each.
const (
	porFrameFile   = "frame"
	porFrameTop    = 1
	porFrameBottom = 2
	porFrameRight  = 3
	porFrameLeft   = 4
	porFrameCorner = 5
)

// Every window the Amiga Pool draws goes through 0x579f4 (clear, frame,
// title), its frame through 0x576d4(x0, y0, x1, y1), the inside's cells:
// the top and bottom bar from cell x0, the sides from row y0, a corner at
// each. The screen is the whole screen's window (1, 1, 38, 23), the bars
// and sides start one cell in.
const (
	porScreenX1 = frameCols - 1 // the right column
	porScreenY1 = frameRows - 1 // the bottom row
)

// porAmigaFrame is the Amiga Pool's frame. Its bars and sides are drawn
// from the window's own corner, not the screen's, so every edge length
// has its own table (key@n, n cells with the corners): corner, the bar's
// first n-2 cells, corner.
func porAmigaFrame() *game.AmigaFrame {
	const s = 0
	at := func(block, x, y int) game.FrameBlit { return game.FrameBlit{Block: block, X: x, Y: y} }
	right, bottom := porScreenX1*frameCell, porScreenY1*frameCell
	screen := []game.FrameBlit{ // 0x576d4's blits, in order
		at(porFrameTop, frameCell, 0),         // 0x57750
		at(porFrameBottom, frameCell, bottom), // 0x577a6
		at(porFrameLeft, 0, frameCell),        // 0x57800
		at(porFrameRight, right, frameCell),   // 0x57856
		at(porFrameCorner, 0, 0),              // 0x5789c
		at(porFrameCorner, right, 0),          // 0x578de
		at(porFrameCorner, 0, bottom),         // 0x57924
		at(porFrameCorner, right, bottom),     // 0x57966
	}
	row := func(r, n int) []game.FrameCell {
		return append(append(frameRow(s, r, 0, 0), frameRow(s, r, 1, n-2)...), frameRow(s, r, porScreenX1, porScreenX1)...)
	}
	col := func(c, n int) []game.FrameCell {
		return append(append(frameCol(s, c, 0, 0), frameCol(s, c, 1, n-2)...), frameCol(s, c, porScreenY1, porScreenY1)...)
	}
	edges := map[string][]game.FrameCell{}
	for n := 2; n <= frameCols; n++ {
		edges["top@"+strconv.Itoa(n)] = row(0, n)
		edges["bottom@"+strconv.Itoa(n)] = row(porScreenY1, n)
		edges["rule@"+strconv.Itoa(n)] = row(0, n) // an inner window's top and bottom: the bars are the same pixels
	}
	for n := 2; n <= frameRows; n++ {
		edges["left@"+strconv.Itoa(n)] = col(0, n)
		edges["right@"+strconv.Itoa(n)] = col(porScreenX1, n)
	}
	return &game.AmigaFrame{File: strings.ToUpper(porFrameFile), Screens: [][]game.FrameBlit{screen}, Pieces: framePieces(s), Edges: edges}
}
