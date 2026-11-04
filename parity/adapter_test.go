package parity

import (
	"reflect"
	"slices"
	"testing"
	"time"

	"gobox/front"
	"gobox/front/classic"
	"gobox/front/present"
)

// known are the View fields classic cannot yet build from State alone, per
// scene: Battle (the ground's tiles, icons, effects, paths and strike are
// the game's own View, not State), X and Y on the overland (por keeps the
// dungeon's square there). Any other field that differs fails.
var known = map[string][]string{"combat": {"Battle"}, "overland": {"X", "Y"}}

// TestAdapter runs por twice over each dev scene and the same actions: as a
// front.Game itself (rules.Engine) and through classic.Game over its
// game.Rules. The Views must be DeepEqual.
func TestAdapter(t *testing.T) {
	actions := []front.Action{{}, {Kind: front.ActResize, X: 50, Y: 30}}
	for _, scene := range []string{"", "shop", "hall", "temple", "appraise", "combat", "overland", "bestiary", "parlay", "password", "who", "effects", "journal", "inventory", "lockeddoor", "loot", "cast"} {
		t.Run("scene "+scene, func(t *testing.T) {
			a, b := engine(t, true, scene), engine(t, true, scene)
			g := classic.NewGame(b.Rules(), b.Data())
			var va, vb front.View
			for i, act := range actions {
				va = a.Apply(act)
				g.Answer(act)
				vb = g.Settle()
				if i == 0 {
					continue
				}
			}
			defer a.Apply(front.Action{Kind: front.ActResize, X: 40, Y: 25}) // the page size is global in por: put it back
			for _, v := range []*front.View{&va, &vb} {
				v.Seq, v.Note.At, v.Events = 0, time.Time{}, nil
			}
			for _, n := range diffFields(va, vb) {
				if slices.Contains(known[scene], n) {
					t.Logf("known gap: %s", n)
				} else {
					t.Errorf("%s differs:\n  engine  %+v\n  adapter %+v", n, reflect.ValueOf(va).FieldByName(n), reflect.ValueOf(vb).FieldByName(n))
				}
			}
		})
	}
}

// TestGoldensAdapter: por's creation goldens through classic.Game alone (no
// engine front.Game under it), the actions answered by the adapter.
func TestGoldensAdapter(t *testing.T) {
	e := partyMenu(t)
	g := classic.NewGame(e.Rules(), e.Data())
	p := &present.Presenter{}
	p.SetEngine(g)
	shown := func(as ...front.Action) front.View {
		for _, a := range as {
			p.Decide(g.View(), a)
			for _, f := range p.Calls() {
				f(g)
			}
			for _, a := range p.Actions() {
				g.Answer(a)
			}
			g.Settle()
		}
		return p.Own(g.View())
	}
	v := shown(front.Action{Kind: front.ActCreate})
	if v.Creation == nil {
		t.Fatal("no creation")
	}
	golden(t, "creation", v)
	golden(t, "pickhead", shown(front.Action{Kind: front.ActPicker, N: front.PickHead}))
}

// TestSheetAdapter: por's sheet goldens through classic.Game alone.
func TestSheetAdapter(t *testing.T) {
	e := engine(t, true, "")
	g := classic.NewGame(e.Rules(), e.Data())
	g.OpenSheet(-1)
	v := g.Settle()
	if v.Sheet == nil {
		t.Fatal("no sheet")
	}
	golden(t, "sheet", v)
	g.ShowMember(1)
	golden(t, "sheet2", g.Settle())
}
