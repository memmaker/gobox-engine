package ui

import (
	"testing"

	"gobox/engine"
)

// devScenes are the engine's dev scenes (engine.DevScene) the invariant walks.
var devScenes = []string{"shop", "hall", "hallpoor", "temple", "appraise", "combat", "wildcombat", "castfight", "missile", "meleefight",
	"fleefight", "overland", "bestiary", "parlay", "password", "portrait", "sheetcast", "cast", "castko", "who", "effects", "journal",
	"talljournal", "runes", "inventory", "lockeddoor", "dual", "camp", "resting", "sheet", "items", "usable", "paladin", "drop", "learn",
	"loot", "lootdetect"}

// TestNoUnseenQuestion: in every dev scene, and on each screen Enter and
// then Esc lead to from it, a question the game asks is on screen: a
// screen's own or the question window's (Presenter.unseen; bug class: the
// hall's undrawn "Pick New Class:", answered by any key).
func TestNoUnseenQuestion(t *testing.T) {
	for _, name := range devScenes {
		for _, key := range []uint16{13, 27} {
			eng, err := engine.NewPool("../data/por", "../data/por/sources")
			if err != nil {
				t.Skip(err)
			}
			eng.Cfg.SaveDir = t.TempDir()
			seed := uint32(7)
			eng.FixedSeed, eng.Explore = &seed, true
			eng.FullIntro.Store(false)
			if eng.DevAction = eng.DevScene(name); eng.DevAction == nil {
				t.Fatalf("no dev scene %q", name)
			}
			data := eng.Data()
			f := NewFront(NewGame(eng, data), data)
			f.Start("")
			f.Frame(flowW, flowH)
			f.Game.SequenceDone("intro")
			f.Game.TitleDone()
			f.Game.Settle()
			for step := range 8 {
				for range 10 {
					f.Frame(flowW, flowH)
					f.Tick()
				}
				if f.P.unseen() {
					st := eng.State()
					t.Errorf("%s, %d keys %#x in: %T asks %+v on no screen", name, step, key, st, f.Game.r.asking(st))
					break
				}
				f.Key(key)
			}
		}
	}
}
