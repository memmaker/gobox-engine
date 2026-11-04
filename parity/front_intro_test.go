package parity

import (
	"os"
	"path/filepath"
	"testing"

	"gobox/front"
	"gobox/front/classic"
	"gobox/front/present"
	"por/rules"
)

// TestFrontIntro: por as cmd/por wires it (present.NewFront over
// classic.NewGame, nobody starting the engine but the front's Open) plays
// the title, TITLE pictures 1 and 2 then the credits, before the party
// menu, and the window border's pieces are in the View.
func TestFrontIntro(t *testing.T) {
	dir := filepath.Join(porDir(t), "data")
	if _, err := os.Stat(filepath.Join(dir, "dos", "TITLE.DAX")); err != nil {
		t.Skip("no game files")
	}
	e, err := rules.New(dir)
	if err != nil {
		t.Fatal(err)
	}
	e.Cfg.SaveDir, e.Cfg.FullIntro = t.TempDir(), true
	f := present.NewFront(classic.NewGame(e.Rules(), e.Data()), e.Data())
	f.Act()
	var in front.IntroUI
	var pics []int
	credits, menu := false, false
	for i := 0; i < 20 && !menu; i++ {
		v := f.Game.View()
		if v.Intro != nil && v.Full && len(v.Pictures) > 0 && (len(pics) == 0 || pics[len(pics)-1] != v.Pictures[0].Block) {
			pics = append(pics, v.Pictures[0].Block)
		}
		credits = credits || v.Credits
		menu = v.MainMenu != nil && v.Intro == nil
		if menu && v.BorderPix[0].Pix == nil {
			t.Error("party menu: no border pieces in the View")
		}
		front.IntroStep(&f.P, &in, v.Intro, i%2 == 1) // a key every other frame cuts each phase
		f.Act()
	}
	if len(pics) != 2 || pics[0] != 1 || pics[1] != 2 || !credits || !menu {
		t.Errorf("title pictures %v, credits %v, party menu %v; want [1 2], credits, menu", pics, credits, menu)
	}
}
