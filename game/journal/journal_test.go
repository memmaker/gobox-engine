package journal

import (
	"reflect"
	"regexp"
	"testing"
	"testing/fstest"
)

var entry = []Ref{
	{regexp.MustCompile(`\bENTRY\s*(\d+)`), func(m []string) []string { return []string{"Journal Entry " + m[1]} }},
	{regexp.MustCompile(`\bTAVERN TALE\s*(\d+)`), func(m []string) []string { return []string{"Tavern Tale " + m[1]} }},
}

func TestRefsAndLayout(t *testing.T) {
	if got := Refs(entry, "TAVERN TALE 5, THEN SEE ENTRY 37."); !reflect.DeepEqual(got, []string{"Tavern Tale 5", "Journal Entry 37"}) {
		t.Fatalf("refs %q", got)
	}
	var p Printed
	if got := p.Add(entry, "YOU RECORD IT IN JOURNAL ", true); got != nil {
		t.Fatalf("partial %q", got)
	}
	if got := p.Add(entry, "ENTRY 37.", false); !reflect.DeepEqual(got, []string{"Journal Entry 37"}) {
		t.Fatalf("printed %q", got)
	}
	if got := p.Add(entry, " MORE", false); got != nil {
		t.Fatalf("again %q", got)
	}
	b, err := Load(fstest.MapFS{"J.TXT": {Data: []byte("# c\n[Journal Entry 37]\nA short line.\n\nA rather long line that wraps.\n")}}, "J.TXT")
	if err != nil || !b.Has("Journal Entry 37") || len(b.Titles()) != 1 || b.Text("Journal Entry 37")[0] != "A short line." {
		t.Fatal(err)
	}
	want := []string{"\x00Journal Entry 37", "", "A short line.", "", "A rather long", "line that wraps.", ""}
	if got := b.Layout([]string{"Journal Entry 37"}, 16, 20, nil); !reflect.DeepEqual(got, want) {
		t.Fatalf("layout %q", got)
	}
}
