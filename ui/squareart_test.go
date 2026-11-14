package ui

import (
	"testing"

	"gobox/game"
	"gobox/rules"
)

// TestSquareArt: a square with its own Art draws it (a tile slot of its
// own), whatever the named tiles; two squares with one picture share it.
func TestSquareArt(t *testing.T) {
	art := game.Picture{File: "WILDCOM", Block: 1, Frame: 3}
	c := &game.Combat{Tiles: []string{"", "x"}, Squares: []game.Ground{{At: rules.BoardPos{X: 1, Y: 1}, Art: art}, {At: rules.BoardPos{X: 2, Y: 1}, Art: art}, {At: rules.BoardPos{X: 3, Y: 1}, Tile: 1}}}
	b := boardIn(c, nil)
	g := b.Ground[1]
	if g[1].Tile != 2 || g[2].Tile != 2 || g[3].Tile != 1 {
		t.Fatalf("tiles %d %d %d", g[1].Tile, g[2].Tile, g[3].Tile)
	}
}
