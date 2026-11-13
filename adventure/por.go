package adventure

// Label kinds, each with its colour on the maps (LabelInks, present.Data).
const (
	PoiUser   = 0
	PoiShop   = 1
	PoiTemple = 2
	PoiInn    = 3
	PoiCivic  = 4
)

// MapLabel is a fixed label of a map square, the game's data (GameTables.MapLabels;
// POR: New Phlan's buildings): its text and kind, and the soundtrack song
// played there (0: the area's).
type MapLabel struct {
	Area int // areaKey: data set<<16 | GEO block
	X, Y int
	Text string
	Kind byte // Poi*
	Song int
}

// PorMapLabels are the buildings of New Phlan (data set 3, GEO block 0;
// ECL3 block 0): the squares whose event id (cell byte & 0x7F, the
// script's [6E82]) leads to a shop, temple, inn, tavern or office in the
// block's step event (ECL 9B4C ONGOTO).
var PorMapLabels = func() []MapLabel {
	var m []MapLabel
	add := func(text string, kind byte, song int, sq ...[2]int) {
		for _, s := range sq {
			m = append(m, MapLabel{Area: 3 << 16, X: s[0], Y: s[1], Text: text, Kind: kind, Song: song})
		}
	}
	add("Boats", PoiCivic, 0, [2]int{15, 1})                                                // 1
	add("Harbor master", PoiCivic, 0, [2]int{11, 1})                                        // 2
	add("Bishop Braccio", PoiTemple, SongTemple, [2]int{10, 5})                             // 5
	add("Temple of Tyr", PoiTemple, SongTemple, [2]int{11, 3}, [2]int{8, 4}, [2]int{8, 6})  // 6
	add("Temple of Sune", PoiTemple, SongTemple, [2]int{1, 1}, [2]int{3, 1}, [2]int{1, 3})  // 7
	add("Inn", PoiInn, SongInn, [2]int{4, 12}, [2]int{6, 12}, [2]int{4, 13}, [2]int{6, 13}, // 9
		[2]int{0, 14}, [2]int{1, 14}, [2]int{2, 14})
	add("Training Hall", PoiCivic, SongTrainingHall, [2]int{5, 0}, [2]int{7, 0}, [2]int{9, 0})  // 10, 17
	add("Shop", PoiShop, SongShop, [2]int{15, 8}, [2]int{9, 10}, [2]int{12, 10}, [2]int{9, 11}, // 19
		[2]int{11, 11})
	add("Tavern", PoiInn, SongTavern, [2]int{10, 8}, [2]int{8, 9}, [2]int{14, 10}, [2]int{15, 10}, // 20
		[2]int{15, 12}, [2]int{15, 13}, [2]int{15, 14})
	add("Jeweler", PoiShop, SongShop, [2]int{8, 10})                                       // 21
	add("Arms and Armor", PoiShop, SongShop, [2]int{13, 8}, [2]int{8, 11}, [2]int{11, 12}, // 22
		[2]int{8, 13}, [2]int{9, 13})
	add("Silver Shop", PoiShop, SongShop, [2]int{11, 10}, [2]int{10, 13})                            // 23
	add("Temple of Tempus", PoiTemple, SongTemple, [2]int{0, 12}, [2]int{2, 12})                     // 24
	add("Gate to slums", PoiCivic, 0, [2]int{0, 4})                                                  // 25
	add("City Hall", PoiCivic, SongCityHall, [2]int{4, 4}, [2]int{4, 3}, [2]int{3, 4}, [2]int{5, 4}) // 27, 26
	return m
}()

// The soundtrack's songs, numbered in the Famicom cartridge's order.
const (
	SongTitle        = 1
	SongNewPhlan     = 2
	SongDungeon1     = 3
	SongDungeon2     = 4
	SongDiogenes     = 5
	SongDungeon3     = 6
	SongYarash       = 7
	SongGraveyard    = 8
	SongValjevo      = 9 // Stojanow Gate and Valjevo Castle
	SongLair         = 10
	SongCombat       = 11
	SongTyranthraxus = 12
	SongWilderness   = 13
	SongShop         = 14
	SongEnding       = 15
	SongGameOver     = 16
	SongEncounter    = 17
	SongVictory      = 18 // with the fanfare; 19 is the loop alone
	SongTemple       = 20
	SongInn          = 21
	SongTrainingHall = 24 // with the fanfare; 22 is the loop alone
	SongCityHall     = 23
	SongTavern       = 25
	SongGypsy        = 26
	SongBattleBegins = 27 // a two-second sting
)

// AreaMusic is a game's soundtrack by area (GameTables.Music; nil: none):
// each GEO block's song, and the songs of the battles fought there.
type AreaMusic struct {
	Songs       map[int]int // GEO block -> its song
	BattleSongs map[int]int // GEO block -> its battles' song; none: A Battle Begins into the combat theme
}

// PorMusic (extra) is POR's: the cartridge's own area table (bank 59 $8C46)
// where its areas match the PC's, the themes' names elsewhere; Tyranthraxus's
// lair (GEO 7) has his own battle song.
var PorMusic = AreaMusic{
	Songs: map[int]int{
		0: SongNewPhlan, 1: SongDungeon3, 2: SongDungeon1,
		3: SongValjevo, 4: SongValjevo, 5: SongValjevo, 6: SongValjevo, 7: SongLair,
		9: SongValjevo, 10: SongGraveyard, 13: SongDungeon1, 14: SongDungeon2, 15: SongDungeon1,
		16: SongYarash, 17: SongGypsy, 18: SongDungeon1, 19: SongDiogenes, 20: SongDungeon1,
		21: SongDungeon2, 22: SongYarash, 23: SongYarash, 24: SongDungeon2,
		25: SongWilderness, 26: SongWilderness, 27: SongWilderness, 28: SongDungeon3,
		29: SongDungeon1, 30: SongYarash, 31: SongDungeon2, 32: SongDungeon1,
	},
	BattleSongs: map[int]int{7: SongTyranthraxus},
}
