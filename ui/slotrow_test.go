package ui

import (
	"strings"
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// longSlot is a fake with one game save of a long name and date.
type longSlot struct{ fake }

func (longSlot) Slots() []game.SaveSlot {
	return []game.SaveSlot{{Name: "A LONG SAVE NAME HERE", Time: "NOV 6 2026 12:34:56", Level: 5}}
}

func (longSlot) LoadSlot(int) error { return nil }

// TestSlotRowLevelDate: a slot row's level and a long date keep a space
// between them (COTAB's "5NOV 6 2026"), and the row keeps its width.
func TestSlotRowLevelDate(t *testing.T) {
	f := &longSlot{fake{st: play.Exploring{}}}
	g := &classicGame{Game: NewGame(f, game.Data{SaveDir: t.TempDir()})}
	p := &Presenter{slots: &SlotsUI{LoadOnly: true}}
	p.SetEngine(g)
	_, rows := p.screenItems()
	r := rows[0].Text
	if !strings.Contains(r, " 5 NOV 6 2026") || len(r) != 34 {
		t.Fatalf("row %q (%d)", r, len(r))
	}
}
