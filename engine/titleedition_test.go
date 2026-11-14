package engine

import (
	"slices"
	"testing"

	"gobox/adventure"
	"gobox/importer"
)

// TestTitleFollowsArt: the title the engine publishes holds every edition's
// steps whatever art shows (sotsb's Amiga credits card, TITLE 5), so F12
// needs no new State; the step plays only while the Amiga art shows
// (Sequence.Plays): before, the DOS title played the card, a picture the
// DOS art lacks (the SSI page stayed with the credit strip).
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
	amiga := e.startScene().Intro.Pictures
	DOSArt.Store(true)
	t.Cleanup(func() { DOSArt.Store(false) })
	if dos := e.startScene().Intro.Pictures; !slices.Contains(dos, card) || !slices.Equal(dos, amiga) {
		t.Fatalf("the published steps follow the art: Amiga %v, DOS %v (card %d)", amiga, dos, card)
	}
	if !f.Manifest.Start.Plays(card, true) || f.Manifest.Start.Plays(card, false) {
		t.Fatalf("the card (step %d) plays with the Amiga art only", card)
	}
}
