package ui

import (
	"reflect"
	"testing"
	"testing/fstest"

	"gobox/game"
)

func TestJournalLayout(t *testing.T) {
	b, err := game.LoadJournal(fstest.MapFS{"J.TXT": {Data: []byte("# c\n[Journal Entry 37]\nA short line.\n\nA rather long line that wraps around the page because it is longer than the text width.\n")}}, "J.TXT")
	if err != nil || !b.Has("Journal Entry 37") || len(b.Titles()) != 1 || b.Text("Journal Entry 37")[0] != "A short line." {
		t.Fatal(err)
	}
	want := []string{"\x00Journal Entry 37", "", "A short line.", "", "A rather long line that wraps around", "the page because it is longer than", "the text width.", ""}
	if got := (pageCut{cols: 40, rows: 25}).layout(b.Texts([]string{"Journal Entry 37"}, nil), nil); !reflect.DeepEqual(got, want) {
		t.Fatalf("layout %q", got)
	}
}
