package ui

import (
	"gobox/adventure"
	"os"
	"path/filepath"
	"testing"

	"gobox/game"
)

// seqHost runs the presenter's calls on the classic game.
type seqHost struct {
	RuleHost
	g *classicGame
}

func (h seqHost) Call(f func(*classicGame)) { f(h.g) }

// withScripts makes start and end the adventure's scripts for the test.
func withScripts(t *testing.T, start, end *adventure.Sequence) {
	s, e := startSeq, endSeq
	t.Cleanup(func() { startSeq, endSeq = s, e })
	SetScripts(start, end)
}

// introAnim is a Start that plays the INTRO animation first.
var introAnim = &adventure.Sequence{Steps: []adventure.SequenceStep{{Kind: adventure.StepAnim, Anim: "INTRO"}}}

// sequenceIn is the sequence g's state plays, with g's data the game's.
func sequenceIn(t *testing.T, g *Game) string {
	d := filesDir
	t.Cleanup(func() { filesDir = d })
	filesDir = g.D.AnimDir()
	withScripts(t, introAnim, nil)
	return presenterOf(g).sequenceUp()
}

// A skip ends State.Sequence with SequenceDone, once, even with no files to play.
func TestSequenceSkip(t *testing.T) {
	r := &fake{st: game.Scene{Kind: game.SceneSequence, Sequence: adventure.SeqIntro, Common: allow(game.ContinueAction{})}}
	g := NewGame(r, game.Data{Files: t.TempDir()})
	sq := sequenceIn(t, g)
	if sq != adventure.SeqIntro || seqAnim(sq).Anim != "INTRO" {
		t.Fatalf("sequence %q", sq)
	}
	var ui SequenceUI
	h := seqHost{g: &classicGame{Game: g}}
	SequenceStep(h, &ui, sq, true)
	SequenceStep(h, &ui, sq, true)
	if len(r.sent) != 1 || r.sent[0] != (game.ContinueAction{}) {
		t.Fatalf("actions %v, want one Continue", r.sent)
	}
}

// With the real INTRO playing, a frame shows and nothing is sent until the skip.
func TestSequenceSkipPlaying(t *testing.T) {
	dir := filepath.Join("..", "data", "pod", "sources", "dos")
	if _, err := os.Stat(filepath.Join(dir, "INTRO.BMA")); err != nil {
		t.Skip("no Pools of Darkness files")
	}
	r := &fake{st: game.Scene{Kind: game.SceneSequence, Sequence: adventure.SeqIntro, Common: allow(game.ContinueAction{})}}
	g := NewGame(r, game.Data{Files: dir})
	sq := sequenceIn(t, g)
	var ui SequenceUI
	h := seqHost{g: &classicGame{Game: g}}
	if f := SequenceStep(h, &ui, sq, false); f == nil || len(r.sent) != 0 {
		t.Fatalf("first step: frame %v, actions %v", f, r.sent)
	}
	SequenceStep(h, &ui, sq, true)
	if len(r.sent) != 1 || r.sent[0] != (game.ContinueAction{}) {
		t.Fatalf("actions %v, want one Continue", r.sent)
	}
}
