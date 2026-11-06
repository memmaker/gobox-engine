package present

import (
	"testing"

	"gobox/front"
)

// TestSaveAsksNameAndConfirm: every game's New Slot asks a name, and saving
// over a slot asks first; neither saves at once.
func TestSaveAsksNameAndConfirm(t *testing.T) {
	v := front.View{Slots: &front.Slots{CanSave: true, Names: []string{"A"}, Times: []string{""}, Levels: []int{0}}}
	p := &Presenter{}
	_, rows := p.screenItems(v)
	rows[0].Do() // New Slot
	rows[1].Do() // A
	if a := p.inv.acts; len(a) != 0 {
		t.Fatalf("sent %+v, want a name asked and an overwrite confirm first", a)
	}
}
