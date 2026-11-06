package present

import (
	"testing"

	"gobox/front"
)

// TestSaveNoConfirmUnnamed: a game naming its own slots saves a new one
// with no name asked, and over one with no confirm.
func TestSaveNoConfirmUnnamed(t *testing.T) {
	saveNoConfirm, saveUnnamed = true, true
	defer func() { saveNoConfirm, saveUnnamed = false, false }()
	v := front.View{Slots: &front.Slots{CanSave: true, Names: []string{"A"}, Times: []string{""}, Levels: []int{0}}}
	p := &Presenter{}
	_, rows := p.screenItems(v)
	rows[0].Do() // New Slot
	rows[1].Do() // A
	if a := p.inv.acts; len(a) != 2 || a[0].Kind != front.ActSaveSlot || a[0].Text != "" || a[1].Kind != front.ActSaveSlot || a[1].Text != "A" {
		t.Fatalf("sent %+v, want saves to the next slot and to A", a)
	}
}
