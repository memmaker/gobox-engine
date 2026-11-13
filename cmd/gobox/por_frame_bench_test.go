package main

import (
	"gobox/engine"
	"os"
	"testing"

	"gobox/game"
	"gobox/ui"
)

// frameFront is por as main starts it, headless, past the title: -explore
// (a Quick Start party in the 3D view) or a dev scene (autocombat: a fight).
func frameFront(b *testing.B, dev string) *ui.Front {
	if err := os.Chdir(".."); err != nil {
		b.Skip(err)
	}
	b.Cleanup(func() { os.Chdir("playtest") })
	eng, err := engine.NewPool("../../data/por", "../../data/por/sources")
	if err != nil {
		b.Skip(err)
	}
	eng.Cfg.SaveDir = b.TempDir()
	eng.Explore = true
	eng.FullIntro.Store(false)
	if dev != "" {
		eng.DevAction = eng.DevScene(dev)
	}
	seed := uint32(1)
	eng.FixedSeed = &seed
	data := eng.Data()
	f := ui.NewFront(ui.NewGame(eng, data), data)
	f.Start("")
	f.Frame(1280, 800)
	f.Game.SequenceDone("intro")
	f.Game.TitleDone()
	f.Game.Settle()
	return f
}

// A frame of walking the 3D view: a step or a turn every 60 frames. A
// regression in what runs per frame (the State built again,
// uploads) shows here; the window's GPU time does not.
func BenchmarkExploreFrame(b *testing.B) {
	f := frameFront(b, "")
	for i := 0; i < 600 && !f.Game.Flags().Exploring; i++ {
		f.Frame(1280, 800)
	}
	if !f.Game.Flags().Exploring {
		b.Fatal("never in the 3D view")
	}
	for range 60 { // the first textures decode and upload
		f.Frame(1280, 800)
	}
	keys := []uint16{0x4800, 0x4D00, 0x4800, 0x4B00}
	b.ResetTimer()
	for i := range b.N {
		if i%60 == 0 {
			f.Key(keys[i/60%len(keys)])
		}
		f.Frame(1280, 800)
		f.Tick()
	}
}

// A frame of a fight (autocombat: the computer fights for the party).
func BenchmarkCombatFrame(b *testing.B) {
	f := frameFront(b, "autocombat")
	fighting := func() bool { _, ok := f.Game.G.State().(game.Fighting); return ok }
	for i := 0; i < 600 && !fighting(); i++ {
		f.Frame(1280, 800)
	}
	if !fighting() {
		b.Skip("no fight")
	}
	for range 60 {
		f.Frame(1280, 800)
	}
	b.ResetTimer()
	for range b.N {
		f.Frame(1280, 800)
		f.Tick()
	}
}
