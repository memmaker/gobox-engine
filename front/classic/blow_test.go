package classic

import (
	"slices"
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// TestBlowShows: a blow the game reports (Attacked, Narrated) reaches the
// front as a swing, a hit, its damage and the line, and the fight's text
// pane keeps the line even when the game's own Text is blank.
func TestBlowShows(t *testing.T) {
	fight := play.Fighting{Common: play.Common{Text: []string{""}, Allowed: play.Allowed{Plain: []play.Action{play.Attack{Target: 2}}}},
		Field: play.Battlefield{Turn: 1, Units: []play.Unit{{ID: 1, Member: 0}, {ID: 2, Member: -1, Foe: true}}}}
	f := &fake{st: fight, out: play.Outcome{Events: []play.Event{play.Attacked{From: 1, To: 2, Hit: true, Damage: 5}, play.Narrated{Text: "ARIN HITS FOR 5"}}}}
	g := NewGame(f, game.Data{})
	if err := g.Answer(front.Action{Kind: front.ActAttack, N: 2}); err != nil {
		t.Fatal(err)
	}
	v := g.Settle()
	var kinds []front.EventKind
	dmg, line := 0, false
	for _, e := range v.Events {
		kinds = append(kinds, e.Kind)
		if e.Kind == front.EvDamage && e.Unit == 2 { // over the one struck
			dmg = e.Amount
		}
		line = line || e.Kind == front.EvMessage && e.Text == "ARIN HITS FOR 5"
	}
	if !slices.Contains(kinds, front.EvAttack) || !slices.Contains(kinds, front.EvHit) || dmg != 5 || !line {
		t.Fatalf("events %+v", v.Events)
	}
	if v.Battle == nil || !slices.Contains(v.Battle.Log, "ARIN HITS FOR 5") {
		t.Fatalf("log %v", v.Battle != nil)
	}
}

// TestBlowHoldsSpeedDelay: a blow is held the game speed's delay (speed ×
// 100 ms, the originals'), and SetSpeed changes it.
func TestBlowHoldsSpeedDelay(t *testing.T) {
	fight := play.Fighting{Common: play.Common{Allowed: play.Allowed{Plain: []play.Action{play.Attack{Target: 2}}}},
		Field: play.Battlefield{Turn: 1, Units: []play.Unit{{ID: 1, Member: 0}, {ID: 2, Member: -1, Foe: true}}}}
	f := &fake{st: fight, out: play.Outcome{Events: []play.Event{play.Attacked{From: 1, To: 2}}}}
	r := newRules(f, game.Data{})
	held := func() int {
		for _, e := range r.Apply(game.Attack{Target: 2}).Events {
			if e.Kind == game.EvAttack {
				return e.Ms
			}
		}
		return -1
	}
	if ms := held(); ms != 400 {
		t.Fatalf("default speed: %d ms", ms)
	}
	r.Apply(game.SetSpeed{N: 2})
	if ms := held(); ms != 200 {
		t.Fatalf("speed 2: %d ms", ms)
	}
}
