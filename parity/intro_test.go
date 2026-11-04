package parity

import (
	"os"
	"path/filepath"
	"testing"

	"por/rules"
)

// TestIntroParity: State.Intro with the title waiting to be played, its
// credits page (State.Flags) and the party menu after it.
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
		at(t, "intro", e)
		if e.Rules().State().Intro != nil {
			seen++
		}
	}
	if seen == 0 {
		t.Error("no intro seen")
	}
}
