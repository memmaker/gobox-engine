package ui

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// A game's song event reaches the presenter as EvSong with its ref as its sound number.
func TestSongEvent(t *testing.T) {
	out := played([]play.Report{{Kind: game.EvSong, Sound: "2"}}, false, nil, 0)
	if len(out) != 1 || out[0].Kind != game.EvSong || soundNo(out[0].Moment) != 2 {
		t.Fatalf("got %+v", out)
	}
}
