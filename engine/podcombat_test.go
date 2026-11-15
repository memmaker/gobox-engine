package engine

import (
	"gobox/ecl"
	"gobox/rules"
	"slices"
	"testing"
)

// podEngine is POD's engine past its start (skipped when POD is not imported).
func podEngine(t *testing.T) *Engine {
	t.Helper()
	e, err := NewPool("../data/pod", "../data/pod/sources")
	if err != nil {
		t.Skip(err)
	}
	e.Cfg.SaveDir = t.TempDir()
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	return e
}

// POD's Combat reads the area's tile set and the vault request (pod-re
// EclCombat, LoadCombatTileSet). Bug class: shared code read a world word
// the rules keep for no game (Engine.worldWord panics): a POD fight outside
// a dungeon and a Combat with no monsters crashed.
func TestPodCombatWords(t *testing.T) {
	e := podEngine(t)
	if !slices.Contains(e.eclVersion.Behaviour.NoMonsterCombat, ecl.CombatVault) {
		t.Fatal("POD's dialect: no vault screen")
	}
	for _, l := range e.artNames().WildTiles { // the wilderness tiles' block by the set: 1 and 2 their own, else 0
		for set, want := range []int{0, 1, 2, 0} {
			e.setWorldWord(rules.WordSymbolSet, set)
			if got := e.tileBlock(l); got != want {
				t.Errorf("set %d: wilderness tile block %d, want %d", set, got, want)
			}
		}
	}
	if e.noMonsterScreen() != nil {
		t.Error("a screen opens with nothing asked")
	}
	r, ok := e.layoutField(fieldVault)
	if !ok {
		t.Fatal("POD's layout has no " + fieldVault)
	}
	e.host.Set(r, 1)
	if e.noMonsterScreen() == nil || e.host.Get(r) != 0 {
		t.Errorf("the vault asked for: no screen, or the request stays (%d)", e.host.Get(r))
	}
}
