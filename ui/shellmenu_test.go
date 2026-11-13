package ui

import (
	"encoding/json"
	"os"
	"testing"
)

func TestPrefs(t *testing.T) {
	p := prefs{Dir: t.TempDir() + "/sub"}
	type opts struct{ A, B int }
	o := opts{A: 1, B: 2}
	if p.Load(&o) {
		t.Fatal("no file loaded")
	}
	if err := p.Save(map[string]int{"A": 5}); err != nil {
		t.Fatal(err)
	}
	if !p.Load(&o) || o != (opts{5, 2}) {
		t.Fatalf("got %+v", o)
	}
	if b, _ := os.ReadFile(p.Path()); string(b) != "{\n \"A\": 5\n}" {
		t.Fatalf("file %q", b)
	}
}

// The Esc menu's file holds the game's options beside blobber's, at the top level, in order.
func TestMenuBoth(t *testing.T) {
	type Extra struct{ Zoom float64 }
	m := shellMenu{Extra: &Extra{2}}
	var o MenuOptions
	o.Torch = 1
	got, _ := json.Marshal(m.both(o))
	want, _ := json.Marshal(struct {
		MenuOptions
		Extra
	}{o, Extra{2}})
	if string(got) != string(want) {
		t.Fatalf("%s\nwant %s", got, want)
	}
}
