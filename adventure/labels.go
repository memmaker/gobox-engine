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

// The soundtrack songs the engine itself plays by role (extra). A game's
// AreaMusic.Files names the recording of each number it has; the role
// numbers are the first soundtrack's (POR's Famicom cartridge order), the
// game's other songs are its own numbers in its AreaMusic.
const (
	SongTitle        = 1
	SongCombat       = 11
	SongWilderness   = 13
	SongShop         = 14
	SongEncounter    = 17
	SongVictory      = 18 // with the fanfare
	SongTemple       = 20
	SongInn          = 21
	SongBattleBegins = 27 // a two-second sting
)

// AreaMusic is a game's soundtrack (GameTables.Music; nil: none): its
// recordings, each GEO block's song and the songs of the battles fought there.
type AreaMusic struct {
	Files       map[int]string // song -> its recording's name in the soundtrack folder (without .ogg)
	AmigaSongs  []int          // the Amiga song Music starts -> its soundtrack song (0: none)
	Songs       map[int]int    // GEO block -> its song
	BattleSongs map[int]int    // GEO block -> its battles' song; none: A Battle Begins into the combat theme
}
