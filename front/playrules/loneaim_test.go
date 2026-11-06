package playrules

import (
	"image"
	"testing"

	"gobox/front"
	"gobox/front/classic"
	"gobox/game"
	"gobox/game/play"
)

// BR2's T: a pending square with one Aimed and Leave starts aiming at once,
// the cursor on the first target; Enter (ActAim there) answers it with the
// Attack at that square, Esc is Leave.
func TestLoneAimWithLeave(t *testing.T) {
	foe := image.Pt(4, 2)
	fight := play.Fighting{Common: play.Common{Allowed: play.Allowed{Plain: []play.Action{play.Leave{}},
		Aimed: []play.Aimed{{Action: play.Attack{}, Squares: []image.Point{foe, {5, 2}}, Targets: 1}}}},
		Field: play.Battlefield{Turn: 1, Units: []play.Unit{{ID: 1, Member: 0}, {ID: 2, Member: -1, Foe: true, At: foe}}}}
	f := &fake{st: fight}
	g := classic.NewGame(New(f, game.Data{}), game.Data{})
	v := g.View()
	if v.Battle == nil || v.Battle.Turn == nil || !v.Battle.Turn.Aim || v.Battle.Cursor != foe {
		t.Fatalf("not aiming at the first target: %+v", v.Battle)
	}
	if err := g.Answer(front.Action{Kind: front.ActAim, X: foe.X, Y: foe.Y}); err != nil {
		t.Fatal(err)
	}
	if len(f.sent) != 1 {
		t.Fatalf("sent %#v", f.sent)
	}
	if a, ok := f.sent[0].(play.Attack); !ok || a.At == nil || *a.At != foe {
		t.Fatalf("sent %#v", f.sent[0])
	}
	g.View()
	if g.Answer(front.Action{Kind: front.ActClose}); len(f.sent) != 2 || f.sent[1] != play.Action(play.Leave{}) {
		t.Fatalf("Esc sent %#v", f.sent)
	}
}
