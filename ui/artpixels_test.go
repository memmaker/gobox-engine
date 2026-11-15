package ui

import (
	"testing"

	"gobox/engine"
	"gobox/game"
	"gobox/importer"
	"gobox/pool"
)

// TestArtPixels: the cotab opening's hand, with F12's either art, is drawn
// from the game's pixels: the archive's decode stood in for them (its raw
// delta frames, the other art's colours).
func TestArtPixels(t *testing.T) {
	eng, err := pool.New("../data/cotab", "../data/cotab/sources")
	if err != nil {
		t.Skip(err)
	}
	fo, _ := importer.Read("../data/cotab")
	eng.SetStart(fo.Manifest.Start)
	eng.SetEnd(fo.Manifest.End)
	eng.Cfg.SaveDir = t.TempDir()
	seed := uint32(1)
	eng.FixedSeed = &seed
	data := eng.Data()
	f := NewFront(NewGame(eng, data), data)
	f.Start("")
	for i := range 400 {
		f.Frame(1280, 800)
		if _, ok := eng.State().(game.Assembling); ok && i > 50 {
			break
		}
		if i%40 == 0 {
			f.Key('\r')
		}
	}
	f.Key('q')
	for range 30 {
		f.Frame(1280, 800)
	}
	f.Key('b')
	for i := 0; i < 400; i++ {
		f.Frame(1280, 800)
		if i == 20 {
			f.Key('\r')
		}
		if x, ok := eng.State().(game.Scene); ok && x.Picture != nil && x.Picture.Block == 10 {
			break
		}
	}
	_, set := f.Arts()
	defer engine.DOSArt.Store(engine.DOSArt.Load())
	defer set(engine.DOSArt.Load())
	for _, dos := range []bool{false, true} {
		engine.DOSArt.Store(dos)
		set(dos)
		f.Frame(1280, 800)
		ps := f.P.pictures()
		if len(ps) == 0 || ps[0].Pix == nil {
			t.Fatalf("DOS %v: no picture with the game's pixels", dos)
		}
		own := ps[0]
		own.Img = nil
		if ps[0].Img == nil || ps[0].Img != pixImage(own, f.P.ink()) {
			t.Errorf("DOS %v: the picture not drawn from the game's pixels (the archive's, or none)", dos)
		}
	}
}
