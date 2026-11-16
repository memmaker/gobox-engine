package importer

import "gobox/adventure"

// brPlaces are BR1's blocks' music triggers (buckrogers-go rules/music.go musicPlace, by disk and block; the blocks are
// unique over the disks) and townPlaces ("salvation", "spaceport": the peaceful streets, Where.Mode "town"). Every trigger is
// in brMusicPlaces (game.Data.MusicPlaces).
var brPlaces = named(brPlaceNames, []adventure.Place{
	{Block: 16, Music: "neo-chicago"}, {Block: 17, Music: "salvation", Town: true},
	{Block: 32, Music: "derelict"}, {Block: 33, Music: "derelict"}, {Block: 34, Music: "space"}, {Block: 35, Music: "asteroid-mine"},
	{Block: 48, Music: "ceres-base"}, {Block: 49, Music: "ceres-base"}, {Block: 50, Music: "pirate-base"}, {Block: 52, Music: "ship-boarding"},
	{Block: 64, Music: "mars-surface"}, {Block: 65, Music: "mars-surface"}, {Block: 66, Music: "mars-gradivus"}, {Block: 67, Music: "mars-gradivus"},
	{Block: 80, Music: "venus-surface"}, {Block: 81, Music: "venus-surface"}, {Block: 82, Music: "venus-base"}, {Block: 83, Music: "venus-base"},
	{Block: 94, Music: "mercury"}, {Block: 95, Music: "spaceport", Town: true}, {Block: 96, Music: "mercury"}, {Block: 97, Music: "mercury"},
	{Block: 98, Music: "mariposa-doomsday"}, {Block: 99, Music: "spaceport", Town: true},
})

// brPlaceNames is each BR1 music place's name as the game's texts call it, the map's area title (buckrogers-go
// rules/music.go placeName).
var brPlaceNames = map[string]string{"neo-chicago": "Neo-Chicago", "salvation": "Salvation", "derelict": "Derelict Ship",
	"space": "Space", "asteroid-mine": "Asteroid Mine", "ceres-base": "Ceres Base", "pirate-base": "Pirate Base",
	"ship-boarding": "Boarded Ship", "mars-surface": "Mars", "mars-gradivus": "Gradivus", "venus-surface": "Venus",
	"venus-base": "Venus Base", "mercury": "Mercury", "spaceport": "Mercury Spaceport", "mariposa-doomsday": "Mariposa"}

// named is ps with each place's Name from names by its music trigger.
func named(names map[string]string, ps []adventure.Place) []adventure.Place {
	for i := range ps {
		ps[i].Name = names[ps[i].Music]
	}
	return ps
}

// br2Places are BR2's blocks' music triggers (buckrogers2-go rules/data.go blockMusic by block; townBlocks 17, 18 and 32 are the
// towns, Where.Mode "town"). The shared blocks (clinic, bank, intro, debug) have none.
var br2Places = []adventure.Place{
	{Block: 17, Music: "caloris", Town: true}, {Block: 18, Music: "salvation", Town: true}, {Block: 21, Music: "ship-boarding"},
	{Block: 32, Music: "losangelorg", Town: true}, {Block: 33, Music: "losangelorg"}, {Block: 34, Music: "losangelorg"},
	{Block: 36, Music: "losangelorg"}, {Block: 37, Music: "losangelorg"}, {Block: 38, Music: "losangelorg"},
	{Block: 35, Music: "ceres-base"},
	{Block: 48, Music: "venus"}, {Block: 49, Music: "venus"}, {Block: 50, Music: "venus"},
	{Block: 64, Music: "luna"}, {Block: 65, Music: "luna"}, {Block: 66, Music: "maelstrom-rider"},
	{Block: 80, Music: "prison"}, {Block: 81, Music: "purge-hq"}, {Block: 84, Music: "purge-hq"}, {Block: 82, Music: "research-base"},
	{Block: 96, Music: "living-ship"}, {Block: 98, Music: "living-ship"}, {Block: 97, Music: "fungus-asteroid"},
	{Block: 112, Music: "floating-school"}, {Block: 113, Music: "genetics-foundation"}, {Block: 114, Music: "mining-rig"},
}
