package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

// timedSlot is a fake with one game save that tells when it was saved.
type timedSlot struct{ fake }

func (timedSlot) Slots() []game.SaveSlot {
	return []game.SaveSlot{{Name: "GAME", At: time.Date(2026, 11, 13, 20, 13, 0, 0, time.Local)}}
}
func (timedSlot) LoadSlot(int) error { return nil }

// TestSlotDatesAlike: gobox's own slot files and the game's slots show
// their time of saving in the one format (game.SlotTime).
func TestSlotDatesAlike(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mine.sav"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	g := &classicGame{Game: NewGame(&timedSlot{fake{st: game.Exploring{}}}, game.Data{SaveDir: dir})}
	p := &Presenter{slots: &SlotsUI{LoadOnly: true}}
	p.SetEngine(g)
	sl := p.slotsView()
	if len(sl.Times) != 2 || sl.Times[1] != "2026-11-13 20:13" {
		t.Fatalf("times %q, want the game's as 2026-11-13 20:13", sl.Times)
	}
	for _, s := range sl.Times {
		if _, err := time.Parse(game.SlotTime, s); err != nil {
			t.Errorf("%q is not in the slot list's format: %v", s, err)
		}
	}
}
