package present

import (
	"image"
	"slices"
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// TestAttackAims: Attack before a target (the bar's, F) aims at the foes the
// game offers by id, the cursor on the first; a square there attacks that id.
func TestAttackAims(t *testing.T) {
	f := &fake{st: play.Fighting{Common: play.Common{Allowed: play.Allowed{Plain: []play.Action{play.Attack{Target: 2}, play.EndTurn{}}}},
		Field: play.Battlefield{Units: []game.Unit{{ID: 1, At: image.Pt(1, 1)}, {ID: 2, At: image.Pt(5, 3), Foe: true}}}}}
	r := newRules(f, game.Data{})
	if res := r.send(play.Attack{Target: 0}); res.Err != nil || len(f.sent) != 0 || !shownOf(r).Combat.Turn.Aim || shownOf(r).Combat.Cursor != image.Pt(5, 3) {
		t.Fatalf("aim start %+v sent %v", res, f.sent)
	}
	r.send(play.Aim{At: image.Pt(5, 3)})
	if len(f.sent) != 1 || f.sent[0] != play.Action(play.Attack{Target: 2}) {
		t.Fatalf("sent %v", f.sent)
	}
}

// TestFledLeaves: a unit that flees leaves the field as one that falls.
func TestFledLeaves(t *testing.T) {
	if es := events([]play.Event{play.Fled{Unit: 4}}); len(es) != 1 || es[0].Kind != game.EvDown || es[0].Unit != 4 {
		t.Fatalf("%+v", es)
	}
}

// TestShootTargets: aiming a shot, the foes on its squares are the targets
// and the cursor is on the nearest, never (0,0).
func TestShootTargets(t *testing.T) {
	shoot := play.Shoot{}
	f := &fake{st: play.Fighting{Common: play.Common{Allowed: play.Allowed{Aimed: []play.Aimed{{Action: shoot, Squares: []image.Point{{4, 4}, {6, 6}, {9, 9}}, Targets: 1}}}},
		Field: play.Battlefield{Turn: 1, Units: []game.Unit{{ID: 1, At: image.Pt(1, 1)}, {ID: 2, At: image.Pt(9, 9), Foe: true}, {ID: 3, At: image.Pt(6, 6), Foe: true}}}}}
	r := newRules(f, game.Data{})
	r.send(play.Shoot{})
	c := shownOf(r).Combat
	if c == nil || !c.Turn.Aim || len(c.Targets) != 2 || c.Cursor != image.Pt(6, 6) {
		t.Fatalf("%+v", c)
	}
}

// TestQuickAllOption: Quick for the whole party, when Allowed has it, is a
// bar entry of its own that sends Quick{All: true}.
func TestQuickAllOption(t *testing.T) {
	r := rules{data: game.Data{VerbKeys: map[string]rune{"Quick All": 'U'}}}
	f := play.Fighting{Common: play.Common{Allowed: play.Allowed{Plain: []play.Action{play.Quick{}, play.Quick{All: true}}}}}
	c := r.combat(f)
	i := slices.IndexFunc(c.Turn.Options, func(o play.Option) bool { return o.Do == any(play.Quick{All: true}) })
	if i < 0 || c.Turn.Options[i].Key == 0 {
		t.Fatalf("%+v", c.Turn.Options)
	}
}

// TestAimNoFoeInReach: aiming with no foe on the aim's squares, Tab still
// cycles every foe and the cursor is drawn on the nearest, not lost.
func TestAimNoFoeInReach(t *testing.T) {
	shoot := play.Shoot{}
	f := &fake{st: play.Fighting{Common: play.Common{Allowed: play.Allowed{Aimed: []play.Aimed{{Action: shoot, Squares: []image.Point{{2, 1}}, Targets: 1}}}},
		Field: play.Battlefield{Turn: 1, Units: []game.Unit{{ID: 1, At: image.Pt(1, 1)}, {ID: 2, At: image.Pt(9, 9), Foe: true}, {ID: 3, At: image.Pt(4, 3), Foe: true}}}}}
	r := newRules(f, game.Data{})
	r.send(play.Shoot{})
	c := shownOf(r).Combat
	if c == nil || !c.Turn.Aim || !c.CursorOn || len(c.Targets) != 2 || c.Cursor != image.Pt(4, 3) {
		t.Fatalf("%+v", c)
	}
}
