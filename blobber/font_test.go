package blobber

import (
	"fmt"
	"testing"
)

// TestFontPage: the Graphics page's Font line opens the list, Esc goes back to it.
func TestFontPage(t *testing.T) {
	g := &Game{page: "Graphics"}
	graphicsFont := g.fontRow()
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

// TestExtraRows: host rows go before the built-in row At, a new page is on the root and its rows get a way back.
func TestExtraRows(t *testing.T) {
	defer func(o map[string][]Row) { ExtraRows = o }(ExtraRows)
	hit := 0
	ExtraRows = map[string][]Row{
		"Graphics": {{Label: "A", At: 1}, {Label: "B", At: 1}, {Label: "Z", At: 4}},
		"Sound":    {{Label: "Vol", Get: func() string { return "7" }, Change: func(d int) { hit += d + 5 }}},
	}
	labels := func(g *Game) (o []string) {
		for _, l := range g.menuOptions() {
			o = append(o, l.label)
		}
		return
	}
	if got := fmt.Sprint(labels(&Game{page: "Graphics"})); got != "[Fullscreen A B Dithered darkness Torch intensity Font Z Exit]" {
		t.Fatal(got)
	}
	if got := fmt.Sprint(labels(&Game{})); got != "[Resume Gameplay Graphics Sound Quit to DOS]" {
		t.Fatal(got)
	}
	g := &Game{page: "Sound"}
	o := g.menuOptions()
	o[0].change(g, 1)
	if fmt.Sprint(labels(g)) != "[Vol Exit]" || o[0].value(g) != "7" || hit != 6 {
		t.Fatal(labels(g), hit)
	}
}
