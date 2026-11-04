package classic

import (
	"os"
	"path/filepath"
	"testing"

	"gobox/front"
	"gobox/game"
)

// seqRules is a game showing the sequence INTRO that keeps its decisions.
type seqRules struct{ got []game.Decision }

func (r *seqRules) State() game.State                 { return game.State{Sequence: "INTRO"} }
func (r *seqRules) Apply(d game.Decision) game.Result { r.got = append(r.got, d); return game.Result{} }
func (r *seqRules) Pending() *game.Ask                { return nil }
func (r *seqRules) Step() game.Result                 { return game.Result{} }

// seqHost runs the presenter's calls on the classic game.
type seqHost struct {
	front.Host
	g front.Game
}

func (h seqHost) Call(f func(front.Game)) { f(h.g) }

// A skip ends State.Sequence with SequenceDone, once, even with no files to play.
func TestSequenceSkip(t *testing.T) {
	r := &seqRules{}
	g := NewGame(r, game.Data{Sequences: []game.Sequence{{Name: "INTRO", Anim: "INTRO"}}, Files: t.TempDir()})
	sq := Sequence(front.View{}, r.State(), g.D.Sequences, g.D.Files).Sequence
	if sq == nil || sq.Anim != "INTRO" {
		t.Fatalf("View.Sequence %+v", sq)
	}
	var ui front.SequenceUI
	h := seqHost{g: g}
	front.SequenceStep(h, &ui, sq, true)
	front.SequenceStep(h, &ui, sq, true)
	if len(r.got) != 1 || r.got[0] != (game.SequenceDone{Name: "INTRO"}) {
		t.Fatalf("decisions %v, want one SequenceDone{INTRO}", r.got)
	}
}

// With the real INTRO playing, a frame shows and nothing is sent until the skip.
func TestSequenceSkipPlaying(t *testing.T) {
	dir := filepath.Join(os.Getenv("HOME"), "Projects/pod-go/data/dos")
	if _, err := os.Stat(filepath.Join(dir, "INTRO.BMA")); err != nil {
		t.Skip("no Pools of Darkness files")
	}
	r := &seqRules{}
	g := NewGame(r, game.Data{Sequences: []game.Sequence{{Name: "INTRO", Anim: "INTRO"}}, Files: dir})
	sq := Sequence(front.View{}, r.State(), g.D.Sequences, g.D.Files).Sequence
	var ui front.SequenceUI
	h := seqHost{g: g}
	if f := front.SequenceStep(h, &ui, sq, false); f == nil || len(r.got) != 0 {
		t.Fatalf("first step: frame %v, decisions %v", f, r.got)
	}
	front.SequenceStep(h, &ui, sq, true)
	if len(r.got) != 1 || r.got[0] != (game.SequenceDone{Name: "INTRO"}) {
		t.Fatalf("decisions %v, want one SequenceDone{INTRO}", r.got)
	}
}
