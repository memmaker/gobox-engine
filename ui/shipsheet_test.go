package ui

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"gobox/asset"
	"gobox/asset/orig"
	"gobox/game"
	"gobox/game/play"
)

func testSheet(own bool) *game.ShipSheet {
	return &game.ShipSheet{Own: own, Name: "MAELSTROM RIG", Type: "CRUISER", Race: "RAM", Tonnage: 400, Length: 120, Width: 40, Cargo: 50,
		AC: 5, Crew: 12, Speed: 3, React: 2, Range: game.RangeRamming, MedSupplies: 7,
		Weapons: []game.WeaponSlot{{Short: "LASER", Mounts: 2, Working: 2, Ammo: 0, Reloads: 0}, {Short: "MISSILE", Mounts: 3, Working: 1, Ammo: 4, Reloads: 2}}}
}

func TestShipSheetFoeUnknown(t *testing.T) {
	v := sfView{Screen: &game.Screen{Kind: game.ScreenSpaceFight, SpaceFight: testFight(), ShipSheet: testSheet(false)}, Ask: &play.Ask{Kind: game.Choice}}
	c := fightPage(&Presenter{}, v)
	cellAt(t, c, 1, 1, "name:", 14, 8)
	cellAt(t, c, 10, 1, "MAELSTROM RIG", 15, 8)
	cellAt(t, c, 1, 4, "tonnage:", 14, 8)
	cellAt(t, c, 10, 4, "  400 tons", 15, 8)
	cellAt(t, c, 10, 5, "  120 feet", 15, 8)
	cellAt(t, c, 7, 9, "    ?", 12, 8)
	cellAt(t, c, 13, 9, "crew:", 14, 8)
	cellAt(t, c, 20, 9, "    ?", 12, 8)
	cellAt(t, c, 7, 10, "    ?", 12, 8)
	cellAt(t, c, 20, 10, "    ?", 12, 8)
	cellAt(t, c, 1, 11, "range:", 14, 8)
	cellAt(t, c, 10, 11, "ramming", 14, 8)
	cellAt(t, c, 1, 13, "Weapons:", 10, 8)
	cellAt(t, c, 13, 13, "Ammo", 10, 8)
	cellAt(t, c, 18, 13, "Reloads", 10, 8)
	cellAt(t, c, 1, 14, "LASER", 10, 8)
	cellAt(t, c, 9, 14, " 2", 10, 8) // the working count only
	cellAt(t, c, 11, 14, "    ?", 12, 8)
	cellAt(t, c, 1, 15, "MISSILE", 14, 8)
	cellAt(t, c, 9, 15, " 1", 14, 8)
	cellAt(t, c, 29, 1, "Hull:  100", 11, 8) // the fight's own panel stays
	cellAt(t, c, 0, 24, sfPause, 15, 0)
	v.Screen.ShipSheet.Known, v.Screen.ShipSheet.Range = true, game.RangeLong
	c = fightPage(&Presenter{}, v)
	cellAt(t, c, 7, 9, "    5", 15, 8)
	cellAt(t, c, 17, 15, "    2", 15, 8)
	cellAt(t, c, 10, 11, "long", 15, 8)
}

// The rocket's R: the own sheet in place of the left half, the own panel and weapons beside it.
func TestShipSheetRocket(t *testing.T) {
	f := testFight()
	f.Log, f.Pause = nil, "Press button or enter to continue"
	v := sfView{Screen: &game.Screen{Kind: game.ScreenShipSheet, SpaceFight: f, ShipSheet: testSheet(true)}, Ask: &play.Ask{Kind: game.Choice}}
	if v.show(&Presenter{}).sfOf() == nil {
		t.Fatal("no fight screen for the sheet")
	}
	c := fightPage(&Presenter{}, v)
	cellAt(t, c, 1, 1, "name:", 11, 8)
	cellAt(t, c, 7, 9, "    5", 15, 8)
	cellAt(t, c, 1, 11, "medical supplies:", 11, 8)
	cellAt(t, c, 20, 11, "    7", 15, 8)
	cellAt(t, c, 11, 15, "    4", 15, 8)
	cellAt(t, c, 29, 1, "Hull:  100", 11, 8)
	cellAt(t, c, 29, 15, "LASER", 10, 8)
	cellAt(t, c, 0, 24, f.Pause, 15, 0)
	if k := c.cells[24][len(f.Pause)]; k.set && k.ch != ' ' {
		t.Fatalf("pause text runs on: %q", k.ch)
	}
	for _, x := range []int{25, 26, 27} { // nothing of the foe's side
		if k := c.cells[18][x]; k.set && k.ch != ' ' {
			t.Fatalf("(%d,18) %q", x, k.ch)
		}
	}
	for _, k := range c.cells[12][10:15] { // the foe's figures not drawn
		if k.set && k.ch != ' ' {
			t.Fatal("foe panel on the rocket's sheet")
		}
	}
	saveSheetPNG(t, v)
}

func TestSpaceFightStatusLineAndPad(t *testing.T) {
	f := testFight()
	f.Log = []game.Message{{Text: "one"}, {Text: "two"}}
	f.StatusLine = "Your Teammate is Dying!"
	v := sfView{Screen: &game.Screen{Kind: game.ScreenSpaceFight, SpaceFight: f}, Ask: crewMenu()}
	c := fightPage(&Presenter{}, v)
	cellAt(t, c, 0, 24, f.StatusLine, 10, 0)
	if k := c.cells[23][0]; k.set && k.ch != ' ' {
		t.Fatal("bar under the status line")
	}
	// the crew menu's padding comes with Break only
	f.StatusLine = ""
	v.Ask = &play.Ask{Kind: game.Choice, Options: []play.Option{{Label: "Hull"}, {Label: "Sens"}}}
	c = fightPage(&Presenter{}, v)
	cellAt(t, c, 1, 19, "one", 10, 0)
	cellAt(t, c, 1, 20, "two", 10, 0)
	cellAt(t, c, 0, 24, "Hull", 0, 15)
}

// The after-combat message: one Loud line in the window (row 19) over the fight, the key pause on row 24.
func TestSpaceFightDestroyedPause(t *testing.T) {
	f := testFight()
	f.Log = []game.Message{{Tone: game.ToneLoud, Text: "Enemy ship destroyed."}}
	f.Pause = "Press button or enter to continue"
	v := sfView{Screen: &game.Screen{Kind: game.ScreenSpaceFight, SpaceFight: f}, Ask: &play.Ask{Kind: game.Choice}}
	c := fightPage(&Presenter{}, v)
	cellAt(t, c, 1, 19, "Enemy ship destroyed.", 15, 0)
	cellAt(t, c, 0, 24, f.Pause, 15, 0)
	cellAt(t, c, 29, 1, "Hull:  100", 11, 8)
	saveSheetPNG(t, v)
}

// saveSheetPNG renders v with BR2's art to SHIPSHEET_PNG.
func saveSheetPNG(t *testing.T, v sfView) {
	out := os.Getenv("SHIPSHEET_PNG")
	if out == "" {
		return
	}
	home, _ := os.UserHomeDir()
	var fl *orig.Loader
	if l, err := orig.NewLoader(filepath.Join(home, "Projects", "buckrogers2-go", "data", "dos")); err == nil {
		useArt(t, l.Dir)
		fl = l
		clear(daxPics)
	}
	bits := brFont(fl)
	font := func(ch byte, text bool) [8]byte { g, _ := asset.Glyph(bits, ch, text); return g } // as fontmenu.go hands the page the game's font
	for i, sc := range []*game.Screen{v.Screen, {Kind: game.ScreenSpaceFight, SpaceFight: testFight(), ShipSheet: testSheet(false)}} {
		w := v
		w.Screen = sc
		c := fightPage(&Presenter{}, w)
		m := image.NewRGBA(image.Rect(0, 0, 320, 200))
		c.draw(m, (&Presenter{}).windowBorder(), image.Point{}, Note{Font: font})
		fh, _ := os.Create(filepath.Join(out, "sheet"+string(rune('0'+i))+".png"))
		png.Encode(fh, m)
		fh.Close()
	}
}

// brFont is BR2's own text font as its port hands it to gobox (Data.Fonts,
// buckrogers2-go rules/data.go ownFont): 8X8D1 block 201, a character c at
// c mod 64 for 0x20..0x7F (a-z the capitals), the symbols from 0x80.
func brFont(l *orig.Loader) []byte {
	if l == nil {
		return nil
	}
	b := l.BlockShared("8X8D1", 201)
	if len(b) < 64*8 {
		return nil
	}
	out := make([]byte, 256*8)
	for ch := 0x20; ch < 0x80; ch++ {
		c := ch
		if c >= 'a' && c <= 'z' {
			c -= 0x20
		}
		copy(out[ch*8:ch*8+8], b[c%64*8:])
	}
	for k := 64; k*8+8 <= len(b) && 0x80+k-64 < 256; k++ {
		copy(out[(0x80+k-64)*8:], b[k*8:k*8+8])
	}
	return out
}
