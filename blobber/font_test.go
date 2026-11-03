package blobber

import "testing"

// TestFontPage: the Graphics page's Font line opens the list, Esc goes back to it.
func TestFontPage(t *testing.T) {
	g := &Game{page: "Graphics"}
	if l := g.menuOptions()[graphicsFont]; l.label != "Font" {
		t.Fatalf("Graphics row %d is %q, not Font", graphicsFont, l.label)
	}
	g.menuOptions()[graphicsFont].change(g, 0)
	if g.page != "Font" || len(g.menuOptions()) != len(FontNames()) {
		t.Fatalf("page %q", g.page)
	}
	if !g.back() || g.page != "Graphics" || g.sel != graphicsFont {
		t.Fatalf("back: page %q sel %d", g.page, g.sel)
	}
}
