package importer

import (
	"slices"
	"testing"
)

// TestBRPlacesInMusicPlaces: every block's trigger is one of game.Data.MusicPlaces (gobox makes their folders), and
// every listed trigger is some block's.
func TestBRPlacesInMusicPlaces(t *testing.T) {
	used := map[string]bool{}
	blocks := map[int]bool{}
	for _, p := range brPlaces {
		used[p.Music] = true
		if blocks[p.Block] {
			t.Errorf("block %d twice", p.Block)
		}
		blocks[p.Block] = true
		if !slices.Contains(brMusicPlaces, p.Music) {
			t.Errorf("block %d: %q not in brMusicPlaces", p.Block, p.Music)
		}
	}
	for _, m := range brMusicPlaces {
		if !used[m] {
			t.Errorf("%q is no block's place", m)
		}
	}
}
