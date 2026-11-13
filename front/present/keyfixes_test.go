package present

import (
	"gobox/front"
	"gobox/game/play"
	"gobox/ui"
	"image"
	"slices"
	"testing"
	"time"
)

// On the party menu a key during the title's fade ends it and acts; a key
// with no row is ignored (it was sent as Enter).
func TestMainMenuKeys(t *testing.T) {
	p := menuOn("Begin Adventuring", "Exit to DOS")
	p.main = true
	p.mm.title = image.NewRGBA(image.Rect(0, 0, 1, 1))
	v := front.View{}
	p.Key('B', v)
	if p.mm.fade() < 1 || len(p.Calls()) != 1 { // Begin is a rule call (StateAct)
		t.Fatalf("first key: fade %v, acts %v", p.mm.fade(), p.inv.acts)
	}
	p.Key('x', v)
	if len(p.inv.acts) != 0 || len(p.Calls()) != 0 {
		t.Fatalf("an unknown key sent %v", p.inv.acts)
	}
}

// SOTSB: Enter during the intro, before the main menu shows. The Enter
// that skips the menu's fade-in is used up, and a menu not yet drawn takes
// no key at all: Load Game (the first row) is not chosen.
func TestIntroEnterNoFallThrough(t *testing.T) {
	v := front.View{}
	p := menuOn("Load Saved Game", "Begin Adventuring")
	p.mm.title = image.NewRGBA(image.Rect(0, 0, 1, 1))
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

// A creation the game ends itself (a save after Keep) gives the party
// menu back: the presenter's sheet does not stay over it, eating its keys.
func TestCreationEndedByGame(t *testing.T) {
	gone, _ := playGame(play.Assembling{}) // no draft: the creation has ended
	p := &Presenter{creating: true, eng: gone}
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

// A typed line reaches the game whole, once, on Enter: never a line half typed.
func TestTypedLineOnce(t *testing.T) {
	p := &Presenter{}
	v := front.View{Ask: &front.Ask{Kind: front.AskText, Max: 15}}
	for _, k := range []uint16{'B', 'o', 'b', '\r'} {
		p.keyAct(k, v)
	}
	if len(p.inv.acts) != 1 || p.inv.acts[0].Kind != front.ActText || p.inv.acts[0].Text != "Bob" {
		t.Fatalf("acts %+v", p.inv.acts)
	}
}

// Camp offers Save beside the rests, as the original's camp does.
func TestCampSave(t *testing.T) {
	p := campPresenter()
	_, it := p.campItems(front.View{Camp: &front.Camp{}})
	if !slices.ContainsFunc(it, func(r ui.MenuRow) bool { return r.Text == "Save" }) {
		t.Fatal("no Save in camp")
	}
}
