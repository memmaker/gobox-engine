package importer

import (
	"gobox/adventure"
	"gobox/game"
	"slices"
	"sort"
	"strings"
)

// combatMenu is the combat menu at DS:2F8C (O05F:0764); @ joins Quick and Delay's keys.
const combatMenu = "Move Aim Use Cast Turn Guard Quick@Delay Bandage View Speed End"

// StaticData is Pools of Darkness' data without a folder or a title: the import
// writes it as data/gamedata.json (pool.GameDataFile).
// It carries both the AdLib and the Amiga music; pool.ReadGameData keeps the
// Amiga's when the folder's sound/ has its files.
func PodStaticData() game.Data {
	d := podGameData("")
	d.AmigaMusic = PodAmigaMusic
	return d
}

// AmigaMusic is the Amiga's Music-X songs and GLIB effects. The Amiga
// program's Sound(cmd) (file 0x2fb18, a4 0x8638) plays SOUNDS.GLB entry
// cmd-3 for cmd 3..13 and pod.mx song cmd-14 for 14..21; the script's SOUND
// n (file 0x20680) calls Sound(n+1) for 2..14 and Sound(n&^128+13) for a
// song. So effect n is entry n-2 and song n is pod.mx's n-1 (SongOf nil).
// 13 and 14 reach songs 0 and 1 on the Amiga; no script plays them, and here
// they play nothing (EffectOf stops at 12).
var PodAmigaMusic = &game.AmigaMusic{Songs: "pod.mx", Effects: "SOUNDS.GLB", EffectOf: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}}

// gameData is the data that does not come from the folder's manifest, over
// adventure folder adv (gobox import pod; "": none): its id, the games it imports from, a tip for every spell
// (spellTips), the Quick Start party, the font, frame, AdLib music and
// effects, camp's rests and Add from Secret.
func podGameData(adv string) game.Data {
	effects := map[int]int{} // effect N (2..12) is ZOUNDS.DAX block N-1 (docs/MUSIC.md)
	for n := 2; n <= 12; n++ {
		effects[n] = n - 1
	}
	return game.Data{Series: "pool", ID: "pod", PartyMenu: partyMenuOrder(), Imports: []string{importSotSB}, SpellRules: podSpellTips, QuickParty: podQuickParty,
		Adventure:   adv,
		Texts:       map[string]string{"Trade to whom?": "Trade with Whom?"},                        // Trade's member pick (O0AA:1181)
		Confirms:    confirms,                                                                       // the original's own asks (pO108_2254)
		VerbKeys:    map[string]rune{"View": 'V'},                                                   // the combat menu's View (s2F8C): gobox's sheet of the member whose turn it is
		VerbOrder:   strings.Fields(strings.ReplaceAll(combatMenu, "@", " ")),                       // s2F8C's order: View between Bandage and Speed
		VerbChords:  map[string]string{"Quick All": "alt+q,alt+a"},                                  // the combat menu's Alt+Q / Alt+A (O05F:0CF2, scan codes 16, 30): the whole party to the computer; plain Q is the member's own (O05F:0B43)
		MenuItems:   []game.MenuItem{{Label: "Play intro", Key: 'I', Sequence: adventure.SeqIntro}}, // gobox replays the intro from the main menu
		MusicPlaces: musicPlaceList(),
		Scans:       journalBook().Picture, // the journal's pictures a Read shows
		// the screen frame (R142:0000, O17E:02E6): BORDERS.DAX block 0, corners TL TR BL BR,
		// top and bottom, the sides, masked on colour 13 (O17E:004B); BORDERS blocks 0..3 are the
		// four frame sets, game word SymbolSet (tileSet) picking one: Load8x8Set(tileSet, 2), set 4
		// loading 0 (pod-re eclmenu.pas WriteGbvm $39, saveload.pas after a load)
		DaxMask:      daxMask,
		Border:       &game.Border{File: "BORDERS", Block: 0, Pieces: [8]int{0, 1, 2, 3, 4, 4, 5, 5}, Mask: &podBorderMask, Sets: podFrameSets, Amiga: podAmigaFrame()},
		AdLib:        &game.AdLib{Music: adLibMusic, Sounds: adLibSounds, Effects: effects, Rate: 9000},
		DungeonTiles: adventure.DungeonTiles,
		Rests:        []game.RestKind{{Text: "Rest", Key: 'R', N: 0, Duration: true}, {Text: "Fix", Key: 'F', N: 1, Fix: true}},
		Commands:     game.CommandEntries(podPartyCommands),
	}
}

var podBorderMask uint8 = 13

// daxMask is each figure file's transparent colour: pO13E_0182 (O13E:0182)
// loads CHEAD, CBODY, COMSPR, ICON and CPIC1 masked on 8; O079:0637 the
// cursor on 13.
var daxMask = map[string]int{"CHEAD": 8, "CBODY": 8, "COMSPR": 8, "ICON": 8, "CPIC1": 8, "CURSOR": 13}

// musicPlaceList is every distinct musicPlaces trigger, sorted.
func musicPlaceList() []string {
	seen := map[string]bool{}
	var out []string
	for _, p := range podMusicPlaces {
		if !seen[p.slug] {
			seen[p.slug] = true
			out = append(out, p.slug)
		}
	}
	sort.Strings(out)
	return out
}

// partyMenuOrder is the party menu in the original's order (DS 06C0), Quick
// Start before Exit; gobox shows the entries the rules allow.
func partyMenuOrder() []game.MenuKey {
	var o []game.MenuKey
	keyed := func(ls ...string) {
		for _, l := range ls {
			o = append(o, game.MenuKey{Label: l, Key: rune(l[0])})
		}
	}
	keyed("Create Character", "Drop Character", "Modify Character", "Train Character")
	o = append(o, game.ChangeClassFunction)
	keyed("View Character", "Add Character", "Remove Character", "Load Saved Game",
		"Save Current Game", "Begin Adventuring", quickStartLabel, "Exit to DOS")
	return o
}

// confirms is gobox's confirms in the original's DS words: Drop O0AA:1549
// (s13DB), DropScribe O0AA:0741, Sell O0AA:1453 (all four on b5362, Yes:
// Sell/SellMany priced, SellFree/SellFreeMany worthless; one item or a
// stack), DropMember O075:2050 (both on b5363, No).
var confirms = map[string]game.Confirm{
	"Drop":         {Asks: []string{"Lose it? "}},
	"DropScribe":   {Asks: []string{"is it Okay to lose it? "}},
	"Sell":         {Asks: []string{"Will you sell it? "}, Yes: true},
	"SellMany":     {Asks: []string{"Will you sell them? "}, Yes: true},
	"SellFree":     {Asks: []string{"Shall I dispose of it for you? "}, Yes: true},
	"SellFreeMany": {Asks: []string{"Would you like to discard them? "}, Yes: true},
	"DropMember":   {Asks: []string{"Drop %s forever? ", "Are you sure? "}},
}

// partyCommands are the game's own party commands.
var podPartyCommands = []game.Option{{Label: "Add from Secret", Key: 'S', Do: game.AddCharacter{Roster: -1, From: importSotSB}}}

// podFrameSets are BORDERS.DAX's frame sets' blocks by tileSet.
var podFrameSets = []int{0, 1, 2, 3}

// The Amiga Pools of Darkness' frame screens (Pools of Darkness, code hunk
// 0): F_405e6(y, x, lib 1, entry) blits pieces of the frame set loaded as
// library 1, FRAME.TLB entry tileSet+1 (F_3e1bc, 0x3e1bc: party byte 0x3A,
// 4 loading set 0, as DOS); amigaimport makes each set block set+1, its
// pieces frames entry-1 (the set's palette, entry 0, is no frame). Every
// blit passes 0, 0: the pieces' hot spots place them. 0x3df08(row) draws
// the rule of row 2, 8, 16 or 20 (entries 7..10).
const (
	podFullFrame = iota // 0x3df80
	podViewFrame        // 0x3e02e: 0x3df80, the rule (row 16), the 3D view's surround with the divider
	podRuleFrame        // 0x3e05c: 0x3df80, the rule (row 16)
)

// FRAME.TLB's frame pieces, per set (their sizes and hot spots from the entries' headers).
const (
	podTopBar    = 1  // 320x8
	podLeftBar   = 2  // 8x184
	podRightBar  = 3  // 8x184, hot spot x -312
	podBottomBar = 4  // 320x16, hot spot y -184
	podRule16    = 9  // 320x7, hot spot y -128
	podViewBox   = 12 // 136x136: the 3D view's surround, the divider its right edge
)

// podAmigaFrameSets are the Amiga FRAME blocks by tileSet (FRAME.TLB entries 1..4).
var podAmigaFrameSets = []int{1, 2, 3, 4}

// podAmigaFrame is the Amiga Pools of Darkness' frame (set 0's blocks; Sets
// the others). Its 3D view's box takes the frame's general edges, as on DOS.
func podAmigaFrame() *game.AmigaFrame {
	at := func(entry, x, y int) game.FrameBlit {
		return game.FrameBlit{Block: podAmigaFrameSets[0], Frame: entry - 1, X: x, Y: y}
	}
	full := []game.FrameBlit{at(podTopBar, 0, 0), at(podLeftBar, 0, 0), at(podRightBar, 312, 0), at(podBottomBar, 0, 184)} // 0x3dfaa .. 0x3dfe6
	rule := at(podRule16, 0, ruleRow*frameCell)                                                                            // 0x3df3e
	view := append(slices.Clone(full), rule, at(podViewBox, 0, 0))                                                         // 0x3e036, 0x3e04a
	ruled := append(slices.Clone(full), rule)                                                                              // 0x3e064
	return &game.AmigaFrame{File: "FRAME", Screens: [][]game.FrameBlit{full, view, ruled}, Pieces: framePieces(podFullFrame),
		Edges: frameEdges(podFullFrame, podRuleFrame, podViewFrame), Sets: podAmigaFrameSets}
}
