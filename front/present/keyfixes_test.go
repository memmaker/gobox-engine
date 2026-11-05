package present

import (
	"gobox/front"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// On the party menu the first key only ends the title's fade (used up);
// the next acts; a key with no row is ignored (it was sent as Enter).
func TestMainMenuKeys(t *testing.T) {
	f := filepath.Join(t.TempDir(), "title.png")
	w, _ := os.Create(f)
	png.Encode(w, image.NewRGBA(image.Rect(0, 0, 1, 1)))
	w.Close()
	p := &Presenter{main: true}
	p.shownView.Backdrop = &front.Picture{File: f}
	v := front.View{MainMenu: []front.ActKind{front.ActBegin, front.ActQuit}}
	p.Key('B', v)
	if p.mm.fade() < 1 || len(p.inv.acts) != 0 {
		t.Fatalf("the key ending the fade fell through: fade %v, acts %v", p.mm.fade(), p.inv.acts)
	}
	p.Key('B', v)
	if len(p.inv.acts) != 1 || p.inv.acts[0].Kind != front.ActBegin {
		t.Fatalf("first key: fade %v, acts %v", p.mm.fade(), p.inv.acts)
	}
	p.inv.acts = nil
	p.Key('x', v)
	if len(p.inv.acts) != 0 {
		t.Fatalf("an unknown key sent %v", p.inv.acts)
	}
}

// SOTSB: Enter during the intro, before the main menu shows. The Enter
// that skips the menu's fade-in is used up, and a menu not yet drawn takes
// no key at all: Load Game (the first row) is not chosen.
func TestIntroEnterNoFallThrough(t *testing.T) {
	f := filepath.Join(t.TempDir(), "title.png")
	w, _ := os.Create(f)
	png.Encode(w, image.NewRGBA(image.Rect(0, 0, 1, 1)))
	w.Close()
	bd := &front.Picture{File: f}
	v := front.View{MainMenu: []front.ActKind{front.ActLoad, front.ActBegin}, Backdrop: bd}
	p := &Presenter{}
	p.Key('\r', v) // the game's View is the menu already, the intro still on screen
	if p.slots != nil || len(p.inv.acts) != 0 {
		t.Fatalf("a key reached the undrawn menu: slots %v, acts %v", p.slots != nil, p.inv.acts)
	}
	p.main, p.shownView = true, v // drawn, fading in
	p.Key('\r', v)
	if p.mm.fade() < 1 || p.slots != nil || len(p.inv.acts) != 0 {
		t.Fatalf("the Enter ending the fade picked: fade %v, slots %v, acts %v", p.mm.fade(), p.slots != nil, p.inv.acts)
	}
	p.mm.at = time.Now() // a click during the fade: used up too
	if !p.Click(0, 0) || p.mm.fade() < 1 || p.slots != nil || len(p.inv.acts) != 0 {
		t.Fatal("the click ending the fade picked")
	}
}
