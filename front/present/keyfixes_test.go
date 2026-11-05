package present

import (
	"gobox/front"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// On the party menu the first key ends the title's fade and acts on the
// menu too; a key with no row is ignored (it was sent as Enter).
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
