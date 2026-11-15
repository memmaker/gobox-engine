package rules1e

import "gobox/rules"

// The named items: an item of a kind many different things share (a
// valuable, the potions' type with its mirrors and holy symbols, a curio,
// misc magic) has an id of its own by its name's words, so it has its own
// icon (items.rec) once it is identified. Numbered once from 1001, by name
// across the games (the same thing has the same id in each): a number is
// never reused or changed; a new one is added at the end. The words are
// each game's own word numbers (rules.json itemWords).

// NamedItem is the id of the item whose name is Words.
type NamedItem struct {
	Words [3]WordID    `json:"words"`
	ID    rules.ItemID `json:"id"`
}

// NamedItems are a game's named items by their words (rules.json
// namedItems; nil with POR's own words: PorNamedItems).
type NamedItems map[[3]WordID]rules.ItemID

// namedTable is l as the table the rules look an item up in.
func namedTable(l []NamedItem) NamedItems {
	m := NamedItems{}
	for _, n := range l {
		m[n.Words] = n.ID
	}
	return m
}

// PorNamedItems are POR's named items.
var PorNamedItems = []NamedItem{
	{[3]WordID{98, 156, 218}, 1001},  // Brass Dragon Figurine
	{[3]WordID{0, 118, 218}, 1002},   // Brass Mirror
	{[3]WordID{217, 167, 197}, 1003}, // Chain of Bone
	{[3]WordID{0, 80, 176}, 1004},    // Copper Brazier
	{[3]WordID{0, 119, 245}, 1005},   // Diamond Necklace
	{[3]WordID{194, 167, 96}, 1006},  // Dust of Disappearance
	{[3]WordID{0, 93, 178}, 1007},    // Electrum Decanter
	{[3]WordID{0, 199, 246}, 1010},   // Emerald Broach
	{[3]WordID{198, 247, 215}, 1012}, // Fine Opal Pendant
	{[3]WordID{0, 174, 215}, 1013},   // Fine Tapestry
	{[3]WordID{174, 231, 196}, 1014}, // Fungus Covered Tapestry
	{[3]WordID{69, 57, 179}, 1015},   // Gold Banded Wand
	{[3]WordID{0, 197, 179}, 1016},   // Gold Chain
	{[3]WordID{98, 157, 179}, 1017},  // Gold Lightning Figurine
	{[3]WordID{0, 132, 179}, 1018},   // Gold Scarab
	{[3]WordID{0, 195, 179}, 1019},   // Gold Statuette
	{[3]WordID{0, 174, 216}, 1020},   // Huge Tapestry
	{[3]WordID{0, 0, 106}, 1021},     // Incense
	{[3]WordID{250, 152, 148}, 1026}, // Iron Holy Symbol Of Tempus
	{[3]WordID{78, 52, 110}, 1027},   // Jewel Studded Bowl
	{[3]WordID{195, 156, 190}, 1028}, // Jewelled Dragon Statuette
	{[3]WordID{141, 177, 190}, 1029}, // Jewelled Silver Wings
	{[3]WordID{175, 167, 114}, 1030}, // Manual of Bodily Health
	{[3]WordID{0, 119, 122}, 1032},   // Pearl Necklace
	{[3]WordID{0, 134, 180}, 1034},   // Platinum Sphere
	{[3]WordID{185, 187, 64}, 1035},  // Potion Extra Healing
	{[3]WordID{239, 167, 64}, 1036},  // Potion of Giant Strength
	{[3]WordID{185, 167, 64}, 1037},  // Potion of Healing
	{[3]WordID{173, 167, 64}, 1039},  // Potion of Speed
	{[3]WordID{158, 50, 230}, 1040},  // Rotting Leather Saddle
	{[3]WordID{0, 130, 230}, 1041},   // Rotting Rug
	{[3]WordID{179, 167, 183}, 1042}, // Sheet of Gold
	{[3]WordID{0, 81, 177}, 1043},    // Silver Brooch
	{[3]WordID{195, 156, 177}, 1044}, // Silver Dragon Statuette
	{[3]WordID{251, 152, 177}, 1045}, // Silver Holy Symbol Of Sune
	{[3]WordID{0, 118, 177}, 1046},   // Silver Mirror
	{[3]WordID{0, 66, 177}, 1048},    // Silver Ring
	{[3]WordID{0, 132, 177}, 1049},   // Silver Scarab
	{[3]WordID{0, 195, 107}, 1050},   // Stone Statuette
	{[3]WordID{0, 0, 174}, 1052},     // Tapestry
	{[3]WordID{249, 152, 252}, 1053}, // Wooden Holy Symbol Of Tyr
	{[3]WordID{0, 152, 60}, 1054},    // Woods Holy Symbol
}

// CotabNamedItems are COTAB's named items.
var CotabNamedItems = []NamedItem{
	{[3]WordID{194, 167, 96}, 1006},  // Dust of Disappearance
	{[3]WordID{0, 116, 108}, 1022},   // Ioun Stone Deep Red
	{[3]WordID{0, 112, 108}, 1023},   // Ioun Stone Pale Blue
	{[3]WordID{120, 117, 108}, 1024}, // Ioun Stone Pink And Green
	{[3]WordID{121, 113, 108}, 1025}, // Ioun Stone Scarlet And Blue
	{[3]WordID{228, 167, 119}, 1031}, // Necklace of Missiles
	{[3]WordID{185, 187, 64}, 1035},  // Potion Extra Healing
	{[3]WordID{227, 167, 64}, 1038},  // Potion of Invisibility
	{[3]WordID{173, 167, 64}, 1039},  // Potion of Speed
	{[3]WordID{0, 118, 177}, 1046},   // Silver Mirror
}

// SotsbNamedItems are SOTSB's named items.
var SotsbNamedItems = []NamedItem{
	{[3]WordID{54, 90, 52}, 1008},  // Elixer of Youth
	{[3]WordID{96, 90, 59}, 1011},  // Eyes of Charming
	{[3]WordID{99, 90, 82}, 1031},  // Necklace of Missiles
	{[3]WordID{74, 90, 73}, 1033},  // Periapt of Health
	{[3]WordID{95, 94, 53}, 1035},  // Potion Extra Healing
	{[3]WordID{95, 90, 53}, 1037},  // Potion of Healing
	{[3]WordID{104, 90, 53}, 1038}, // Potion of Invisibility
	{[3]WordID{92, 90, 53}, 1039},  // Potion of Speed
	{[3]WordID{0, 82, 93}, 1047},   // Silver Necklace
	{[3]WordID{69, 90, 68}, 1051},  // Stone of Good Luck
}

// PodNamedItems are POD's named items.
var PodNamedItems = []NamedItem{
	{[3]WordID{54, 90, 52}, 1009},  // Elixir of Youth
	{[3]WordID{96, 90, 59}, 1011},  // Eyes of Charming
	{[3]WordID{99, 90, 82}, 1031},  // Necklace of Missiles
	{[3]WordID{74, 90, 73}, 1033},  // Periapt of Health
	{[3]WordID{95, 94, 53}, 1035},  // Potion Extra Healing
	{[3]WordID{95, 90, 53}, 1037},  // Potion of Healing
	{[3]WordID{104, 90, 53}, 1038}, // Potion of Invisibility
	{[3]WordID{92, 90, 53}, 1039},  // Potion of Speed
	{[3]WordID{69, 90, 68}, 1051},  // Stone of Good Luck
}
