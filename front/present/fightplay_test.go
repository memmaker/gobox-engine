package present

import (
	"testing"

	"gobox/front"
)

func TestPlayMakesSnapshotTiles(t *testing.T) {
	var p Presenter
	pc := front.Picture{Pix: []byte{1, 2, 3, 4}, PixW: 2}
	p.Play([]front.CombatEvent{{Ms: 500, Battle: &front.Battle{TilePix: []front.Picture{pc}}}})
	if b := p.replay.q[0].Battle; len(b.Tiles) != 1 || b.Tiles[0] == nil {
		t.Fatal("snapshot tiles not made")
	}
}
