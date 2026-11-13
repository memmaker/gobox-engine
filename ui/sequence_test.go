package ui

import (
	"os"
	"path/filepath"
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// seqHost runs the presenter's calls on the classic game.
type seqHost struct {
	RuleHost
	g *classicGame
}

func (h seqHost) Call(f func(*classicGame)) { f(h.g) }

// sequenceIn is the sequence g's state plays, with g's data the game's.
func sequenceIn(t *testing.T, g *Game) *game.Sequence {
	s, d := sequences, filesDir
	t.Cleanup(func() { sequences, filesDir = s, d })
	sequences, filesDir = g.D.Sequences, g.D.AnimDir()
	return presenterOf(g).sequenceUp()
}

// A skip ends State.Sequence with SequenceDone, once, even with no files to play.
func TestSequenceSkip(t *testing.T) {
	r := &fake{st: play.Scene{Kind: play.SceneSequence, Sequence: "INTRO", Common: allow(play.Continue{})}}
	g := NewGame(r, game.Data{Sequences: []game.Sequence{{Name: "INTRO", Anim: "INTRO"}}, Files: t.TempDir()})
	sq := sequenceIn(t, g)
	if sq == nil || sq.Anim != "INTRO" {
		t.Fatalf("sequence %+v", sq)
	}
	var ui SequenceUI
	h := seqHost{g: &classicGame{Game: g}}
	SequenceStep(h, &ui, sq, true)
	SequenceStep(h, &ui, sq, true)
	if len(r.sent) != 1 || r.sent[0] != (play.Continue{}) {
		t.Fatalf("actions %v, want one Continue", r.sent)
	}
}

// With the real INTRO playing, a frame shows and nothing is sent until the skip.
func TestSequenceSkipPlaying(t *testing.T) {
	dir := filepath.Join("..", "data", "pod", "sources", "dos")
	if _, err := os.Stat(filepath.Join(dir, "INTRO.BMA")); err != nil {
		t.Skip("no Pools of Darkness files")
	}
	r := &fake{st: play.Scene{Kind: play.SceneSequence, Sequence: "INTRO", Common: allow(play.Continue{})}}
	g := NewGame(r, game.Data{Sequences: []game.Sequence{{Name: "INTRO", Anim: "INTRO"}}, Files: dir})
	sq := sequenceIn(t, g)
	var ui SequenceUI
	h := seqHost{g: &classicGame{Game: g}}
	if f := SequenceStep(h, &ui, sq, false); f == nil || len(r.sent) != 0 {
		t.Fatalf("first step: frame %v, actions %v", f, r.sent)
	}
	SequenceStep(h, &ui, sq, true)
	if len(r.sent) != 1 || r.sent[0] != (play.Continue{}) {
		t.Fatalf("actions %v, want one Continue", r.sent)
	}
}
