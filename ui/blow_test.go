package ui

import (
	"slices"
	"testing"

	"gobox/game"
)

// TestBlowShows: a blow the game reports (Attacked, Narrated) reaches the
// front as a swing, a hit, its damage and the line. The battle log is the
// engine's (Battlefield.Log, TestCombatLogNote), never built from the events.
func TestBlowShows(t *testing.T) {
	fight := game.Fighting{Common: game.Common{Text: []string{""}, Allowed: game.Allowed{Plain: []game.Action{game.Attack{Target: 2}}}},
		Field: game.Battlefield{Turn: 1, Units: []game.Unit{{ID: 1, Member: 0}, {ID: 2, Member: -1, Foe: true}}}}
	f := &fake{st: fight, out: game.Outcome{Events: []game.Event{game.Attacked{From: 1, To: 2, Hit: true, Damage: 5}, game.Narrated{Text: "ARIN HITS FOR 5"}}}}
	g := NewGame(f, game.Data{})
	if err := g.Answer(Action{Kind: ActAttack, N: 2}); err != nil {
		t.Fatal(err)
	}
	evs := settled(g)
	var kinds []game.EventKind
	dmg, line := 0, false
	for _, e := range evs {
		kinds = append(kinds, e.Kind)
		if e.Kind == game.EvDamage && e.Unit == 2 { // over the one struck
			dmg = e.N
		}
		line = line || e.Kind == game.EvMessage && e.Text == "ARIN HITS FOR 5"
	}
	if !slices.Contains(kinds, game.EvAttack) || !slices.Contains(kinds, game.EvHit) || dmg != 5 || !line {
		t.Fatalf("events %+v", evs)
	}
}

// TestBlowHoldsSpeedDelay: a blow is held the game speed's delay (speed ×
// 100 ms, the originals'), and SetSpeed changes it.
func TestBlowHoldsSpeedDelay(t *testing.T) {
	fight := game.Fighting{Common: game.Common{Allowed: game.Allowed{Plain: []game.Action{game.Attack{Target: 2}}}},
		Field: game.Battlefield{Turn: 1, Units: []game.Unit{{ID: 1, Member: 0}, {ID: 2, Member: -1, Foe: true}}}}
	f := &fake{st: fight, out: game.Outcome{Events: []game.Event{game.Attacked{From: 1, To: 2}}}}
	r := newRules(f, game.Data{})
	held := func() int {
		for _, e := range r.send(game.Attack{Target: 2}).Events {
			if e.Kind == game.EvAttack {
				return e.Ms
			}
		}
		return -1
	}
	if ms := held(); ms != 400 {
		t.Fatalf("default speed: %d ms", ms)
	}
	r.send(tSetSpeed{N: 2})
	if ms := held(); ms != 200 {
		t.Fatalf("speed 2: %d ms", ms)
	}
}
