package engine

import (
	"gobox/ecl"
	"gobox/game"
	"gobox/rules"
	"gobox/rules1e"
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

// A scripted POD fight gets the adventure's reserve (GameTables.Reserve):
// ten of the area's monster off the map, counting for nothing, until
// Reinforce gates two in beside the first combatant, on the caller's side
// (pod-re CharFightsAlone, combatst.pas:822, magic.pas:2870). Bug class:
// the engine asked the rules for a fact they hold for no game ("fact not
// held"): every scripted POD fight panicked.
func TestPodReserve(t *testing.T) {
	e := devFirstTurnOn(t, advEngine(t, advEnvPOD), "reservefight", false)
	s := e.state()
	want := e.game().Reserve.Count
	var reserve []*rules1e.Character
	for _, c := range s.Party {
		if c.Status == rules1e.StatusReserve {
			reserve = append(reserve, c)
			if c.Okay != 0 || s.CombatSize(c) != 0 || s.CombatIndex(c) == 0 {
				t.Fatalf("a reserve unit in play %d, size %d, slot %d", c.Okay, s.CombatSize(c), s.CombatIndex(c))
			}
		}
	}
	if want == 0 || len(reserve) != want {
		t.Fatalf("%d in reserve, want the table's %d", len(reserve), want)
	}
	units := func() (foes, ours int) {
		f, ok := e.build().(game.Fighting)
		if !ok {
			t.Fatalf("not a fight: %T", e.build())
		}
		for _, u := range f.Field.Units {
			if u.Foe {
				foes++
			} else {
				ours++
			}
		}
		return
	}
	foes, ours := units()
	if foes != 3 {
		t.Errorf("%d foes on the board, want the scene's 3: the reserve is off it", foes)
	}
	if s.FoesHealthPercent(); s.FoesHealth != 100 {
		t.Errorf("the monsters' health %d%%, want 100: the reserve counts for nothing", s.FoesHealth)
	}
	by, first := s.Acting(), s.Party[0]
	if n := s.Reinforce(by, 2); n != 2 {
		t.Fatalf("reinforced by %d, want 2", n)
	}
	for _, c := range reserve[:2] {
		dx, dy := s.CombatX(c)-s.CombatX(first), s.CombatY(c)-s.CombatY(first)
		if c.Status != rules1e.StatusOkay || c.Okay != 1 || c.Side != by.Side || c.HP != c.MaxHP || s.CombatSize(c) == 0 || max(dx, -dx, dy, -dy) != 1 {
			t.Errorf("a reinforcement: status %d, in play %d, side %d, hp %d/%d, size %d, %d,%d from the first combatant", c.Status, c.Okay, c.Side, c.HP, c.MaxHP, s.CombatSize(c), dx, dy)
		}
	}
	if reserve[2].Status != rules1e.StatusReserve {
		t.Error("a third unit left the reserve")
	}
	e.rebuild()
	if f2, o2 := units(); f2 != foes || o2 != ours+2 {
		t.Errorf("the board after: %d foes, %d of ours; want %d, %d", f2, o2, foes, ours+2)
	}
	// the fight over: the reserve was not beaten, and it leaves with the monsters, the reinforcements too
	res := e.ruleset().EndFight(rules.Fight{}, nil)
	if res.Beaten != 0 || len(s.Party) != ours {
		t.Errorf("after the fight: %d beaten, %d left in the party; want 0, the %d members", res.Beaten, len(s.Party), ours)
	}
}
