package importer

// quickStartLabel is the party menu's Quick Start option.
const quickStartLabel = "Quick Start"

// importSotSB is the game id of Silver Blades, whose rosters and saves
// PoD imports.
const importSotSB = "sotsb"

// musicPlace is the music trigger of an ECL1 area block; town marks a city's streets.
type podMusicPlace struct {
	slug string
	town bool
}

// musicPlaces maps ECL1 area blocks to their place trigger; floors of one dungeon share one.
var podMusicPlaces = map[int]podMusicPlace{
	16: {"new-phlan", true}, 23: {"new-phlan", true}, 24: {"new-phlan", true}, 27: {"new-phlan", true},
	17: {"moonsea", false}, 25: {"moonsea", false},
	18: {"limbo", false},
	19: {"zhentil-keep", true}, 49: {"zhentil-keep", true}, 48: {"manshoons-tower", false},
	20: {"dragonspear", true},
	21: {"dragonhorn-steading", false},
	22: {"kimarrs-fortress", false},
	26: {"thentia", true}, 36: {"arcams-caves", false},
	32: {"dragon-caves", false}, 33: {"dragon-caves", false}, 34: {"dragon-caves", false},
	35: {"drow-caves", false}, 41: {"drow-caves", false},
	37: {"giant-mines", false},
	38: {"pool-of-darkness", false}, 39: {"pool-of-darkness", false}, 40: {"pool-of-darkness", false}, 42: {"pool-of-darkness", false},
	50: {"temple-of-tyr", false}, 54: {"temple-of-tyr", false},
	51: {"moander", false}, 52: {"moander", false}, 53: {"moander", false}, 55: {"moander", false},
	64: {"iron-wall", false},
	65: {"gothmenes-palace", false}, 73: {"gothmenes-palace", false}, 74: {"gothmenes-palace", false},
	66: {"myth-drannor", false},
	67: {"red-tower", false}, 70: {"red-tower", false}, 75: {"red-tower", false}, 76: {"red-tower", false}, 77: {"red-tower", false},
	68: {"drow-fortress", false}, 69: {"drow-fortress", false}, 71: {"drow-fortress", false}, 72: {"drow-fortress", false},
	80: {"kalistes-web", false}, 81: {"kalistes-web", false}, 82: {"kalistes-web", false}, 83: {"kalistes-web", false},
	84: {"kalistes-web", false}, 85: {"kalistes-web", false}, 86: {"kalistes-web", false},
}
