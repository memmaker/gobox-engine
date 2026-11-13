package ui

import (
	"image"
	"testing"

	"gobox/game"
)

// TestAreaPicFollowsArt: the overland's picture is made again when the art
// changes under the same terrain (F12 kept the other art's squares).
func TestAreaPicFollowsArt(t *testing.T) {
	const side = 2
	square := func(c byte) map[string]game.Picture {
		px := make([]byte, side*side)
		for i := range px {
			px[i] = c
		}
		return map[string]game.Picture{"terrain/0": {W: side, H: side, Pixels: px, Clear: -1}}
	}
	terrain, size := []int{0}, image.Pt(1, 1)
	var a areaPic
	a.resolve(terrain, size, square(1))
	dos := a.pics[0][0]
	a.resolve(terrain, size, square(2))
	if a.pics[0][0] == dos {
		t.Error("the other art left the overland's squares as they were")
	}
}
