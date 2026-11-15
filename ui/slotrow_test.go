package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gobox/game"
)

// longSlot is a fake with one game save of a long name and date.
type longSlot struct{ fake }

func (longSlot) Slots() []game.SaveSlot {
	return []game.SaveSlot{{Name: "A LONG SAVE NAME HERE", Time: "NOV 6 2026 12:34:56", Level: 5}}
}

func (longSlot) LoadSlot(int) error { return nil }

func (longSlot) SaveLevel(save []byte) (int, error) { return len(save), nil } // a save of n bytes: level n

// TestSlotRowLevelDate: a slot row's level stands under the "Level"
// heading, right aligned with it, a space before the date that ends the
// row (COTAB's "5NOV 6 2026"), and the row keeps its width; a slot file of
// gobox's own shows the level the game reads from it.
func TestSlotRowLevelDate(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mine.sav"), []byte("1234567"), 0o644); err != nil {
		t.Fatal(err)
	}
	f := &longSlot{fake{st: game.Exploring{}}}
	g := &classicGame{Game: NewGame(f, game.Data{SaveDir: dir})}
	p := &Presenter{slots: &SlotsUI{LoadOnly: true}}
	p.SetEngine(g)
	_, rows := p.screenItems()
	sl := p.slotsView()
	w, lc := slotCols(40, sl.Times)
	for i, lv := range map[int]string{0: "7", 1: "5"} { // gobox's own slot first, then the game's
		r := rows[i].Text
		if len(r) != w || strings.TrimSpace(r[lc:lc+slotLevelW]) != lv || r[lc+slotLevelW-1:lc+slotLevelW] != lv || r[lc+slotLevelW] != ' ' {
			t.Errorf("row %d %q (%d wide): want level %s right aligned in columns %d..%d, a space after", i, r, len(r), lv, lc, lc+slotLevelW-1)
		}
		if !strings.HasSuffix(r, sl.Times[i]) {
			t.Errorf("row %d %q does not end with its date %q", i, r, sl.Times[i])
		}
	}
}
