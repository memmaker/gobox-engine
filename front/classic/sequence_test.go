package classic

import (
	"os"
	"path/filepath"
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// seqRules is a game that keeps its actions.
type seqRules = fake

// seqHost runs the presenter's calls on the classic game.
type seqHost struct {
	front.Host
	g front.Game
}

func (h seqHost) Call(f func(front.Game)) { f(h.g) }

// A skip ends State.Sequence with SequenceDone, once, even with no files to play.
func TestSequenceSkip(t *testing.T) {
	r := &fake{st: play.Scene{Kind: play.SceneSequence, Sequence: "INTRO", Common: allow(play.Continue{})}}
	g := NewGame(r, game.Data{Sequences: []game.Sequence{{Name: "INTRO", Anim: "INTRO"}}, Files: t.TempDir()})
	sq := Sequence(front.View{}, shownOf(g.r).Sequence, g.D.Sequences, g.D.Files).Sequence
	if sq == nil || sq.Anim != "INTRO" {
		t.Fatalf("View.Sequence %+v", sq)
	}
	var ui front.SequenceUI
	h := seqHost{g: g}
	front.SequenceStep(h, &ui, sq, true)
	front.SequenceStep(h, &ui, sq, true)
	if len(r.sent) != 1 || r.sent[0] != (play.Continue{}) {
		t.Fatalf("actions %v, want one Continue", r.sent)
	}
}

// With the real INTRO playing, a frame shows and nothing is sent until the skip.
func TestSequenceSkipPlaying(t *testing.T) {
	dir := filepath.Join(os.Getenv("HOME"), "Projects/pod-go/data/dos")
	if _, err := os.Stat(filepath.Join(dir, "INTRO.BMA")); err != nil {
		t.Skip("no Pools of Darkness files")
	}
	r := &fake{st: play.Scene{Kind: play.SceneSequence, Sequence: "INTRO", Common: allow(play.Continue{})}}
	g := NewGame(r, game.Data{Sequences: []game.Sequence{{Name: "INTRO", Anim: "INTRO"}}, Files: dir})
	sq := Sequence(front.View{}, shownOf(g.r).Sequence, g.D.Sequences, g.D.Files).Sequence
	var ui front.SequenceUI
	h := seqHost{g: g}
	if f := front.SequenceStep(h, &ui, sq, false); f == nil || len(r.sent) != 0 {
		t.Fatalf("first step: frame %v, actions %v", f, r.sent)
	}
	front.SequenceStep(h, &ui, sq, true)
	if len(r.sent) != 1 || r.sent[0] != (play.Continue{}) {
		t.Fatalf("actions %v, want one Continue", r.sent)
	}
}
