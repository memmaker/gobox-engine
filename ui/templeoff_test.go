package ui

import (
	"testing"

	"gobox/game"
)

// templeState is a temple with these services up for one member.
func templeState(sv ...game.Service) game.Trading {
	return game.Trading{Kind: game.TradeTemple, Services: sv, Common: game.Common{Party: []game.Member{{Name: "Ann"}}}}
}

// TestTempleOff (Felix's ruling): a service that would not help the selected
// member is greyed out: its row has nothing to pick, for key, mouse and pad
// alike (ui.RowsOff); a needed one is picked.
func TestTempleOff(t *testing.T) {
	g, _ := playGame(templeState(game.Service{Name: "Cure Disease"}, game.Service{Name: "Cure Light Wounds", Needs: []int{0}}))
	p := &Presenter{}
	p.SetEngine(g)
	p.camp.screen = templeHeal
	_, rows := p.templeItems()
	if rows[0].Do != nil || rows[1].Do == nil {
		t.Errorf("rows picked: %v %v, want off, on", rows[0].Do != nil, rows[1].Do != nil)
	}
}

// TestTemplePick: a service picked asks for the members it would help:
// their cards are marked, another is no pick, each one picked is healed
// and the question stays up until Exit.
func TestTemplePick(t *testing.T) {
	st := templeState(game.Service{Name: "Cure Disease", Needs: []int{1}})
	st.Party = []game.Member{{Name: "Ann"}, {Name: "Bob"}}
	g, _ := playGame(st)
	p := &Presenter{}
	p.SetEngine(g)
	p.camp.screen = templeHeal
	if _, rows := p.templeItems(); !p.cardMarked(1) || p.cardMarked(0) {
		t.Fatalf("highlighted service marks %v %v, want Bob only", p.cardMarked(0), p.cardMarked(1))
	} else {
		rows[0].Do()
	}
	if p.q == nil || p.q.Ask.Prompt != "Cure Disease for whom?" || p.q.Cand != 1 {
		t.Fatalf("question %+v", p.q)
	}
	if p.q.Answer(Action{Kind: ActMember, N: 0}) || len(p.calls) != 0 {
		t.Errorf("Ann picked: %d calls", len(p.calls))
	}
	if p.q.Answer(Action{Kind: ActMember, N: 1}) || len(p.calls) != 1 {
		t.Errorf("Bob picked: %d calls, want 1, the question up", len(p.calls))
	}
	if !p.q.Answer(Action{Kind: ActClose}) {
		t.Error("Esc leaves it up")
	}
}
