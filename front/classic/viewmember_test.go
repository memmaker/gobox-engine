package classic

import (
	"slices"
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// TestViewMemberOpensSheet: View (a shop's, temple's) is front state: the
// member's sheet comes up with no call to the game, and CloseSheet ends it.
func TestViewMemberOpensSheet(t *testing.T) {
	f := &fake{st: play.Exploring{Common: play.Common{Party: []play.Member{{Name: "ARIN"}}}}}
	r := newRules(f, game.Data{})
	r.Apply(game.ViewMember{Member: 0})
	if shownOf(r).SheetPage == nil || len(f.sent) != 0 {
		t.Fatalf("sheet %v, sent %v", shownOf(r).SheetPage != nil, f.sent)
	}
	r.Apply(game.CloseSheet{})
	if shownOf(r).SheetPage != nil {
		t.Fatal("sheet still up")
	}
}

// Camp with no AdvanceTimeBy allowed offers no timed rest (Camp.NoRest).
func TestCampNoRest(t *testing.T) {
	for _, allow := range []bool{false, true} {
		st := play.Camping{}
		if allow {
			st.Allowed.Plain = []play.Action{play.AdvanceTimeBy{}}
		}
		if got := shownOf(newRules(&fake{st: st}, game.Data{})).Allowed["NoRest"]; got == allow {
			t.Errorf("AdvanceTimeBy allowed %v: NoRest %v", allow, got)
		}
	}
}

// TestFightView: a fight's turn offers View where the series keys it
// (VerbKeys["View"]) and the rules have no ViewTarget of their own; it opens the active member's sheet over the fight, with no
// call to the game, and CloseSheet returns to the same turn.
func TestFightView(t *testing.T) {
	fight := play.Fighting{Common: play.Common{Party: []play.Member{{Name: "ARIN"}, {Name: "MAREN"}}, Allowed: allow(play.EndTurn{}).Allowed},
		Field: play.Battlefield{Turn: 2, Units: []play.Unit{{ID: 1, Member: 0}, {ID: 2, Member: 1}}}}
	f := &fake{st: fight}
	if v := NewGame(f, game.Data{}).View(); slices.ContainsFunc(v.Battle.Turn.Options, func(o front.TurnOption) bool { return o.Label == "View" }) {
		t.Fatal("View the series doesn't key")
	}
	g := NewGame(f, game.Data{VerbKeys: map[string]rune{"View": 'V'}})
	n := -1
	for i, o := range g.View().Battle.Turn.Options {
		if o.Label == "View" {
			n = i
		}
	}
	if n < 0 {
		t.Fatalf("no View in %+v", g.View().Battle.Turn.Options)
	}
	if o := NewGame(f, game.Data{VerbKeys: map[string]rune{"View": 'V'}, VerbOrder: []string{"View", "End"}}).View().Battle.Turn.Options; o[0].Label != "View" {
		t.Errorf("VerbOrder: %+v, want View first", o)
	}
	g.Answer(front.Action{Kind: front.ActOption, N: n})
	if v := g.View(); v.SheetPage == nil || v.Sheet == nil || v.Battle == nil || v.Battle.Turn == nil || len(f.sent) != 0 {
		t.Fatalf("sheet %v, battle %v, sent %v", v.SheetPage != nil, v.Battle != nil, f.sent)
	}
	g.CloseSheet()
	if v := g.View(); v.SheetPage != nil || v.Battle == nil || v.Battle.Turn == nil || v.Battle.Turn.Unit != 2 {
		t.Fatal("not back on the turn")
	}
	f.st = play.Fighting{Common: play.Common{Party: fight.Party, Allowed: allow(play.ViewTarget{}).Allowed}, Field: fight.Field}
	views := 0
	for _, o := range g.View().Battle.Turn.Options {
		if o.Label == "View" {
			views++
		}
	}
	if views != 1 { // the rules' ViewTarget alone
		t.Fatalf("%d Views beside the rules' ViewTarget", views)
	}
}

// TestFightSheetPage: the rules' page (Member.Sheet) reaches classic's
// look-only fight sheet through game.Member.Page, not the bare fallback.
func TestFightSheetPage(t *testing.T) {
	fight := play.Fighting{Common: play.Common{Party: []play.Member{{Name: "ARIN", Sheet: &game.SheetPage{Name: "ARIN", Damage: "1d8", Sex: "Male", Age: 20}}}}}
	g := NewGame(&fake{st: fight}, game.Data{})
	g.OpenSheet(0)
	if sp := g.View().SheetPage; sp == nil || sp.Damage != "1d8" || sp.Sex != "Male" || sp.Age != 20 {
		t.Fatalf("page %+v", sp)
	}
}

// TestFightItems: Gear in a fight opens the turn member's items over the
// fight, with no call to the game; closing them returns to the same turn.
func TestFightItems(t *testing.T) {
	f := &fake{st: play.Fighting{Common: play.Common{Selected: 0, Party: []play.Member{{Name: "ARIN"}, {Name: "MAREN"}}},
		Field: play.Battlefield{Turn: 9, Units: []play.Unit{{ID: 9, Member: 1}}}}}
	r := newRules(f, game.Data{})
	r.Apply(game.OpenItems{Member: -1})
	if s := shownOf(r); s.Items != game.ItemsOwn || s.Combat == nil || len(f.sent) != 0 || s.Selected != 1 {
		t.Fatalf("items %v combat %v sent %v", s.Items, s.Combat != nil, f.sent)
	}
	r.Apply(game.CloseItems{})
	if s := shownOf(r); s.Items != game.ItemsNone || s.SheetPage != nil || s.Combat == nil {
		t.Fatalf("after close: items %v sheet %v", s.Items, s.SheetPage != nil)
	}
}

// TestViewGearOpensItems: a View menu's Gear the rules give as a
// front-state reply (Question.Do) opens that member's items, not a Reply.
func TestViewGearOpensItems(t *testing.T) {
	f := &fake{st: play.Fighting{Common: play.Common{Party: []play.Member{{Name: "ARIN"}, {Name: "MAREN"}, {Name: "VEX"}},
		Ask: &play.Question{Replies: []string{"Gear", "Exit"}, Do: []game.Decision{game.OpenItems{Member: 2}}}},
		Field: play.Battlefield{Turn: 9, Units: []play.Unit{{ID: 9, Member: 1}}}}}
	r := newRules(f, game.Data{})
	r.Apply(game.Answer{N: 0})
	if s := shownOf(r); s.Items != game.ItemsOwn || s.Selected != 2 || len(f.sent) != 0 || s.Pending != nil {
		t.Fatalf("items %v selected %d sent %v pending %v", s.Items, s.Selected, f.sent, s.Pending)
	}
	g := NewGame(f, game.Data{})
	g.r = r
	if v := g.View(); v.Inventory == nil || v.Ask != nil { // the items draw, over the question
		t.Fatalf("inventory %v ask %v", v.Inventory != nil, v.Ask != nil)
	}
	r.Apply(game.CloseItems{})
	if shownOf(r).Pending == nil {
		t.Fatal("the View menu is gone")
	}
}

// TestSpeedMenu: a Done menu's Speed (Question.Do SpeedMenu) is gobox's
// menu: Slower and Faster step the one speed (0..9), Exit closes it; the
// game is never sent a Reply.
func TestSpeedMenu(t *testing.T) {
	f := &fake{st: play.Fighting{Common: play.Common{Party: []play.Member{{Name: "ARIN"}},
		Ask: &play.Question{Replies: []string{"Guard", "Speed", "Exit"}, Do: []game.Decision{nil, game.SpeedMenu{}}}}}}
	r := newRules(f, game.Data{})
	r.Apply(game.Answer{N: 1})
	if a := shownOf(r).Pending; a == nil || len(a.Options) != 3 || a.Options[0].Label != "Faster" {
		t.Fatalf("menu %+v", a)
	}
	r.Apply(game.Answer{N: 1})
	r.Apply(game.Answer{N: 1})
	r.Apply(game.Answer{N: 0})
	if r.speed != 5 || len(f.sent) != 0 {
		t.Fatalf("speed %d sent %v", r.speed, f.sent)
	}
	r.Apply(game.Answer{N: 2})
	if r.speedUp {
		t.Fatal("still up")
	}
}

// TestVerbOrderWords: a VerbOrder in the Gold Box menus' words (Use, Delay,
// Done) places gobox's Items, Wait and End; naming Aim gives gobox's free Aim.
func TestVerbOrderWords(t *testing.T) {
	fight := play.Fighting{Common: play.Common{Party: []play.Member{{Name: "ARIN"}}, Allowed: allow(play.EndTurn{}, play.UseItem{}, play.Delay{}).Allowed},
		Field: play.Battlefield{Turn: 1, Units: []play.Unit{{ID: 1, Member: 0}}}}
	o := NewGame(&fake{st: fight}, game.Data{VerbOrder: []string{"Aim", "Done", "Delay", "Use"}}).View().Battle.Turn.Options
	var ls []string
	for _, x := range o {
		ls = append(ls, x.Label)
	}
	if !slices.Equal(ls, []string{"Aim", "End", "Wait", "Items"}) {
		t.Errorf("bar %v, want Aim End Wait Items", ls)
	}
}

// TestFreeAimInOrder: gobox's free Aim takes its VerbOrder place (COTAB's
// third), not the bar's end.
func TestFreeAimInOrder(t *testing.T) {
	fight := play.Fighting{Common: play.Common{Party: []play.Member{{Name: "ARIN"}}, Allowed: allow(play.EndTurn{}, play.UseItem{}).Allowed},
		Field: play.Battlefield{Turn: 1, Units: []play.Unit{{ID: 1, Member: 0}}}}
	o := NewGame(&fake{st: fight}, game.Data{VerbKeys: map[string]rune{"Aim": 'A'}, VerbOrder: []string{"Aim", "Use", "Done"}}).View().Battle.Turn.Options
	if len(o) == 0 || o[0].Label != "Aim" {
		t.Errorf("bar %+v, want Aim first", o)
	}
}

// TestPickViewMember: a pick whose Allowed offers play.ViewMember lets the
// player look first (Ask.View), and View goes to the rules, not the sheet.
func TestPickViewMember(t *testing.T) {
	st := play.Scene{Common: play.Common{Party: []play.Member{{Name: "ARIN"}}, Ask: &play.Question{Text: "Who?", Replies: []string{"ARIN"}, List: true}}}
	st.Allowed.Plain = []play.Action{play.Reply{N: 0}, play.ViewMember{Member: 0}}
	f := &fake{st: st}
	r := newRules(f, game.Data{})
	if a := shownOf(r).Pending; a == nil || !a.View {
		t.Fatalf("pick ask %+v: no View", a)
	}
	r.Apply(game.ViewMember{Member: 0})
	if len(f.sent) != 1 || f.sent[0] != play.Action(play.ViewMember{Member: 0}) || shownOf(r).SheetPage != nil {
		t.Fatalf("sent %v, sheet %v", f.sent, shownOf(r).SheetPage != nil)
	}
}
