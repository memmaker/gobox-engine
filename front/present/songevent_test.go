package present

import (
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// A game's song event reaches the presenter as front.EvSong with its ref as Amount.
func TestSongEvent(t *testing.T) {
	out := playedEvents([]play.Report{{Kind: game.EvSong, Sound: "2"}}, false, nil, 0)
	if len(out) != 1 || out[0].Kind != front.EvSong || out[0].Amount != 2 {
		t.Fatalf("got %+v", out)
	}
}
