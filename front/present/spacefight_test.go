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
	"gobox/game/play"
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

func crewMenu() *play.Ask {
	a := &play.Ask{Kind: game.Choice, Break: 5}
	for _, s := range []string{"Fire", "Target", "Load", "Quit", "View", "Command", "Aid", "Sensors"} {
		a.Options = append(a.Options, play.Option{Label: s, Key: rune(s[0])})
	}
	return a
}

// sfView is a fight screen up and the question on its bar.
type sfView struct {
	Screen *game.Screen
	Ask    *play.Ask
}

// show puts v up on p: its screen a moment's, its question the game's.
func (v sfView) show(p *Presenter) *Presenter {
	if p.eng == nil {
		g, _ := playGame(play.Exploring{})
		p.SetEngine(g)
	}
	p.played = &frame{Moment: Moment{Report: play.Report{Screen: v.Screen}}}
	p.eng.r.refusalAsk = v.Ask
	return p
}

// fightPage is p's page with v up.
func fightPage(p *Presenter, v sfView) *cellPage {
	v.show(p)
	c := newCellPage(image.Point{}, p.ink(), nil)
	c.spaceFight(p.sfOf(), p.sfState(p.sfOf()), p.askUp(), p.gameScreen())
	return c
}

func cellAt(t *testing.T, c *cellPage, x, y int, s string, fg, bg int) {
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
	v := sfView{Screen: &game.Screen{Kind: game.ScreenSpaceFight, SpaceFight: testFight()}, Ask: crewMenu()}
	c := fightPage(p, v)
	cellAt(t, c, 1, 1, "Closing    ", 12, 0)
	cellAt(t, c, 1, 2, "Range: 3", 12, 0)
	cellAt(t, c, 29, 1, "Hull:  100", 11, 8)
	cellAt(t, c, 29, 3, "Sens:    0", 12, 8) // zero: red
	cellAt(t, c, 10, 12, "    0", 12, 0)
	cellAt(t, c, 10, 13, "   20", 10, 0)
	cellAt(t, c, 29, 14, "Weapons:", 10, 8)
	cellAt(t, c, 29, 15, "LASER", 10, 8)
	cellAt(t, c, 37, 15, " 2", 10, 8)
	cellAt(t, c, 29, 16, "MISSILE", 14, 8) // a mount down
	cellAt(t, c, 37, 16, " 1", 14, 8)
	// the first log line wraps at the last space before column 37: 3 rows + 2 + the prompt padded = 2 pages
	sh := p.sfState(p.sfOf())
	if len(sh.Pages) != 2 || !sh.More {
		t.Fatalf("pages %d more %v, want 2 and a pause", len(sh.Pages), sh.More)
	}
	cellAt(t, c, 1, 19, "BUCK fires lasers and hits for 12", 11, 0)
	cellAt(t, c, 1, 20, "points of damage on the pirate ship!", 11, 0)
	cellAt(t, c, 1, 21, "The pirate fires missiles and misses.", 14, 0)
	cellAt(t, c, 0, 24, sfPause, 15, 0)
	if c.cells[23][0].set && c.cells[23][0].ch != ' ' {
		t.Fatal("the bar is up during the pause")
	}
	// a key: the next page, the prompt on row 21, the bar on rows 23-24
	if !p.sfKey('x') {
		t.Fatal("the pause took no key")
	}
	c = fightPage(p, v)
	cellAt(t, c, 1, 19, "WILMA is dead!!", 12, 0)
	cellAt(t, c, 1, 21, "what does BUCK want to do?", 13, 0)
	cellAt(t, c, 0, 23, "Fire", 0, 15) // the highlight
	cellAt(t, c, 5, 23, "T", 15, 0)
	cellAt(t, c, 6, 23, "arget", 10, 0)
	cellAt(t, c, 0, 24, "C", 15, 0)
	if p.sfKey(keyOf(input.Down)); p.sf.sel != 5 {
		t.Fatalf("Down: highlight %d, want 5 (Command)", p.sf.sel)
	}
	c = fightPage(p, v)
	cellAt(t, c, 0, 24, "Command", 0, 15)
	p.sfKey(keyOf(input.Esc))
	p.sfKey('a')
	if n := len(p.inv.acts); n != 2 || p.inv.acts[0] != (front.FrontAction{Kind: front.ActClose}) || p.inv.acts[1] != (front.FrontAction{Kind: front.ActChoose, N: 6}) {
		t.Fatalf("acts %v, want Esc ActClose and A ActChoose 6", p.inv.acts)
	}
	// a page that grows with the same lines keeps its page; the next overflow pauses again
	f := testFight()
	f.Log = append(f.Log, game.Message{Text: "a"}, game.Message{Text: "b"})
	v.Screen.SpaceFight, v.Ask = f, nil
	if sh := p.sfState(f); sh.Page != 1 || !sh.More {
		t.Fatalf("grown log: page %d more %v", sh.Page, sh.More)
	}
	// Sensors out: the foe's panel is the notice alone
	f = testFight()
	f.SensorsOut = true
	v = sfView{Screen: &game.Screen{Kind: game.ScreenSpaceFight, SpaceFight: f}}
	c = fightPage(&Presenter{}, v)
	cellAt(t, c, 4, 14, "Sensors Destroyed", 12, 8)
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
func savePNG(t *testing.T, _ *cellPage, v sfView) {
	out := os.Getenv("SPACEFIGHT_PNG")
	if out == "" {
		return
	}
	home, _ := os.UserHomeDir()
	var fl *dax.Loader
	if l, err := dax.NewLoader(filepath.Join(home, "Projects", "buckrogers2-go", "data", "dos")); err == nil {
		useArt(t, l.Dir)
		fl = l
		clear(daxPics)
	}
	bits := brFont(fl)
	font := func(ch byte, text bool) [8]byte { g, _ := fonts.Glyph(bits, ch, text); return g } // as fontmenu.go hands the page the game's font
	for i, f := range []*game.SpaceFight{testFight(), v.Screen.SpaceFight} {
		f.Blink = game.Blink{}
		v := sfView{Screen: &game.Screen{Kind: game.ScreenSpaceFight, SpaceFight: f}, Ask: crewMenu()}
		p := &Presenter{}
		p.sf.page = 9
		c := fightPage(p, v)
		m := image.NewRGBA(image.Rect(0, 0, 320, 200))
		c.draw(m, (&Presenter{}).windowBorder(), image.Point{}, Note{Font: font})
		name := strings.TrimSuffix(out, ".png") + string(rune('0'+i)) + ".png"
		fh, _ := os.Create(name)
		png.Encode(fh, m)
		fh.Close()
	}
}
