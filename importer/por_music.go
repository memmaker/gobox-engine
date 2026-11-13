package importer

import "gobox/adventure"

// POR's soundtrack songs beyond the engine's roles (adventure.Song*),
// numbered in the Famicom cartridge's order.
const (
	songNewPhlan     = 2
	songDungeon1     = 3
	songDungeon2     = 4
	songDiogenes     = 5
	songDungeon3     = 6
	songYarash       = 7
	songGraveyard    = 8
	songValjevo      = 9 // Stojanow Gate and Valjevo Castle
	songLair         = 10
	songTyranthraxus = 12
	songEndingTheme  = 15
	songGameOver     = 16
	songVictoryLoop  = 19 // adventure.SongVictory's loop alone
	songTrainingLoop = 22 // songTrainingHall's loop alone
	songCityHall     = 23
	songTrainingHall = 24 // with the fanfare
	songTavern       = 25
	songGypsy        = 26
)

// porArt is POR's ArtNames (Init, ovr11:1244, ovr33:0022).
var porArt = adventure.ArtNames{
	TileSheets:    []adventure.TileLoad{{File: "SQRPACI", Block: 1}, {File: "SQRPACI", Block: 2, First: 0x80}},
	DungeonTiles:  []adventure.TileLoad{{File: "DUNGCOM", Block: 1, Last: 0x18}, {File: "RANDCOM", Block: 1, First: 0x22, Last: 5}},
	WildTiles:     []adventure.TileLoad{{File: "WILDCOM", Block: 1, Last: 0x21}, {File: "RANDCOM", Block: 1, First: 0x22, Last: 5}},
	AreaTiles:     []adventure.TileLoad{{File: "BACPAC", Block: 1, Last: 0x28}},
	AreaIcon:      "ICON",
	CombatSprites: "COMSPR",
	Title:         "TITLE",
	BorderCells:   []int{0x114, 0x115, 0x116},
}

// porMapLabels are the buildings of New Phlan (data set 3, GEO block 0;
// ECL3 block 0): the squares whose event id (cell byte & 0x7F, the
// script's [6E82]) leads to a shop, temple, inn, tavern or office in the
// block's step event (ECL 9B4C ONGOTO).
var porMapLabels = func() []adventure.MapLabel {
	var m []adventure.MapLabel
	add := func(text string, kind byte, song int, sq ...[2]int) {
		for _, s := range sq {
			m = append(m, adventure.MapLabel{Area: 3 << 16, X: s[0], Y: s[1], Text: text, Kind: kind, Song: song})
		}
	}
	add("Boats", adventure.PoiCivic, 0, [2]int{15, 1})                                                          // 1
	add("Harbor master", adventure.PoiCivic, 0, [2]int{11, 1})                                                  // 2
	add("Bishop Braccio", adventure.PoiTemple, adventure.SongTemple, [2]int{10, 5})                             // 5
	add("Temple of Tyr", adventure.PoiTemple, adventure.SongTemple, [2]int{11, 3}, [2]int{8, 4}, [2]int{8, 6})  // 6
	add("Temple of Sune", adventure.PoiTemple, adventure.SongTemple, [2]int{1, 1}, [2]int{3, 1}, [2]int{1, 3})  // 7
	add("Inn", adventure.PoiInn, adventure.SongInn, [2]int{4, 12}, [2]int{6, 12}, [2]int{4, 13}, [2]int{6, 13}, // 9
		[2]int{0, 14}, [2]int{1, 14}, [2]int{2, 14})
	add("Training Hall", adventure.PoiCivic, songTrainingHall, [2]int{5, 0}, [2]int{7, 0}, [2]int{9, 0})            // 10, 17
	add("Shop", adventure.PoiShop, adventure.SongShop, [2]int{15, 8}, [2]int{9, 10}, [2]int{12, 10}, [2]int{9, 11}, // 19
		[2]int{11, 11})
	add("Tavern", adventure.PoiInn, songTavern, [2]int{10, 8}, [2]int{8, 9}, [2]int{14, 10}, [2]int{15, 10}, // 20
		[2]int{15, 12}, [2]int{15, 13}, [2]int{15, 14})
	add("Jeweler", adventure.PoiShop, adventure.SongShop, [2]int{8, 10})                                       // 21
	add("Arms and Armor", adventure.PoiShop, adventure.SongShop, [2]int{13, 8}, [2]int{8, 11}, [2]int{11, 12}, // 22
		[2]int{8, 13}, [2]int{9, 13})
	add("Silver Shop", adventure.PoiShop, adventure.SongShop, [2]int{11, 10}, [2]int{10, 13})                  // 23
	add("Temple of Tempus", adventure.PoiTemple, adventure.SongTemple, [2]int{0, 12}, [2]int{2, 12})           // 24
	add("Gate to slums", adventure.PoiCivic, 0, [2]int{0, 4})                                                  // 25
	add("City Hall", adventure.PoiCivic, songCityHall, [2]int{4, 4}, [2]int{4, 3}, [2]int{3, 4}, [2]int{5, 4}) // 27, 26
	return m
}()

// porMusic (extra) is POR's soundtrack: free music from OpenGameArt in the
// adventure's soundtrack folder (CREDITS.txt there names them); the Amiga
// songs' counterparts (title the title, battle the combat theme, the party
// menu the training hall, Phlan New Phlan, party killed the game over, the
// overland map the wilderness); the cartridge's own area table (bank 59
// $8C46) where its areas match the PC's, the themes' names elsewhere;
// Tyranthraxus's lair (GEO 7) has his own battle song.
var porMusic = adventure.AreaMusic{
	Files: map[int]string{
		adventure.SongTitle: "title", songNewPhlan: "newphlan", songDungeon1: "dungeon1", songDungeon2: "dungeon2",
		songDiogenes: "diogenes", songDungeon3: "dungeon3", songYarash: "yarash", songGraveyard: "graveyard",
		songValjevo: "valjevo", songLair: "lair", adventure.SongCombat: "combat", songTyranthraxus: "tyranthraxus",
		adventure.SongWilderness: "wilderness", adventure.SongShop: "shop", songEndingTheme: "ending", songGameOver: "gameover",
		adventure.SongEncounter: "encounter", adventure.SongVictory: "victory", songVictoryLoop: "victory", adventure.SongTemple: "temple", adventure.SongInn: "inn",
		songTrainingLoop: "traininghall", songTrainingHall: "traininghall", songCityHall: "cityhall", songTavern: "tavern",
		songGypsy: "gypsy", adventure.SongBattleBegins: "battlebegins",
	},
	AmigaSongs: []int{0: 0, 1: adventure.SongTitle, 2: adventure.SongCombat, 3: songTrainingHall, 4: songNewPhlan, 5: songGameOver, 6: adventure.SongWilderness},
	Songs: map[int]int{
		0: songNewPhlan, 1: songDungeon3, 2: songDungeon1,
		3: songValjevo, 4: songValjevo, 5: songValjevo, 6: songValjevo, 7: songLair,
		9: songValjevo, 10: songGraveyard, 13: songDungeon1, 14: songDungeon2, 15: songDungeon1,
		16: songYarash, 17: songGypsy, 18: songDungeon1, 19: songDiogenes, 20: songDungeon1,
		21: songDungeon2, 22: songYarash, 23: songYarash, 24: songDungeon2,
		25: adventure.SongWilderness, 26: adventure.SongWilderness, 27: adventure.SongWilderness, 28: songDungeon3,
		29: songDungeon1, 30: songYarash, 31: songDungeon2, 32: songDungeon1,
	},
	BattleSongs: map[int]int{7: songTyranthraxus},
}
