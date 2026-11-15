package main

import (
	"os"
	"path/filepath"
	"testing"

	"gobox/game"
	"gobox/pool"
	"gobox/ui"
)

// TestShippedSlotsLoad (bug class: a slot format change refuses slots
// already on disk): every slot a game ships under its saves folder's saves
// loads: POR's own, sotsb's imported original save (pod and cotab ship none).
func TestShippedSlotsLoad(t *testing.T) {
	for id, saves := range map[string]string{"por": "sources/save", "sotsb": "saves"} {
		t.Run(id, func(t *testing.T) {
			adv := "../../data/" + id
			slots, err := os.ReadDir(filepath.Join(adv, saves, "saves"))
			if err != nil {
				t.Fatal(err)
			}
			eng, err := pool.New(adv, adv+"/sources")
			if err != nil {
				t.Fatal(err)
			}
			eng.Cfg.SaveDir = t.TempDir()
			if err := os.CopyFS(eng.Cfg.SaveDir, os.DirFS(filepath.Join(adv, saves))); err != nil {
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
