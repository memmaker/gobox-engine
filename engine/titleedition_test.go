package engine

import (
	"slices"
	"testing"

	"gobox/adventure"
	"gobox/importer"
)

// TestTitleFollowsArt: the title's Amiga-only step (sotsb's credits card,
// TITLE 5) plays only while the Amiga art shows, also after F12 at the
// title: before, the steps were fixed at SetStart and the DOS title played
// the card, a picture the DOS art lacks (the SSI page stayed with the
// credit strip).
func TestTitleFollowsArt(t *testing.T) {
	t.Setenv(advEnvSOTSB, "../data/sotsb")
	e := advEngine(t, advEnvSOTSB)
	if !e.amigaUp() {
		t.Skip("no Amiga art imported")
	}
	f, err := importer.Read("../data/sotsb")
	if err != nil {
		t.Fatal(err)
	}
	card := slices.IndexFunc(f.Manifest.Start.Steps, func(s adventure.SequenceStep) bool { return s.Edition != 0 }) + 1
	e.Open("")
	shown := func() []int { return e.startScene().Intro.Pictures }
	if !slices.Contains(shown(), card) {
		t.Fatalf("Amiga art: step %d not in %v", card, shown())
	}
	DOSArt.Store(true)
	t.Cleanup(func() { DOSArt.Store(false) })
	if slices.Contains(shown(), card) {
		t.Fatalf("DOS art: the Amiga card (step %d) still in %v", card, shown())
	}
}
