package games

// The monsters every fight adds by area screen (DS:08D8 monster, DS:08DD
// icon), ten of them, standing aside (status 9, hired-NPC control).
var (
	areaMonsters     = [...]byte{59, 6, 2, 8, 59}
	areaMonsterIcons = [...]byte{59, 6, 2, 8, 59}
)

const (
	areaMonsterCount   = 10
	areaMonsterStatus  = 9
	areaMonsterControl = 50
)

// The random scroll of O055:1ABE.
const (
	scrollType   = 39
	scrollPlus   = 1
	scrollWeight = 1
	scrollPrice  = 3000
	scrollSpells = 3
	spellLast    = 126
)

var scrollWords = [3]byte{102, 39, 40}

// SOUND's operand: from 128 on a MUSIC block, 2..14 a digitized sound.
const (
	soundMusicFlag    = 128
	soundDigitizedMin = 2
	soundDigitizedMax = 14
)
