package importer

import (
	"gobox/adventure"
	"gobox/game"
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
		Texts:       map[string]string{"Trade to whom?": "Trade with Whom?"},              // Trade's member pick (O0AA:1181)
		Confirms:    confirms,                                                             // the original's own asks (pO108_2254)
		VerbKeys:    map[string]rune{"View": 'V'},                                         // the combat menu's View (s2F8C): gobox's sheet of the member whose turn it is
		VerbOrder:   strings.Fields(strings.ReplaceAll(combatMenu, "@", " ")),             // s2F8C's order: View between Bandage and Speed
		VerbChords:  map[string]string{"Quick All": "alt+q,alt+a"},                        // the combat menu's Alt+Q / Alt+A (O05F:0CF2, scan codes 16, 30): the whole party to the computer; plain Q is the member's own (O05F:0B43)
		MenuItems:   []game.MenuItem{{Label: "Play intro", Key: 'I', Sequence: seqIntro}}, // gobox replays the intro from the main menu
		MusicPlaces: musicPlaceList(),
		Scans:       journalBook().Picture, // the journal's pictures a Read shows
		// the screen frame (R142:0000, O17E:02E6): BORDERS.DAX block 0, corners TL TR BL BR,
		// top and bottom, the sides, masked on colour 13 (O17E:004B)
		DaxMask:      daxMask,
		Border:       &game.Border{File: "BORDERS", Block: 0, Pieces: [8]int{0, 1, 2, 3, 4, 4, 5, 5}, Mask: &podBorderMask},
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
	for _, l := range []string{"Create Character", "Drop Character", "Modify Character", "Train Character",
		"Human Change Class", "View Character", "Add Character", "Remove Character", "Load Saved Game",
		"Save Current Game", "Begin Adventuring", quickStartLabel, "Exit to DOS"} {
		o = append(o, game.MenuKey{Label: l, Key: rune(l[0])})
	}
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
