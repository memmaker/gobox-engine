package importer

import "gobox/game"

// The front's data of BR1 (game.Data's declarative fields). The values are
// buckrogers-go presentdata/data.go's (the front's data of the old engine,
// which gobox does not import), the evidence named at each.

// brDaxMask is pO157_0166's masked load: COMSPR, ICON, CHARS and CPIC<disk> on colour 8 (presentdata daxMask).
var brDaxMask = map[string]int{"CHARS": 8, "COMSPR": 8, "ICON": 8, "CPIC1": 8, "CPIC2": 8, "CPIC3": 8, "CPIC4": 8, "CPIC5": 8, "CPIC6": 8}

// brBorderMask is the colour pO1B0_004F loads the border's symbols masked on (presentdata borderMask).
var brBorderMask uint8 = 13

// brEdgeSpan is the edge tables' length (presentdata edgeSpan): gobox draws a shorter edge from a table's head and tail.
const brEdgeSpan = 256

// brEdge is pO1B0_047E's line of n cells: the start piece, a bar, the skipped cell (the start's stud), bars, the skipped
// cell (the end's stud), a bar and the end piece (presentdata edge).
func brEdge(start, bar, stud, end, n int) []int {
	l := make([]int, n)
	for i := range l {
		l[i] = bar
	}
	l[0], l[2], l[n-3], l[n-1] = start, stud, stud, end
	return l
}

// brBorder is the original's frame: BORDERS.DAX block 0 (symbol slot 6, symbols 391..402 as frames 0..11), drawn by
// pO1B0_047E along the lines of d0DDC and pR15F_00FF (a rule, ends 9 and 7); frames: 0..3 the corners, 4 and 5 the bars,
// 6..9 the tees, 10 and 11 the studs (presentdata border).
func brBorder() *game.Border {
	return &game.Border{File: "BORDERS", Block: 0, Mask: &brBorderMask,
		Pieces: [8]int{0, 1, 3, 2, 4, 4, 5, 5}, // top-left, top-right, bottom-left, bottom-right, top, bottom, left, right
		Edges: map[string][]int{
			"top":     brEdge(0, 4, 10, 1, brEdgeSpan),
			"bottom":  brEdge(3, 4, 10, 2, brEdgeSpan),
			"rule":    brEdge(9, 4, 10, 7, brEdgeSpan),
			"left":    brEdge(0, 5, 11, 3, brEdgeSpan),
			"right":   brEdge(1, 5, 11, 2, brEdgeSpan),
			"divider": brEdge(6, 5, 11, 8, brEdgeSpan),
		},
	}
}

// brPartyMenu is BR1's team menu order and hotkeys (GAME.OVR), by gobox's function labels (presentdata partyMenu).
var brPartyMenu = []game.MenuKey{{Label: "Create Character", Key: 'C'}, {Label: "Drop Character", Key: 'P'}, {Label: "Modify Character", Key: 'M'},
	{Label: "Icon Select", Key: 'I'}, {Label: "Train Character", Key: 'T'}, {Label: "View Character", Key: 'V'}, {Label: "Add Character", Key: 'A'},
	{Label: "Remove Character", Key: 'R'}, {Label: "Load Saved Game", Key: 'L'}, {Label: "Save Current Game", Key: 'S'}, {Label: "Begin Adventuring", Key: 'B'},
	{Label: "Quick Start", Key: 'Q'}, {Label: "Exit to DOS", Key: 'E'}}

// brMusicPlaces is every Where.Music trigger of BR1 (buckrogers-go rules/music.go MusicPlaces).
var brMusicPlaces = []string{"neo-chicago", "salvation", "derelict", "space", "asteroid-mine", "ceres-base", "pirate-base",
	"ship-boarding", "mars-surface", "mars-gradivus", "venus-surface", "venus-base", "mercury", "spaceport", "mariposa-doomsday"}

// brQuickParty is the Quick Start team: one of each career, a second Warrior up front (buckrogers-go rules/quickstart.go
// QuickParty; BR draws combat icons, so Head and Body stay empty).
var brQuickParty = []game.QuickPick{
	{Race: "Desert Runner", Gender: "Male", Class: "Warrior", Need: map[string]int{"str": 17, "con": 16}, Name: "Rook"},
	{Race: "Mercurian", Gender: "Female", Class: "Warrior", Need: map[string]int{"str": 15, "dex": 16}, Name: "Kade"},
	{Race: "Terran", Gender: "Male", Class: "Rocket Jock", Need: map[string]int{"dex": 16, "int": 13}, Name: "Ace"},
	{Race: "Martian", Gender: "Female", Class: "Rogue", Need: map[string]int{"dex": 16, "cha": 14}, Name: "Vex"},
	{Race: "Tinker", Gender: "Female", Class: "Engineer", Need: map[string]int{"tech": 17, "int": 12}, Name: "Gizmo"},
	{Race: "Venusian", Gender: "Male", Class: "Medic", Need: map[string]int{"wis": 16, "int": 14}, Name: "Doc"},
}

// br1GameData is BR1's game.Data for gamedata.json. Not here: Art (json:"-": the engine builds the terrain and
// marker pictures, engine.overlandArt) and Scans (the engine's journal, Engine.Data).
func br1GameData(id string) game.Data {
	return game.Data{Series: brSeries, ID: id, Border: brBorder(), QuickParty: brQuickParty, DaxMask: brDaxMask,
		Texts:       map[string]string{"Drop Character": "Purge Character"}, // gdconsts.pas menuTable: P is Purge Character (DropCharacterMenu: deletes), R Remove Character from Team (saves it)
		PartyMenu:   brPartyMenu,
		VerbChords:  map[string]string{"Quick All": "alt+q,alt+a"}, // O065 @0839: Alt-Q, Alt-A quick the whole party
		Coin:        game.Credits,
		MusicPlaces: brMusicPlaces,
	}
}

// brSilentSound is the sound tables of a game whose original sound driver the import does not decode (adventure.SoundFile):
// one song, song 0, that starts no channel, and an empty program. BR1's START.EXE tables have no layout yet (none of the
// old engine's front played them either: it had only the player's own music, by Data.MusicPlaces). The shared launch
// (audio.NewDriver) needs a table; this one plays nothing.
const brSilentSound = `{"program":[{}],"speaker":[[0,0,0,0]],"tandy":[[0,0,0,0]],"vibratos":[],"envelopes":[],` +
	`"speakerPeriods":[0,0,0,0,0,0,0,0,0,0,0,0],"tandyPeriods":[0,0,0,0,0,0,0,0,0,0,0,0]}`

// br2PartyMenu is BR2's team menu: BR1's with "Begin Adventuring" before "Quick Start" (buckrogers2-go rules/data.go partyMenuKeys).
var br2PartyMenu = []game.MenuKey{{Label: "Create Character", Key: 'C'}, {Label: "Drop Character", Key: 'P'}, {Label: "Modify Character", Key: 'M'},
	{Label: "Icon Select", Key: 'I'}, {Label: "Train Character", Key: 'T'}, {Label: "View Character", Key: 'V'}, {Label: "Add Character", Key: 'A'},
	{Label: "Remove Character", Key: 'R'}, {Label: "Load Saved Game", Key: 'L'}, {Label: "Save Current Game", Key: 'S'}, {Label: "Quick Start", Key: 'Q'},
	{Label: "Begin Adventuring", Key: 'B'}, {Label: "Exit to DOS", Key: 'E'}}

// br2MusicPlaces is every Where.Music trigger of BR2 (buckrogers2-go rules/data.go MusicPlaces).
var br2MusicPlaces = []string{"space", "caloris", "salvation", "ship-boarding", "losangelorg", "ceres-base", "venus", "luna",
	"maelstrom-rider", "prison", "purge-hq", "research-base", "living-ship", "fungus-asteroid", "floating-school",
	"genetics-foundation", "mining-rig"}

// br2GameData is BR2's game.Data (buckrogers2-go rules/data.go Data): BR1's border, texts, chords, coin and quick party (the old BR2
// engine's QuickParty is BR1's), BR2's masks, party menu, "Aid" key and music places.
func br2GameData(id string) game.Data {
	d := br1GameData(id)
	d.DaxMask = map[string]int{"COMSPR": 8, "CHARS": 8, "CPIC1": 8, "CURSOR": 13, "8X8D0": 13, "8X8D1": 13, "BORDERS": 13} // O0E5 chead_cbody_comspr_icon, O139:00F5, cursor load
	d.VerbKeys = map[string]rune{"Aid": 'H'}
	d.PartyMenu = br2PartyMenu
	d.MusicPlaces = br2MusicPlaces
	return d
}
