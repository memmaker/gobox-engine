package present

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"gobox/dax"
	"gobox/front"
	"gobox/game"
	"gobox/ui/fonts"
)

func testSheet(own bool) *game.ShipSheet {
	return &game.ShipSheet{Own: own, Name: "MAELSTROM RIG", Type: "CRUISER", Race: "RAM", Tonnage: 400, Length: 120, Width: 40, Cargo: 50,
		AC: 5, Crew: 12, Speed: 3, React: 2, Range: game.RangeRamming, MedSupplies: 7,
		Weapons: []game.WeaponSlot{{Short: "LASER", Mounts: 2, Working: 2, Ammo: 0, Reloads: 0}, {Short: "MISSILE", Mounts: 3, Working: 1, Ammo: 4, Reloads: 2}}}
}

func TestShipSheetFoeUnknown(t *testing.T) {
	v := front.View{Screen: &game.Screen{Kind: game.ScreenSpaceFight, SpaceFight: testFight(), ShipSheet: testSheet(false)}, Ask: &front.Ask{Kind: front.AskMenu}}
	c := fightPage(&Presenter{}, v)
	at(t, c, 1, 1, "name:", 14, 8)
	at(t, c, 10, 1, "MAELSTROM RIG", 15, 8)
	at(t, c, 1, 4, "tonnage:", 14, 8)
	at(t, c, 10, 4, "  400 tons", 15, 8)
	at(t, c, 10, 5, "  120 feet", 15, 8)
	at(t, c, 7, 9, "    ?", 12, 8)
	at(t, c, 13, 9, "crew:", 14, 8)
	at(t, c, 20, 9, "    ?", 12, 8)
	at(t, c, 7, 10, "    ?", 12, 8)
	at(t, c, 20, 10, "    ?", 12, 8)
	at(t, c, 1, 11, "range:", 14, 8)
	at(t, c, 10, 11, "ramming", 14, 8)
	at(t, c, 1, 13, "Weapons:", 10, 8)
	at(t, c, 13, 13, "Ammo", 10, 8)
	at(t, c, 18, 13, "Reloads", 10, 8)
	at(t, c, 1, 14, "LASER", 10, 8)
	at(t, c, 9, 14, " 2", 10, 8) // the working count only
	at(t, c, 11, 14, "    ?", 12, 8)
	at(t, c, 1, 15, "MISSILE", 14, 8)
	at(t, c, 9, 15, " 1", 14, 8)
	at(t, c, 29, 1, "Hull:  100", 11, 8) // the fight's own panel stays
	at(t, c, 0, 24, sfPause, 15, 0)
	v.Screen.ShipSheet.Known, v.Screen.ShipSheet.Range = true, game.RangeLong
	c = fightPage(&Presenter{}, v)
	at(t, c, 7, 9, "    5", 15, 8)
	at(t, c, 17, 15, "    2", 15, 8)
	at(t, c, 10, 11, "long", 15, 8)
}

// The rocket's R: the own sheet in place of the left half, the own panel and weapons beside it.
func TestShipSheetRocket(t *testing.T) {
	f := testFight()
	f.Log, f.Pause = nil, "Press button or enter to continue"
	v := front.View{Screen: &game.Screen{Kind: game.ScreenShipSheet, SpaceFight: f, ShipSheet: testSheet(true)}, Ask: &front.Ask{Kind: front.AskMenu}}
	if sfOf(v) == nil {
		t.Fatal("no fight screen for the sheet")
	}
	c := fightPage(&Presenter{}, v)
	at(t, c, 1, 1, "name:", 11, 8)
	at(t, c, 7, 9, "    5", 15, 8)
	at(t, c, 1, 11, "medical supplies:", 11, 8)
	at(t, c, 20, 11, "    7", 15, 8)
	at(t, c, 11, 15, "    4", 15, 8)
	at(t, c, 29, 1, "Hull:  100", 11, 8)
	at(t, c, 29, 15, "LASER", 10, 8)
	at(t, c, 0, 24, f.Pause, 15, 0)
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
	v := front.View{Screen: &game.Screen{Kind: game.ScreenSpaceFight, SpaceFight: f}, Ask: crewMenu()}
	c := fightPage(&Presenter{}, v)
	at(t, c, 0, 24, f.StatusLine, 10, 0)
	if k := c.cells[23][0]; k.set && k.ch != ' ' {
		t.Fatal("bar under the status line")
	}
	// the crew menu's padding comes with Break only
	f.StatusLine = ""
	v.Ask = &front.Ask{Kind: front.AskMenu, Menu: front.Menu{Items: []front.MenuItem{{Text: "Hull"}, {Text: "Sens"}}}}
	c = fightPage(&Presenter{}, v)
	at(t, c, 1, 19, "one", 10, 0)
	at(t, c, 1, 20, "two", 10, 0)
	at(t, c, 0, 24, "Hull", 0, 15)
}

// saveSheetPNG renders v with BR2's art to SHIPSHEET_PNG.
func saveSheetPNG(t *testing.T, v front.View) {
	out := os.Getenv("SHIPSHEET_PNG")
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
	for i, sc := range []*game.Screen{v.Screen, {Kind: game.ScreenSpaceFight, SpaceFight: testFight(), ShipSheet: testSheet(false)}} {
		w := v
		w.Screen = sc
		c := fightPage(&Presenter{}, w)
		m := image.NewRGBA(image.Rect(0, 0, 320, 200))
		c.draw(m, w, Note{Font: font})
		fh, _ := os.Create(filepath.Join(out, "sheet"+string(rune('0'+i))+".png"))
		png.Encode(fh, m)
		fh.Close()
	}
}
