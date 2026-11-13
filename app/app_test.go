package app

import (
	"encoding/json"
	"gobox/ui"
	"os"
	"slices"
	"testing"
)

func TestParseKeys(t *testing.T) {
	got := ParseKeys(`a\r\M\C\P\w\x\`, map[byte]Key{'r': 13})
	want := []Key{'a', 13, Right, Left, Next, Wait, '\\', 'x', '\\'}
	if !slices.Equal(got, want) {
		t.Fatalf("%x, want %x", got, want)
	}
}

func TestPrefs(t *testing.T) {
	p := Prefs{Dir: t.TempDir() + "/sub"}
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
	m := Menu{Extra: &Extra{2}}
	got, _ := json.Marshal(m.both(ui.BlobberOptions{Torch: 1}))
	want, _ := json.Marshal(struct {
		ui.BlobberOptions
		Extra
	}{ui.BlobberOptions{Torch: 1}, Extra{2}})
	if string(got) != string(want) {
		t.Fatalf("%s\nwant %s", got, want)
	}
}
