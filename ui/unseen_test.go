package ui

import (
	"slices"
	"testing"

	"gobox/pool"
)

// devScenes are the engine's dev scenes (engine.DevScene) the invariant walks.
var devScenes = []string{"shop", "hall", "hallpoor", "temple", "appraise", "combat", "wildcombat", "castfight", "missile", "meleefight",
	"fleefight", "reservefight", "overland", "bestiary", "parlay", "password", "portrait", "sheetcast", "cast", "castko", "who", "effects", "journal",
	"talljournal", "runes", "inventory", "lockeddoor", "dual", "camp", "resting", "sheet", "items", "usable", "paladin", "drop", "learn",
	"loot", "lootdetect"}

// porScenes are the dev scenes that name POR's own data (a tavern tale):
// no screen of the other games.
var porScenes = []string{"journal"}

// TestNoUnseenQuestion: in every dev scene of every Pool game imported
// here, and on each screen Enter and then Esc lead to from it, a question
// the game asks is on screen (Presenter.stuck; bug class: the hall's
// undrawn "Pick New Class:", answered by any key).
func TestNoUnseenQuestion(t *testing.T) {
	for _, g := range []string{"por", "cotab", "sotsb", "pod"} {
		t.Run(g, func(t *testing.T) { noUnseenQuestion(t, g) })
	}
}

func noUnseenQuestion(t *testing.T, g string) {
	for _, name := range devScenes {
		if g != "por" && slices.Contains(porScenes, name) {
			continue
		}
		for _, key := range []uint16{13, 27} {
			eng, err := pool.New("../data/"+g, "../data/"+g+"/sources")
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
				for range 3 { // as flowKey: the state settles and the screen is drawn
					f.Frame(flowW, flowH)
					f.Tick()
				}
				if f.P.stuck() {
					st := eng.State()
					t.Errorf("%s, %d keys %#x in: %T asks %s on no screen", name, step, key, st, f.P.asksID())
					break
				}
				f.Key(key)
			}
		}
	}
}
