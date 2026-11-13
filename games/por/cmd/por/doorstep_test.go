package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"gobox/games/por/playpor"
	"gobox/games/por/porgame"
	"gobox/ui"
)

// noTitle fails when a frame of f draws the title's picture: only the main
// menu (the party menu before play) shows it.
func noTitle(t *testing.T, f *ui.Front, n int) {
	t.Helper()
	for i := range n {
		f.Frame(1280, 800)
		f.Tick()
		if s := f.P.Shown(); f.TitleDrawn() {
			t.Fatalf("frame %d: the title's picture drawn (menu %v hall %v exploring %v door %v)", i, s.MainMenu, f.Game.Where().Hall, f.Game.Flags().Exploring, s.Door != nil)
		}
	}
}

// TestLockedDoorOverView: back-for-good's first step north is Podol
// Plaza's locked door. Its menu (ovr15:0D6D, PendDoor) is over the 3D
// view; it was ModeNone, so Assembling, and the front drew the main menu
// over the title's picture.
func TestLockedDoorOverView(t *testing.T) {
	if err := os.Chdir("../.."); err != nil {
		t.Skip(err)
	}
	t.Cleanup(func() { os.Chdir("cmd/por") })
	src := "../../data/por/sources/save/saves/back-for-good"
	if _, err := os.Stat(src); err != nil {
		t.Skip(err)
	}
	eng, err := porgame.New("../../data/por/sources/assets", "../../data/por/sources")
	if err != nil {
		t.Skip(err)
	}
	dir := t.TempDir() // a copy: the save is never written
	if err := os.MkdirAll(filepath.Join(dir, "saves"), 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("cp", "-R", src, filepath.Join(dir, "saves")).CombinedOutput(); err != nil {
		t.Fatal(err, string(out))
	}
	eng.Cfg.SaveDir = dir
	eng.Explore = true
	eng.FullIntro.Store(false)
	seed := uint32(1)
	eng.FixedSeed = &seed
	data := playpor.Data(eng)
	f := ui.NewFront(ui.NewGame(eng, data), data)
	f.Start("back-for-good")
	f.Frame(1280, 800)
	f.Game.SequenceDone("intro")
	f.Game.TitleDone()
	f.Game.Settle()
	for i := 0; i < 600 && !f.Game.Flags().Exploring; i++ {
		f.Frame(1280, 800)
		f.Tick()
	}
	if ex, area := f.Game.Flags().Exploring, f.Game.Where().Area; !ex || area != "18" {
		t.Fatalf("exploring %v area %s: the save did not load", ex, area)
	}
	noTitle(t, f, 10)
	f.Key(0x4800) // the first key opens the Front (Key: !started)
	f.Frame(1280, 800)
	f.Key(0x4800) // north: the door
	noTitle(t, f, 30)
	if ex, s := f.Game.Flags().Exploring, f.P.Shown(); !ex || s.Door == nil || s.MainMenu {
		t.Fatalf("exploring %v door %v menu %v: not the locked door over the 3D view", ex, s.Door != nil, s.MainMenu)
	}
}

// TestHallNoTitle: the training hall's party menu is over the 3D view,
// never the title's picture.
func TestHallNoTitle(t *testing.T) {
	if err := os.Chdir("../.."); err != nil {
		t.Skip(err)
	}
	t.Cleanup(func() { os.Chdir("cmd/por") })
	eng, err := porgame.New("../../data/por/sources/assets", "../../data/por/sources")
	if err != nil {
		t.Skip(err)
	}
	eng.Cfg.SaveDir = t.TempDir()
	eng.Explore = true
	eng.FullIntro.Store(false)
	eng.DevAction = eng.DevScene("hall")
	seed := uint32(1)
	eng.FixedSeed = &seed
	data := playpor.Data(eng)
	f := ui.NewFront(ui.NewGame(eng, data), data)
	f.Start("")
	f.Frame(1280, 800)
	f.Game.SequenceDone("intro")
	f.Game.TitleDone()
	f.Game.Settle()
	inHall := func() bool { return f.Game.Where().Hall && f.P.Shown().MainMenu } // the hall's party menu
	for i := 0; i < 600 && !inHall(); i++ {
		f.Frame(1280, 800)
		f.Tick()
	}
	if !inHall() {
		t.Fatal("never in the hall")
	}
	noTitle(t, f, 30)
}
