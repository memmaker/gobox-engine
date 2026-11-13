package main

import (
	"os"
	"path/filepath"
	"testing"

	"gobox/engine"
	"gobox/game"
	"gobox/ui"
)

// TestShippedSlotsLoad (bug class: a slot format change refuses slots
// already on disk): every slot a game ships under sources/save/saves loads.
func TestShippedSlotsLoad(t *testing.T) {
	for _, id := range []string{"por", "pod", "cotab", "sotsb"} {
		t.Run(id, func(t *testing.T) {
			src := "../../data/" + id + "/sources"
			slots, err := os.ReadDir(filepath.Join(src, "save", "saves"))
			if err != nil {
				t.Skip(err)
			}
			eng, err := engine.NewPool(src+"/assets", src)
			if err != nil {
				t.Skip(err)
			}
			eng.Cfg.SaveDir = t.TempDir()
			if err := os.CopyFS(eng.Cfg.SaveDir, os.DirFS(filepath.Join(src, "save"))); err != nil {
				t.Fatal(err)
			}
			g := ui.NewGame(eng, eng.Data())
			g.Open("")
			for i := 0; g.Mode() == game.ModeTitle && i < 5; i++ {
				g.TitleDone()
			}
			for _, s := range slots {
				if s.IsDir() && !eng.LoadNamed(s.Name(), true) {
					t.Errorf("slot %q refused", s.Name())
				}
			}
		})
	}
}
