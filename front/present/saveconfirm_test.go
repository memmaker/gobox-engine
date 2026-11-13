package present

import "testing"

// TestSaveAsksNameAndConfirm: every game's New Slot asks a name, and saving
// over a slot asks first; neither saves at once.
func TestSaveAsksNameAndConfirm(t *testing.T) {
	g, _ := slotsGame(t)
	p := &Presenter{slots: &SlotsUI{}}
	p.SetEngine(g)
	_, rows := p.screenItems()
	rows[0].Do() // New Slot
	rows[1].Do() // a save
	if a := p.inv.acts; len(a) != 0 {
		t.Fatalf("sent %+v, want a name asked and an overwrite confirm first", a)
	}
}
