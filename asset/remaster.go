package asset

import (
	"embed"
	"io/fs"
)

// remasterFS holds each game's remastered art by game id (remaster/<id>/):
// the wall table assets/walls/walls.txt and the maps maps/NAME.map (the wall
// images are the shared walls).
//
//go:embed all:remaster
var remasterFS embed.FS

// remasterAreaMaps names, by game id, the remastered map of each GEO block (remaster_areas.go).
var remasterAreaMaps = map[string]map[int]string{"por": porAreaMaps}

// Remaster is game id's remastered art and the map name of each area id
// (nil, nil: the game has none).
func Remaster(id string) (fs.FS, map[int]string) {
	maps, ok := remasterAreaMaps[id]
	if !ok {
		return nil, nil
	}
	sub, err := fs.Sub(remasterFS, "remaster/"+id)
	if err != nil {
		panic(err) // a valid path in an embedded tree
	}
	return sub, maps
}
