package present

import (
	"slices"
	"testing"

	"gobox/game"
	"gobox/game/play"
)

func TestJournalQuest(t *testing.T) {
	q := game.Quest{ID: "Rats", Giver: "Bram", Place: "Inn", Brief: []string{"a", "b"}, Done: true}
	want := []inked{{"Given by Bram", jInk}, {"Inn", jFaded}, {"", jInk}, {"a", jInk}, {"", jInk}, {"b", jInk}, {"", jInk}}
	if got := questLines(q, 40); !slices.Equal(got, want) {
		t.Errorf("quest page %+v, want %+v", got, want)
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
