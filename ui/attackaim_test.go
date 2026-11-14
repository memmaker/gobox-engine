package ui

import (
	"slices"
	"testing"

	"gobox/game"
	"gobox/rules"
)

// TestAttackAims: Attack before a target (the bar's, F) aims at the foes the
// game offers by id, the cursor on the first; a square there attacks that id.
func TestAttackAims(t *testing.T) {
	f := &fake{st: game.Fighting{Common: game.Common{Allowed: game.Allowed{Plain: []game.Action{game.Attack{Target: 2}, game.EndTurn{}}}},
		Field: game.Battlefield{Units: []game.Unit{{ID: 1, At: rules.BoardPos{X: 1, Y: 1}}, {ID: 2, At: rules.BoardPos{X: 5, Y: 3}, Foe: true}}}}}
	r := newRules(f, game.Data{})
	if res := r.send(game.Attack{Target: 0}); res.Err != nil || len(f.sent) != 0 || !shownOf(r).Combat.Turn.Aim || r.cursor != (rules.BoardPos{X: 5, Y: 3}) {
		t.Fatalf("aim start %+v sent %v", res, f.sent)
	}
	r.send(game.Aim{At: rules.BoardPos{X: 5, Y: 3}})
	if len(f.sent) != 1 || f.sent[0] != game.Action(game.Attack{Target: 2}) {
		t.Fatalf("sent %v", f.sent)
	}
}

// TestFledLeaves: a unit that flees leaves the field as one that falls.
func TestFledLeaves(t *testing.T) {
	if es := events([]game.Event{game.Fled{Unit: 4}}); len(es) != 1 || es[0].Kind != game.EvDown || es[0].Unit != 4 {
		t.Fatalf("%+v", es)
	}
}

// TestShootTargets: aiming a shot, the foes on its squares are the targets
// and the cursor is on the nearest, never (0,0).
func TestShootTargets(t *testing.T) {
	shoot := game.Shoot{}
	f := &fake{st: game.Fighting{Common: game.Common{Allowed: game.Allowed{Aimed: []game.Aimed{{Action: shoot, Squares: []rules.BoardPos{{X: 4, Y: 4}, {X: 6, Y: 6}, {X: 9, Y: 9}}, Targets: 1}}}},
		Field: game.Battlefield{Turn: 1, Units: []game.Unit{{ID: 1, At: rules.BoardPos{X: 1, Y: 1}}, {ID: 2, At: rules.BoardPos{X: 9, Y: 9}, Foe: true}, {ID: 3, At: rules.BoardPos{X: 6, Y: 6}, Foe: true}}}}}
	r := newRules(f, game.Data{})
	r.send(game.Shoot{})
	c := shownOf(r).Combat
	if c == nil || !c.Turn.Aim || len(c.Targets) != 2 || r.cursor != (rules.BoardPos{X: 6, Y: 6}) {
		t.Fatalf("%+v", c)
	}
}

// TestQuickAllOption: Quick for the whole party, when Allowed has it, is a
// bar entry of its own that sends Quick{All: true}.
func TestQuickAllOption(t *testing.T) {
	r := frontState{data: game.Data{VerbKeys: map[string]rune{"Quick All": 'U'}}}
	f := game.Fighting{Common: game.Common{Allowed: game.Allowed{Plain: []game.Action{game.Quick{}, game.Quick{All: true}}}}}
	c := r.combat(f)
	i := slices.IndexFunc(c.Turn.Options, func(o game.Option) bool { return o.Do == any(game.Quick{All: true}) })
	if i < 0 || c.Turn.Options[i].Key == 0 {
		t.Fatalf("%+v", c.Turn.Options)
	}
}

// TestAimNoFoeInReach: aiming with no foe on the aim's squares, Tab still
// cycles every foe and the cursor is drawn on the nearest, not lost.
func TestAimNoFoeInReach(t *testing.T) {
	shoot := game.Shoot{}
	f := &fake{st: game.Fighting{Common: game.Common{Allowed: game.Allowed{Aimed: []game.Aimed{{Action: shoot, Squares: []rules.BoardPos{{X: 2, Y: 1}}, Targets: 1}}}},
		Field: game.Battlefield{Turn: 1, Units: []game.Unit{{ID: 1, At: rules.BoardPos{X: 1, Y: 1}}, {ID: 2, At: rules.BoardPos{X: 9, Y: 9}, Foe: true}, {ID: 3, At: rules.BoardPos{X: 4, Y: 3}, Foe: true}}}}}
	r := newRules(f, game.Data{})
	r.send(game.Shoot{})
	c := shownOf(r).Combat
	if c == nil || !c.Turn.Aim || len(c.Targets) != 2 || r.cursor != (rules.BoardPos{X: 4, Y: 3}) {
		t.Fatalf("%+v", c)
	}
}

func events(es []game.Event) []game.Report { return eventsAt(es, defaultSpeed*speedStepMs, nil) }
