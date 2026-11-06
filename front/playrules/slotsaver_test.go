package playrules

import (
	"os"
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// saving is slotted that also saves into its slots (play.SlotSaver).
type saving struct {
	slotted
	saved int
}

func (s *saving) SaveSlot(n int) error { s.saved = n + 1; return nil }

// TestSlotSaverOneStore: a game that saves into its own slots is the only
// store: no gobox slot file is written or listed.
func TestSlotSaverOneStore(t *testing.T) {
	g := &saving{slotted: slotted{fake: fake{st: play.Exploring{}}}}
	r := New(g, game.Data{}).(*rules)
	r.dir = t.TempDir()
	if res := r.Apply(game.Save{Name: "A"}); res.Err != nil || g.saved != 1 {
		t.Fatalf("save over A: %+v slot %d", res, g.saved-1)
	}
	if res := r.Apply(game.Save{Name: "New"}); res.Err != nil || g.saved != 2 {
		t.Fatalf("new save: %+v slot %d", res, g.saved-1)
	}
	if es, _ := os.ReadDir(r.dir); len(es) > 0 {
		t.Fatalf("gobox wrote its own slot files %v", es)
	}
	if s := r.State(); len(s.Saves) != 1 {
		t.Fatalf("saves %v: want the game's one", s.Saves)
	}
}
