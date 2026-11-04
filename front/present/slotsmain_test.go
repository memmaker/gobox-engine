package present

import (
	"testing"

	"gobox/front"
	"gobox/gpu"
)

// slotsGame is the engine as the slot page asks it: two saves.
type slotsGame struct {
	front.Game
	loaded string
}

func (g *slotsGame) SlotsView(loadOnly bool) *front.Slots {
	return &front.Slots{Names: []string{"A", "B"}, Times: []string{"", ""}, Levels: []int{1, 1}, LoadOnly: loadOnly}
}
func (g *slotsGame) Resets() int { return 0 }
func (g *slotsGame) Load(name string, loadOnly bool) bool {
	g.loaded = name
	return true
}

// TestSlotsFromMainMenu: the main menu's Load opens the slot page, where
// Down moves the highlight and Enter loads the save under it.
func TestSlotsFromMainMenu(t *testing.T) {
	g := &slotsGame{}
	p := &Presenter{}
	p.SetEngine(g)
	dst := gpu.NewImage(640, 400)
	defer dst.Deallocate()
	note := Note{Glyph: func(byte) [8]byte { return [8]byte{} }, Font: func(byte, bool) [8]byte { return [8]byte{} }}
	v := front.View{MainMenu: []front.ActKind{front.ActLoad, front.ActQuit}}
	p.Draw(dst, note, v)
	p.send(front.ActLoad, 0)
	if p.slots == nil {
		t.Fatal("Load did not open the slot page")
	}
	p.Draw(dst, note, v)
	sel := p.scr.sel
	p.Key(0x5000, v) // Down
	p.Draw(dst, note, v)
	if p.scr.sel == sel {
		t.Fatalf("Down left the highlight on %d", sel)
	}
	p.Key(13, v)
	for _, f := range p.Calls() {
		f(g)
	}
	if g.loaded != "B" {
		t.Fatalf("Enter loaded %q, want B", g.loaded)
	}
}
