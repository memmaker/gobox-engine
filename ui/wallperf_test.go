package ui

import (
	"testing"
	"testing/fstest"

	"gobox/asset"
)

// The 3D view's first frame decodes no texture up front: the pool series is
// hundreds of large PNGs (seconds of decode); View.Load decodes one when drawn.
func TestNewViewDecodesLazily(t *testing.T) {
	art = fstest.MapFS{}
	var p Presenter
	p.newView()
	if n := len(p.view.Textures); n != 0 {
		t.Fatalf("newView decoded %d textures up front", n)
	}
	if n := asset.WallNames(series); len(n) > 0 && p.view.Load(n[0]) == nil {
		t.Fatalf("Load(%q) is nil", n[0])
	}
}

func BenchmarkNewView(b *testing.B) {
	art = fstest.MapFS{}
	for b.Loop() {
		var p Presenter
		p.newView()
	}
}
