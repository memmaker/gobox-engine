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

// TestBR2PlacesInMusicPlaces is the same for BR2, whose "space" trigger is the rocket's (buckrogers2-go adv.go: mode 9 reports "space"),
// no block's.
func TestBR2PlacesInMusicPlaces(t *testing.T) {
	used := map[string]bool{"space": true}
	blocks := map[int]bool{}
	for _, p := range br2Places {
		used[p.Music] = true
		if blocks[p.Block] {
			t.Errorf("block %d twice", p.Block)
		}
		blocks[p.Block] = true
		if !slices.Contains(br2MusicPlaces, p.Music) {
			t.Errorf("block %d: %q not in br2MusicPlaces", p.Block, p.Music)
		}
	}
	for _, m := range br2MusicPlaces {
		if !used[m] {
			t.Errorf("%q is no block's place", m)
		}
	}
}
