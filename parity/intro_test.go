package parity

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"gobox/front/classic"
	"por/rules"
)

// TestIntroParity: the title waiting to be played, its credits page
// (State.Flags, flags_test.go) and the party menu after it.
func TestIntroParity(t *testing.T) {
	dir := filepath.Join(porDir(t), "data")
	if _, err := os.Stat(filepath.Join(dir, "dos", "TITLE.DAX")); err != nil {
		t.Skip("no game files")
	}
	e, err := rules.New(dir)
	if err != nil {
		t.Fatal(err)
	}
	e.Cfg.SaveDir = t.TempDir()
	e.Start(e.Run)
	seen := 0
	for _, step := range []func(){func() {}, e.TitleCredits, e.TitleDone} {
		step()
		e.Apply()
		v, s := e.View(), e.Rules().State()
		b := v
		b.Intro, b.Credits = nil, false
		c := classic.Flags(classic.Intro(b, s), s)
		if !reflect.DeepEqual(c.Intro, v.Intro) || c.Credits != v.Credits {
			t.Errorf("intro %+v credits %v, want %+v %v", c.Intro, c.Credits, v.Intro, v.Credits)
		}
		if v.Intro != nil {
			seen++
		}
	}
	if seen == 0 {
		t.Error("no intro seen")
	}
}
