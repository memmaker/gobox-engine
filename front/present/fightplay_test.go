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

// Space during the computer's turn passes every queued timed moment; a
// moment waiting for a key still waits.
func TestSpaceFastForward(t *testing.T) {
	var p Presenter
	b := &front.Battle{}
	p.Play([]front.CombatEvent{{Ms: 5000, Battle: b}, {Ms: 5000, Battle: b}, {Ms: 5000, Battle: b}, {Key: true, Battle: b}})
	if !p.Playing() {
		t.Fatal("not playing")
	}
	p.Key(' ', front.View{})
	if ev := p.playing(); ev == nil || !ev.Key {
		t.Fatalf("after Space: %+v, want the key moment", ev)
	}
}

// A 5-frame blast (5 timed scenes of 60 ms) plays, and Space passes it.
func TestBlastPlaysAndSpacePasses(t *testing.T) {
	var p Presenter
	var evs []front.CombatEvent
	for i := 0; i < 5; i++ {
		evs = append(evs, front.CombatEvent{Kind: front.EvFrame, Ms: 60, Scene: &front.View{}})
	}
	p.Play(evs)
	if p.playing() == nil {
		t.Fatal("blast not shown")
	}
	p.Key(' ', front.View{})
	if ev := p.playing(); ev != nil || p.Playing() {
		t.Fatalf("after Space: %+v", ev)
	}
}
