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

// On the party menu a key during the title's fade ends it and acts; a key
// with no row is ignored (it was sent as Enter).
func TestMainMenuKeys(t *testing.T) {
	f := filepath.Join(t.TempDir(), "title.png")
	w, _ := os.Create(f)
	png.Encode(w, image.NewRGBA(image.Rect(0, 0, 1, 1)))
	w.Close()
	p := &Presenter{main: true}
	p.shownView.Backdrop = &front.Picture{File: f}
	v := front.View{MainMenu: []front.ActKind{front.ActBegin, front.ActQuit}}
	p.Key('B', v)
	if p.mm.fade() < 1 || len(p.inv.acts) != 1 || p.inv.acts[0].Kind != front.ActBegin {
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

// creationGone is a game whose creation has ended.
type creationGone struct{ front.Game }

func (creationGone) CreationView() *front.Creation { return nil }
func (creationGone) Resets() int                   { return 0 }

// A creation the game ends itself (a save after Keep) gives the party
// menu back: the presenter's sheet does not stay over it, eating its keys.
func TestCreationEndedByGame(t *testing.T) {
	p := &Presenter{creating: true, eng: creationGone{}}
	v := p.own(front.View{MainMenu: []front.ActKind{front.ActCreate}})
	if p.creating || v.MainMenu == nil {
		t.Fatalf("creating %v, menu %v", p.creating, v.MainMenu)
	}
}

// A menu the game changes starts on its first row, not where the old one's highlight was.
func TestMainMenuNewRows(t *testing.T) {
	var m mainUI
	m.newRows([]string{"Load", "Party", "Quit"})
	m.sel = 2
	m.newRows([]string{"Begin", "Party", "Quit"})
	if m.sel != 0 {
		t.Fatalf("sel %d", m.sel)
	}
	m.sel, m.sub = 1, true // the submenu opened by the player keeps its own highlight
	m.newRows([]string{"Create", "Exit"})
	if m.sel != 1 {
		t.Fatalf("submenu sel %d", m.sel)
	}
}

// A card with no portrait shows the member's name, wrapped to its window.
func TestNameLines(t *testing.T) {
	if got := nameLines("Buck Rogers Junior", 6, 5); len(got) != 3 || got[0] != "Buck" || got[1] != "Rogers" || got[2] != "Junior" {
		t.Fatalf("%q", got)
	}
	if got := nameLines("Al Bo Cyrillic", 6, 1); len(got) != 1 || got[0] != "Al Bo" {
		t.Fatalf("%q", got)
	}
}

// A picture with no At fills the area (the picture window), not nothing.
func TestPictureNoAt(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for i := range src.Pix {
		src.Pix[i] = 255
	}
	m := image.NewRGBA(image.Rect(0, 0, 8, 8))
	drawPictures(m, []front.Picture{{Img: src}}, image.Pt(8, 8))
	if m.RGBAAt(4, 4).R != 255 {
		t.Fatal("the picture was not drawn")
	}
}
