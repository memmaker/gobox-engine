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

// newestFirst is a SlotSaver listing its slots newest first: a save moves
// the slot saved to the top, shifting the others.
type newestFirst struct {
	fake
	names []string
}

func (g *newestFirst) Slots() []game.SaveSlot {
	var s []game.SaveSlot
	for _, n := range g.names {
		s = append(s, game.SaveSlot{Name: n})
	}
	return s
}
func (g *newestFirst) LoadSlot(n int) error { g.names = append(g.names, "loaded:"+g.names[n]); return nil }
func (g *newestFirst) SaveSlot(n int) error {
	name := "new"
	if n < len(g.names) {
		name = g.names[n]
		g.names = append(g.names[:n:n], g.names[n+1:]...)
	}
	g.names = append([]string{name}, g.names...)
	return nil
}

// TestSlotSaverNewestFirst: indexes shift after a save; gobox finds each
// slot by name in Slots() read anew, never by an index kept across a save.
func TestSlotSaverNewestFirst(t *testing.T) {
	g := &newestFirst{fake: fake{st: play.Exploring{}}, names: []string{"A", "B", "C"}}
	r := New(g, game.Data{}).(*rules)
	r.dir = t.TempDir()
	r.Apply(game.Save{Name: "C"}) // C to the top: C A B
	r.Apply(game.Save{Name: "B"}) // B C A
	if got := g.names; len(got) != 3 || got[0] != "B" || got[1] != "C" || got[2] != "A" {
		t.Fatalf("slots %v", got)
	}
	if s := r.State().Saves; s[0].Name != "B" {
		t.Fatalf("listed %v", s)
	}
}
