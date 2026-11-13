package importer

import "slices"

// musicPlace is an area's own music trigger (game.Where.Music, gobox
// data/music/<slug>) and the kind of place it is (Where.Mode).
type sotsbMusicPlace struct{ slug, mode, name string }

// musicPlaces are the areas by geo block (GEO1..3; the floors of one
// dungeon share a trigger).
var sotsbMusicPlaces = map[int]sotsbMusicPlace{
	16: {"new-verdigris", "town", "New Verdigris"},
	32: {"old-verdigris", "dungeon", "Old Verdigris"},
	33: {"well-of-knowledge", "dungeon", "Well of Knowledge"},
	34: {"black-circle-hq", "dungeon", "Black Circle Headquarters"},
	48: {"mines", "dungeon", "Mines 1"}, 49: {"mines", "dungeon", "Mines 2"},
	50: {"temple-of-tyr", "dungeon", "Temple of Tyr"},
	64: {"the-dungeon", "dungeon", "The Dungeon 1"}, 65: {"the-dungeon", "dungeon", "The Dungeon 2"}, 66: {"the-dungeon", "dungeon", "The Dungeon 3"},
	68: {"drider-base", "dungeon", "Drider Base"},
	80: {"crevasses", "wilderness", "Crevasses 1"}, 82: {"crevasses", "wilderness", "Crevasses 2"},
	81: {"frost-giant-village", "wilderness", "Frost Giant Village"},
	96: {"dreadlords-castle", "dungeon", "Dreadlord's Castle 1"}, 97: {"dreadlords-castle", "dungeon", "Dreadlord's Castle 2"}, 98: {"dreadlords-castle", "dungeon", "Dreadlord's Castle 3"},
}

// MusicPlaces is every trigger musicPlaces sets (game.Data.MusicPlaces).
func SotsbMusicPlaces() []string {
	var l []string
	for _, p := range sotsbMusicPlaces {
		l = append(l, p.slug)
	}
	slices.Sort(l)
	return slices.Compact(l)
}
