package present

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gobox/dax"
	"gobox/front"
	"gobox/game"
	"gobox/input"
	"gobox/ui/fonts"
)

func testFight() *game.SpaceFight {
	return &game.SpaceFight{
		Foe:    game.Picture{File: "SHIPS", Block: 40},
		Status: "Closing    ", Range: 3,
		Own:     game.ShipSystems{Now: [6]int{100, 0, 30, 40, 50, 60}, Max: [6]int{200, 40, 30, 40, 50, 60}},
		Enemy:   game.ShipSystems{Now: [6]int{0, 20, 30, 40, 50, 60}, Max: [6]int{80, 40, 30, 40, 50, 60}},
		Weapons: []game.WeaponSlot{{Short: "LASER", Mounts: 2, Working: 2}, {Short: "MISSILE", Mounts: 2, Working: 1}, {}},
		Log: []game.Message{
			{Tone: game.ToneOwn, Text: "BUCK fires lasers and hits for 12 points of damage on the pirate ship!"},
			{Tone: game.ToneWarn, Text: "The pirate fires missiles and misses."},
			{Tone: game.ToneBad, Text: "WILMA is dead!!"},
			{Tone: game.TonePrompt, Text: "what does BUCK want to do?"},
		},
	}
}

func crewMenu() *front.Ask {
	a := &front.Ask{Kind: front.AskMenu, Break: 5}
	for _, s := range []string{"Fire", "Target", "Load", "Quit", "View", "Command", "Aid", "Sensors"} {
		a.Items = append(a.Items, front.MenuItem{Text: s, Key: s[0]})
	}
	return a
}

func fightPage(p *Presenter, v front.View) *cellPage {
	c := newCellPage(v)
	c.spaceFight(sfOf(v), p.sfState(v, sfOf(v)), v.Ask)
	return c
}

func at(t *testing.T, c *cellPage, x, y int, s string, fg, bg int) {
	t.Helper()
	for i := 0; i < len(s); i++ {
		k := c.cells[y][x+i]
		if !k.set || k.ch != s[i] || k.fg != egaInk[fg] || k.bg != egaInk[bg] {
			t.Fatalf("(%d,%d) %q: got %q set %v fg %v bg %v, want %q fg%d bg%d", x+i, y, s, k.ch, k.set, k.fg, k.bg, s[i], fg, bg)
		}
	}
}

func TestSpaceFightLayout(t *testing.T) {
	p := &Presenter{}
	v := front.View{Screen: &game.Screen{Kind: game.ScreenSpaceFight, SpaceFight: testFight()}, Ask: crewMenu()}
	c := fightPage(p, v)
	at(t, c, 1, 1, "Closing    ", 12, 0)
	at(t, c, 1, 2, "Range: 3", 12, 0)
	at(t, c, 29, 1, "Hull:  100", 11, 8)
	at(t, c, 29, 3, "Sens:    0", 12, 8) // zero: red
	at(t, c, 10, 12, "    0", 12, 0)
	at(t, c, 10, 13, "   20", 10, 0)
	at(t, c, 29, 14, "Weapons:", 10, 8)
	at(t, c, 29, 15, "LASER", 10, 8)
	at(t, c, 37, 15, " 2", 10, 8)
	at(t, c, 29, 16, "MISSILE", 14, 8) // a mount down
	at(t, c, 37, 16, " 1", 14, 8)
	// the first log line wraps at the last space before column 37: 3 rows + 2 + the prompt padded = 2 pages
	sh := p.sfState(v, sfOf(v))
	if len(sh.Pages) != 2 || !sh.More {
		t.Fatalf("pages %d more %v, want 2 and a pause", len(sh.Pages), sh.More)
	}
	at(t, c, 1, 19, "BUCK fires lasers and hits for 12", 11, 0)
	at(t, c, 1, 20, "points of damage on the pirate ship!", 11, 0)
	at(t, c, 1, 21, "The pirate fires missiles and misses.", 14, 0)
	at(t, c, 0, 24, sfPause, 15, 0)
	if c.cells[23][0].set && c.cells[23][0].ch != ' ' {
		t.Fatal("the bar is up during the pause")
	}
	// a key: the next page, the prompt on row 21, the bar on rows 23-24
	if !p.sfKey('x', v) {
		t.Fatal("the pause took no key")
	}
	c = fightPage(p, v)
	at(t, c, 1, 19, "WILMA is dead!!", 12, 0)
	at(t, c, 1, 21, "what does BUCK want to do?", 13, 0)
	at(t, c, 0, 23, "Fire", 0, 15) // the highlight
	at(t, c, 5, 23, "T", 15, 0)
	at(t, c, 6, 23, "arget", 10, 0)
	at(t, c, 0, 24, "C", 15, 0)
	if p.sfKey(keyOf(input.Down), v); p.sf.sel != 5 {
		t.Fatalf("Down: highlight %d, want 5 (Command)", p.sf.sel)
	}
	c = fightPage(p, v)
	at(t, c, 0, 24, "Command", 0, 15)
	p.sfKey(keyOf(input.Esc), v)
	p.sfKey('a', v)
	if n := len(p.inv.acts); n != 1 || p.inv.acts[0] != (front.Action{Kind: front.ActChoose, N: 6}) {
		t.Fatalf("acts %v, want Esc nothing and A ActChoose 6", p.inv.acts)
	}
	// a page that grows with the same lines keeps its page; the next overflow pauses again
	f := testFight()
	f.Log = append(f.Log, game.Message{Text: "a"}, game.Message{Text: "b"})
	v.Screen.SpaceFight, v.Ask = f, nil
	if sh := p.sfState(v, f); sh.Page != 1 || !sh.More {
		t.Fatalf("grown log: page %d more %v", sh.Page, sh.More)
	}
	// Sensors out: the foe's panel is the notice alone
	f = testFight()
	f.SensorsOut = true
	v = front.View{Screen: &game.Screen{Kind: game.ScreenSpaceFight, SpaceFight: f}}
	c = fightPage(&Presenter{}, v)
	at(t, c, 4, 14, "Sensors Destroyed", 12, 8)
	if c.cells[12][10].set {
		t.Fatal("foe figures shown with the sensors out")
	}
	savePNG(t, c, v)
}

func TestSfWrapPad(t *testing.T) {
	w := sfWrap(strings.Repeat("x", 40))
	if len(w) != 2 || len(w[0]) != 36 || len(w[1]) != 3 {
		t.Fatalf("no space: %q", w)
	}
	pg := sfPages([]game.Message{{Text: "one"}, {Text: "q"}}, true) // a line, a blank, the question on row 21
	if len(pg) != 1 || pg[0][0].Text != "one" || pg[0][1].Text != "" || pg[0][2].Text != "q" {
		t.Fatalf("pad: %+v", pg)
	}
	pg = sfPages([]game.Message{{Text: "a"}, {Text: "pilot", Clear: true}, {Text: ""}, {Text: "q"}}, true)
	if len(pg) != 2 || pg[1][0].Text != "pilot" || pg[1][2].Text != "q" {
		t.Fatalf("clear: %+v", pg)
	}
}

// savePNG renders the page (with the game's art when BR2's data is there)
// to SPACEFIGHT_PNG.
func savePNG(t *testing.T, _ *cellPage, v front.View) {
	out := os.Getenv("SPACEFIGHT_PNG")
	if out == "" {
		return
	}
	home, _ := os.UserHomeDir()
	if l, err := dax.NewLoader(filepath.Join(home, "Projects", "buckrogers2-go", "data", "dos")); err == nil {
		defer func(old *dax.Loader) { daxLoader = old; clear(daxPics) }(daxLoader)
		daxLoader = l
		clear(daxPics)
	}
	bits := fonts.Get("IBM_BIOS.f8")
	font := func(ch byte, up bool) [8]byte {
		if up && ch >= 'a' && ch <= 'z' {
			ch -= 32
		}
		g, _ := fonts.Glyph(bits, ch, true)
		return g
	}
	for i, f := range []*game.SpaceFight{testFight(), v.Screen.SpaceFight} {
		f.Blink = game.Blink{}
		v := front.View{Screen: &game.Screen{Kind: game.ScreenSpaceFight, SpaceFight: f}, Ask: crewMenu()}
		p := &Presenter{}
		p.sf.page = 9
		c := fightPage(p, v)
		m := image.NewRGBA(image.Rect(0, 0, 320, 200))
		c.draw(m, v, Note{Font: font})
		name := strings.TrimSuffix(out, ".png") + string(rune('0'+i)) + ".png"
		fh, _ := os.Create(name)
		png.Encode(fh, m)
		fh.Close()
	}
}
