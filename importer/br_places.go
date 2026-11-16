package importer

import "gobox/adventure"

// brPlaces are BR1's blocks' music triggers (buckrogers-go rules/music.go musicPlace, by disk and block; the blocks are
// unique over the disks) and townPlaces ("salvation", "spaceport": the peaceful streets, Where.Mode "town"). Every trigger is
// in brMusicPlaces (game.Data.MusicPlaces).
var brPlaces = []adventure.Place{
	{Block: 16, Music: "neo-chicago"}, {Block: 17, Music: "salvation", Town: true},
	{Block: 32, Music: "derelict"}, {Block: 33, Music: "derelict"}, {Block: 34, Music: "space"}, {Block: 35, Music: "asteroid-mine"},
	{Block: 48, Music: "ceres-base"}, {Block: 49, Music: "ceres-base"}, {Block: 50, Music: "pirate-base"}, {Block: 52, Music: "ship-boarding"},
	{Block: 64, Music: "mars-surface"}, {Block: 65, Music: "mars-surface"}, {Block: 66, Music: "mars-gradivus"}, {Block: 67, Music: "mars-gradivus"},
	{Block: 80, Music: "venus-surface"}, {Block: 81, Music: "venus-surface"}, {Block: 82, Music: "venus-base"}, {Block: 83, Music: "venus-base"},
	{Block: 94, Music: "mercury"}, {Block: 95, Music: "spaceport", Town: true}, {Block: 96, Music: "mercury"}, {Block: 97, Music: "mercury"},
	{Block: 98, Music: "mariposa-doomsday"}, {Block: 99, Music: "spaceport", Town: true},
}
