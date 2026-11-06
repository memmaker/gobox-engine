package present

import (
	"strings"
	"testing"

	"gobox/front"
)

// TestSlotRowLevelDate: a slot row's level and a long date keep a space
// between them (COTAB's "5NOV 6 2026"), and the row keeps its width.
func TestSlotRowLevelDate(t *testing.T) {
	v := front.View{Slots: &front.Slots{LoadOnly: true, Names: []string{"A LONG SAVE NAME HERE"}, Times: []string{"NOV 6 2026 12:34:56"}, Levels: []int{5}}}
	_, rows := (&Presenter{}).screenItems(v)
	r := rows[0].Text
	if !strings.Contains(r, " 5 NOV 6 2026") || len(r) != 34 {
		t.Fatalf("row %q (%d)", r, len(r))
	}
}
