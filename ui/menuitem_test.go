package ui

import "testing"

// A game.MenuItem's line plays its sequence in gobox, and its end never reaches the game.
func TestMenuItemPlaysSequence(t *testing.T) {
	old := menuEntries
	defer func() { menuEntries, playSeq = old, "" }()
	menuEntries = []Entry{{Text: "Play intro", Key: 'I', Seq: "intro"}}
	rows := (&mainUI{}).entries(nil, nil, false)
	if len(rows) != 1 || rows[0].Text != "Play intro" {
		t.Fatalf("rows %v", rows)
	}
	rows[0].Do()
	if playSeq != "intro" {
		t.Fatalf("playSeq %q", playSeq)
	}
	classicGame{}.SequenceDone("intro") // a nil Game: it would panic if forwarded
	if playSeq != "" {
		t.Fatal("still playing")
	}
}
