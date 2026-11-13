package game

import (
	"reflect"
	"testing"
	"testing/fstest"
)

func TestLayout(t *testing.T) {
	b, err := LoadJournal(fstest.MapFS{"J.TXT": {Data: []byte("# c\n[Journal Entry 37]\nA short line.\n\nA rather long line that wraps.\n")}}, "J.TXT")
	if err != nil || !b.Has("Journal Entry 37") || len(b.Titles()) != 1 || b.Text("Journal Entry 37")[0] != "A short line." {
		t.Fatal(err)
	}
	want := []string{"\x00Journal Entry 37", "", "A short line.", "", "A rather long", "line that wraps.", ""}
	if got := b.Layout([]string{"Journal Entry 37"}, 16, 20, nil); !reflect.DeepEqual(got, want) {
		t.Fatalf("layout %q", got)
	}
}
