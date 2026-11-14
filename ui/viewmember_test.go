package ui

import (
	"slices"
	"testing"

	"gobox/game"
)

// TestViewMemberOpensSheet: View (a shop's, temple's) is front state: the
// member's sheet comes up with no call to the game, and CloseSheet ends it.
func TestViewMemberOpensSheet(t *testing.T) {
	f := &fake{st: game.Exploring{Common: game.Common{Party: []game.Member{{Name: "ARIN"}}}}}
	r := newRules(f, game.Data{})
	r.send(game.ViewMember{Member: 0})
	if shownOf(r).SheetPage == nil || len(f.sent) != 0 {
		t.Fatalf("sheet %v, sent %v", shownOf(r).SheetPage != nil, f.sent)
	}
	r.send(tCloseSheet{})
	if shownOf(r).SheetPage != nil {
		t.Fatal("sheet still up")
	}
}

// Camp with no AdvanceTimeBy allowed offers no timed rest (Camp.NoRest).
func TestCampNoRest(t *testing.T) {
	for _, allow := range []bool{false, true} {
		st := game.Camping{}
		if allow {
			st.Allowed.Plain = []game.Action{game.AdvanceTimeBy{}}
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
	fight := game.Fighting{Common: game.Common{Party: []game.Member{{Name: "ARIN"}, {Name: "MAREN"}}, Allowed: allow(game.EndTurn{}).Allowed},
		Field: game.Battlefield{Turn: 2, Units: []game.Unit{{ID: 1, Member: 0}, {ID: 2, Member: 1}}}}
	f := &fake{st: fight}
	if g0 := NewGame(f, game.Data{}); slices.ContainsFunc(turnOptions(g0.combat(), g0.D), func(o game.Option) bool { return o.Label == "View" }) {
		t.Fatal("View the series doesn't key")
	}
	g := NewGame(f, game.Data{VerbKeys: map[string]rune{"View": 'V'}})
	n := -1
	for i, o := range turnOptions(g.combat(), g.D) {
		if o.Label == "View" {
			n = i
		}
	}
	if n < 0 {
		t.Fatalf("no View in %+v", turnOptions(g.combat(), g.D))
	}
	if o := turnOptsOf(NewGame(f, game.Data{VerbKeys: map[string]rune{"View": 'V'}, VerbOrder: []string{"View", "End"}})); o[0].Label != "View" {
		t.Errorf("VerbOrder: %+v, want View first", o)
	}
	g.Answer(Action{Kind: ActOption, N: n})
	p := presenterOf(g)
	if p.sheetPage() == nil || !p.sheetShown() || g.combat() == nil || g.combat().Turn == nil || len(f.sent) != 0 {
		t.Fatalf("sheet %v, battle %v, sent %v", p.sheetPage() != nil, g.combat() != nil, f.sent)
	}
	g.CloseSheet()
	if p.sheetPage() != nil || g.combat() == nil || g.combat().Turn == nil || g.combat().Turn.Unit != 2 {
		t.Fatal("not back on the turn")
	}
	f.st = game.Fighting{Common: game.Common{Party: fight.Party, Allowed: allow(game.ViewTarget{}).Allowed}, Field: fight.Field}
	views := 0
	for _, o := range turnOptions(g.combat(), g.D) {
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
	fight := game.Fighting{Common: game.Common{Party: []game.Member{{Name: "ARIN", Page: &game.SheetPage{Name: "ARIN", Damage: "1d8", Sex: "Male", Age: 20}}}}}
	g := NewGame(&fake{st: fight}, game.Data{})
	g.OpenSheet(0)
	if sp := presenterOf(g).sheetPage(); sp == nil || sp.Damage != "1d8" || sp.Sex != "Male" || sp.Age != 20 {
		t.Fatalf("page %+v", sp)
	}
}

// TestFightItems: Gear in a fight opens the turn member's items over the
// fight, with no call to the game; closing them returns to the same turn.
func TestFightItems(t *testing.T) {
	f := &fake{st: game.Fighting{Common: game.Common{Selected: 0, Party: []game.Member{{Name: "ARIN"}, {Name: "MAREN"}}},
		Field: game.Battlefield{Turn: 9, Units: []game.Unit{{ID: 9, Member: 1}}}}}
	r := newRules(f, game.Data{})
	r.send(game.OpenItems{Member: -1})
	if s := shownOf(r); s.Items != game.ItemsOwn || s.Combat == nil || len(f.sent) != 0 || s.Selected != 1 {
		t.Fatalf("items %v combat %v sent %v", s.Items, s.Combat != nil, f.sent)
	}
	r.send(tCloseItems{})
	if s := shownOf(r); s.Items != game.ItemsNone || s.SheetPage != nil || s.Combat == nil {
		t.Fatalf("after close: items %v sheet %v", s.Items, s.SheetPage != nil)
	}
}

// TestViewGearOpensItems: a View menu's Gear the rules give as a
// front-state reply (Question.Do) opens that member's items, not a Reply.
func TestViewGearOpensItems(t *testing.T) {
	f := &fake{st: game.Fighting{Common: game.Common{Party: []game.Member{{Name: "ARIN"}, {Name: "MAREN"}, {Name: "VEX"}},
		Ask: &game.Question{Replies: []string{"Gear", "Exit"}, Do: []game.Action{game.OpenItems{Member: 2}}}},
		Field: game.Battlefield{Turn: 9, Units: []game.Unit{{ID: 9, Member: 1}}}}}
	r := newRules(f, game.Data{})
	r.send(tAnswer{N: 0})
	if s := shownOf(r); s.Items != game.ItemsOwn || s.Selected != 2 || len(f.sent) != 0 || s.Pending != nil {
		t.Fatalf("items %v selected %d sent %v pending %v", s.Items, s.Selected, f.sent, s.Pending)
	}
	g := NewGame(f, game.Data{})
	g.r = r
	if p := presenterOf(g); !p.inventoryUp() || p.askUp() != nil { // the items draw, over the question
		t.Fatalf("inventory %v ask %v", presenterOf(g).inventoryUp(), presenterOf(g).askUp() != nil)
	}
	r.send(tCloseItems{})
	if shownOf(r).Pending == nil {
		t.Fatal("the View menu is gone")
	}
}

// TestSpeedMenu: a Done menu's Speed (Question.Do SpeedMenu) is gobox's
// menu: Slower and Faster step the one speed (0..9), Exit closes it; the
// game is never sent a Reply.
func TestSpeedMenu(t *testing.T) {
	f := &fake{st: game.Fighting{Common: game.Common{Party: []game.Member{{Name: "ARIN"}},
		Ask: &game.Question{Replies: []string{"Guard", "Speed", "Exit"}, Do: []game.Action{nil, game.SpeedMenu{}}}}}}
	r := newRules(f, game.Data{})
	r.send(tAnswer{N: 1})
	if a := shownOf(r).Pending; a == nil || len(a.Options) != 3 || a.Options[0].Label != "Faster" {
		t.Fatalf("menu %+v", a)
	}
	r.send(tAnswer{N: 1})
	r.send(tAnswer{N: 1})
	r.send(tAnswer{N: 0})
	if r.speed != 5 || len(f.sent) != 0 {
		t.Fatalf("speed %d sent %v", r.speed, f.sent)
	}
	r.send(tAnswer{N: 2})
	if r.speedUp {
		t.Fatal("still up")
	}
}

// TestVerbOrderWords: a VerbOrder in the Gold Box menus' words (Use, Delay,
// Done) places gobox's Items, Wait and End; an Aim in it adds nothing (the
// free-cursor Aim is gone).
func TestVerbOrderWords(t *testing.T) {
	fight := game.Fighting{Common: game.Common{Party: []game.Member{{Name: "ARIN"}}, Allowed: allow(game.EndTurn{}, game.UseItem{}, game.Delay{}).Allowed},
		Field: game.Battlefield{Turn: 1, Units: []game.Unit{{ID: 1, Member: 0}}}}
	o := turnOptsOf(NewGame(&fake{st: fight}, game.Data{VerbOrder: []string{"Aim", "Done", "Delay", "Use"}}))
	var ls []string
	for _, x := range o {
		ls = append(ls, x.Label)
	}
	if !slices.Equal(ls, []string{"End", "Wait", "Items"}) {
		t.Errorf("bar %v, want End Wait Items", ls)
	}
}

// TestPickViewMember: a pick whose Allowed offers game.ViewMember lets the
// player look first (Ask.View), and View goes to the rules, not the sheet.
func TestPickViewMember(t *testing.T) {
	st := game.Scene{Common: game.Common{Party: []game.Member{{Name: "ARIN"}}, Ask: &game.Question{Text: "Who?", Replies: []string{"ARIN"}, List: true}}}
	st.Allowed.Plain = []game.Action{game.Reply{N: 0}, game.ViewMember{Member: 0}}
	f := &fake{st: st}
	r := newRules(f, game.Data{})
	if a := shownOf(r).Pending; a == nil || !a.View {
		t.Fatalf("pick ask %+v: no View", a)
	}
	r.send(game.ViewMember{Member: 0})
	if len(f.sent) != 1 || f.sent[0] != game.Action(game.ViewMember{Member: 0}) || shownOf(r).SheetPage != nil {
		t.Fatalf("sent %v, sheet %v", f.sent, shownOf(r).SheetPage != nil)
	}
}

// turnOptsOf is g's fight's turn options as the bar offers them.
func turnOptsOf(g *Game) []game.Option { return turnOptions(g.combat(), g.D) }
