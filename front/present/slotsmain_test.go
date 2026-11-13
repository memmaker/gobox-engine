package present

import (
	"fmt"
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
	"gobox/gpu"
)

// numbered is a fake whose saves are numbered: Load's data says which one.
type numbered struct {
	*fake
	n int
}

func (f *numbered) Save() ([]byte, error) { f.n++; return []byte(fmt.Sprint("save", f.n)), nil }

// slotsGame is the presenter's game with two saves, A and B.
func slotsGame(t *testing.T) (*classicGame, *numbered) {
	f := &numbered{fake: &fake{st: play.Exploring{Common: play.Common{Party: []game.Member{{Name: "A"}}}}}}
	g := &classicGame{Game: NewGame(f, game.Data{SaveDir: t.TempDir()})}
	for _, n := range []string{"A", "B"} {
		if err := g.Save(n); err != nil {
			t.Fatal(err)
		}
	}
	return g, f
}

// TestSlotsFromMainMenu: the main menu's Load opens the slot page, where
// Down moves the highlight and Enter loads the save under it.
func TestSlotsFromMainMenu(t *testing.T) {
	g, f := slotsGame(t)
	p := &Presenter{}
	p.SetEngine(g)
	dst := gpu.NewImage(640, 400)
	defer dst.Deallocate()
	note := Note{Glyph: func(byte) [8]byte { return [8]byte{} }, Font: func(byte, bool) [8]byte { return [8]byte{} }}
	p.Draw(dst, note)
	p.send(front.ActLoad, 0)
	if p.slots == nil {
		t.Fatal("Load did not open the slot page")
	}
	p.Draw(dst, note)
	sel := p.scr.sel
	p.Key(0x5000) // Down
	p.Draw(dst, note)
	if p.scr.sel == sel {
		t.Fatalf("Down left the highlight on %d", sel)
	}
	p.Key(13)
	for _, c := range p.Calls() {
		c(g)
	}
	second := g.SlotsView(true).Names[1] // the row under the highlight after Down
	if want := map[string]string{"A": "save1", "B": "save2"}[second]; string(f.data) != want {
		t.Fatalf("Enter loaded %q, want %s's %s", f.data, second, want)
	}
}
