package classic

import (
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// A game's song event reaches the presenter as front.EvSong with its ref as Amount.
func TestSongEvent(t *testing.T) {
	out := combatEvents([]play.Report{{Kind: game.EvSong, Sound: "2"}}, front.View{}, nil, 0)
	if len(out) != 1 || out[0].Kind != front.EvSong || out[0].Amount != 2 {
		t.Fatalf("got %+v", out)
	}
}
