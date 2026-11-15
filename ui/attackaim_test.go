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

// TestShootTargets: aiming a shot, the figures on its squares are the targets
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

// TestAimNoFigureInReach: aiming ground with no figure on the aim's squares,
// nobody is a target (none highlighted) and the cursor is on the nearest
// square the aim takes, not lost.
func TestAimNoFigureInReach(t *testing.T) {
	shoot := game.Shoot{}
	f := &fake{st: game.Fighting{Common: game.Common{Allowed: game.Allowed{Aimed: []game.Aimed{{Action: shoot, Squares: []rules.BoardPos{{X: 5, Y: 1}, {X: 2, Y: 1}}, Targets: 1}}}},
		Field: game.Battlefield{Turn: 1, Units: []game.Unit{{ID: 1, At: rules.BoardPos{X: 1, Y: 1}}, {ID: 2, At: rules.BoardPos{X: 9, Y: 9}, Foe: true}, {ID: 3, At: rules.BoardPos{X: 4, Y: 3}, Foe: true}}}}}
	r := newRules(f, game.Data{})
	r.send(game.Shoot{})
	c := shownOf(r).Combat
	if c == nil || !c.Turn.Aim || len(c.Targets) != 0 || r.cursor != (rules.BoardPos{X: 2, Y: 1}) {
		t.Fatalf("targets %v cursor %v", c.Targets, r.cursor)
	}
}

// TestAimTargetRules: an aimed action with no square its rules take is
// refused with a note and no aim; aiming, every figure on the aim's squares
// is a target, friend or foe, and no other; giving the aim up leaves none.
func TestAimTargetRules(t *testing.T) {
	cast, shoot := game.Cast{Member: 0, Spell: 3}, game.Shoot{}
	units := []game.Unit{{ID: 1, At: rules.BoardPos{X: 1, Y: 1}}, {ID: 2, At: rules.BoardPos{X: 2, Y: 1}}, {ID: 3, At: rules.BoardPos{X: 4, Y: 3}, Foe: true},
		{ID: 4, At: rules.BoardPos{X: 9, Y: 9}, Foe: true}, {ID: 5, At: rules.BoardPos{X: 4, Y: 4}, Foe: true, Dead: true}}
	f := &fake{st: game.Fighting{Common: game.Common{Allowed: game.Allowed{Plain: []game.Action{game.EndTurn{}}, Aimed: []game.Aimed{{Action: cast, Targets: 1},
		{Action: shoot, Squares: []rules.BoardPos{{X: 2, Y: 1}, {X: 4, Y: 3}, {X: 4, Y: 4}, {X: 6, Y: 6}}, Targets: 1}}}},
		Field: game.Battlefield{Turn: 1, Units: units}}}
	r := newRules(f, game.Data{})
	for _, a := range []game.Action{cast, game.Attack{}} { // no square to cast at, no foe to attack
		r.note = game.Note{}
		if res := r.send(a); res.Err != nil || len(f.sent) != 0 || r.aim != nil || r.note.Text != noTargetsText {
			t.Fatalf("%T: %+v sent %v aim %v note %q", a, res, f.sent, r.aim, r.note.Text)
		}
	}
	r.send(shoot)
	if c := shownOf(r).Combat; !c.Turn.Aim || !slices.Equal(c.Targets, []int{2, 3}) {
		t.Fatalf("aiming: targets %v, want the ally and the foe on its squares", c.Targets)
	}
	r.send(game.Cancel{})
	if c := shownOf(r).Combat; c.Turn.Aim || len(c.Marks) != 0 || len(f.sent) != 0 {
		t.Fatalf("cancelled: aim %v marks %v sent %v", c.Turn.Aim, c.Marks, f.sent)
	}
}

func events(es []game.Event) []game.Report { return eventsAt(es, defaultSpeed*speedStepMs, nil) }

// TestTurnsCast: the fight menu's Cast (Member -1) of an aimed spell is
// gobox's aim over the squares the game offers; one cast over rounds, which
// the game offers plain, is sent at once with no aim: its aim comes when it
// goes off (the game's own, alone in Allowed).
func TestTurnsCast(t *testing.T) {
	now, timed := game.Cast{Member: 2, Spell: 34}, game.Cast{Member: 2, Spell: 3}
	st := game.Fighting{Common: game.Common{Allowed: game.Allowed{Plain: []game.Action{timed, game.EndTurn{}},
		Aimed: []game.Aimed{{Action: now, Squares: []rules.BoardPos{{X: 4, Y: 3}}, Targets: 1}}}},
		Field: game.Battlefield{Turn: 1, Units: []game.Unit{{ID: 1, At: rules.BoardPos{X: 1, Y: 1}}, {ID: 3, At: rules.BoardPos{X: 4, Y: 3}, Foe: true}}}}
	f := &fake{st: st}
	r := newRules(f, game.Data{})
	if r.send(game.Cast{Member: -1, Spell: 34}); r.aim == nil || len(f.sent) != 0 {
		t.Fatalf("the aimed spell: aim %v sent %v", r.aim, f.sent)
	}
	r.send(game.Aim{At: rules.BoardPos{X: 4, Y: 3}})
	if c, ok := f.sent[len(f.sent)-1].(game.Cast); len(f.sent) != 1 || !ok || c.Spell != 34 || len(c.Ats) != 1 || r.aim != nil {
		t.Fatalf("aimed: sent %+v aim %v", f.sent, r.aim)
	}
	f.sent = nil
	if r.send(game.Cast{Member: -1, Spell: 3}); r.aim != nil || len(f.sent) != 1 {
		t.Fatalf("the spell cast over rounds: aim %v sent %+v, want it sent with no aim", r.aim, f.sent)
	}
	// it goes off: the game's aim alone, with no square its rules take, is given up with the note
	f.st = game.Fighting{Common: game.Common{Allowed: game.Allowed{Plain: []game.Action{game.Leave{}}, Aimed: []game.Aimed{{Action: timed, Targets: 1}}}}, Field: st.Field}
	f.sent, r.note = nil, game.Note{}
	r.send(game.EndTurn{})
	if len(f.sent) != 2 || f.sent[1] != game.Action(game.Leave{}) || r.note.Text != noTargetsText {
		t.Fatalf("no square to aim at: sent %+v note %q", f.sent, r.note.Text)
	}
}
