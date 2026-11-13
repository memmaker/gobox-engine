package present

import (
	"slices"
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

func TestJournalAndPowers(t *testing.T) {
	s := shown{Quests: []game.Quest{{ID: "Rats", Giver: "Bram", Place: "Inn", Brief: []string{"a", "b"}, Done: true}},
		PowerPick: &game.PowerPick{Name: "Ann", What: "Surge", Options: []game.PowerOption{{ID: 3, Name: "Smite", Aim: 2, Range: 4}}},
		Holding:   game.HoldDoor, Door: &game.Door{Locked: true, Pick: true}}
	v := Waits(front.View{}, s.Door, s.PowerPick)
	if !v.Door.Pick || v.Door.Bash || v.Powers.Options[0].Aim != front.AimFoe || v.Powers.Options[0].Range != 4 {
		t.Errorf("waits %+v %+v", v.Door, v.Powers)
	}
}

// The journal lists what the state holds: its quests (open first) and its notes.
func TestJournalReadsState(t *testing.T) {
	g, _ := playGame(play.Exploring{Common: play.Common{
		Quests: []game.Quest{{ID: "Done", Done: true}, {ID: "Open"}},
		Lore:   []game.Lore{{Title: "Note", ID: 7}}}})
	p := &Presenter{}
	p.SetEngine(g)
	p.openJournal()
	if want := []int{-1, 1, -1, 0, -2}; !slices.Equal(p.journal.rows, want) {
		t.Errorf("quest rows %v, want %v", p.journal.rows, want)
	}
	p.journal.setTab(1, p.common())
	p.journal.sel = 0
	if id := p.journal.removable(p.common()); id != 7 {
		t.Errorf("removable note %d, want 7", id)
	}
}
