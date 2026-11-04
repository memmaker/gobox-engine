package parity

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"gobox/front"
	"gobox/front/classic"
	"gobox/front/present"
)

// known are the View fields classic cannot yet build from State alone, per
// scene: X and Y on the overland (por keeps the
// dungeon's square there). Any other field that differs fails.
var known = map[string][]string{"overland": {"X", "Y"}}

// TestAdapter runs por twice over each dev scene and the same actions: as a
// front.Game itself (rules.Engine) and through classic.Game over its
// game.Rules. The Views must be DeepEqual.
func TestAdapter(t *testing.T) {
	actions := []front.Action{{}, {Kind: front.ActResize, X: 50, Y: 30}}
	for _, scene := range []string{"", "shop", "hall", "temple", "appraise", "combat", "overland", "bestiary", "parlay", "password", "who", "effects", "journal", "inventory", "lockeddoor", "loot", "cast"} {
		t.Run("scene "+scene, func(t *testing.T) {
			a, b := engine(t, true, scene), engine(t, true, scene)
			g := classic.NewGame(b.Rules(), b.Data())
			var va, vb front.View
			for i, act := range actions {
				va = a.Apply(act)
				g.Answer(act)
				vb = g.Settle()
				if i == 0 {
					continue
				}
			}
			defer a.Apply(front.Action{Kind: front.ActResize, X: 40, Y: 25}) // the page size is global in por: put it back
			for _, v := range []*front.View{&va, &vb} {
				v.Seq, v.Note.At, v.Events = 0, time.Time{}, nil
				noPartyIcons(v.Battle)
			}
			for _, n := range diffFields(va, vb) {
				if slices.Contains(known[scene], n) {
					t.Logf("known gap: %s", n)
				} else {
					t.Errorf("%s differs:\n  engine  %+v\n  adapter %+v", n, reflect.ValueOf(va).FieldByName(n), reflect.ValueOf(vb).FieldByName(n))
				}
			}
		})
	}
}

// TestGoldensAdapter: por's creation goldens through classic.Game alone (no
// engine front.Game under it), the actions answered by the adapter.
func TestGoldensAdapter(t *testing.T) {
	e := partyMenu(t)
	g := classic.NewGame(e.Rules(), e.Data())
	p := &present.Presenter{}
	p.SetEngine(g)
	shown := func(as ...front.Action) front.View {
		for _, a := range as {
			p.Decide(g.View(), a)
			for _, f := range p.Calls() {
				f(g)
			}
			for _, a := range p.Actions() {
				g.Answer(a)
			}
			g.Settle()
		}
		return p.Own(g.View())
	}
	v := shown(front.Action{Kind: front.ActCreate})
	if v.Creation == nil {
		t.Fatal("no creation")
	}
	golden(t, "creation", v)
	golden(t, "pickhead", shown(front.Action{Kind: front.ActPicker, N: front.PickHead}))
}

// TestSheetAdapter: por's sheet goldens through classic.Game alone.
func TestSheetAdapter(t *testing.T) {
	e := engine(t, true, "")
	g := classic.NewGame(e.Rules(), e.Data())
	g.OpenSheet(-1)
	v := g.Settle()
	if v.Sheet == nil {
		t.Fatal("no sheet")
	}
	golden(t, "sheet", v)
	g.ShowMember(1)
	golden(t, "sheet2", g.Settle())
}

// noPartyIcons blanks the party units' icons: por composes none (gobox does,
// TestBattleView).
func noPartyIcons(b *front.Battle) {
	if b == nil {
		return
	}
	for i := range b.Units {
		if b.Units[i].Party {
			b.Units[i].Icon, b.Units[i].IconSize, b.IconPix[i] = nil, image.Point{}, front.Picture{}
		}
	}
}

// writePNG is a w x h noise PNG in a temp dir.
func writePNG(t *testing.T, w, h int) string {
	m := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := range m.Pix {
		m.Pix[i] = byte(i * 7)
	}
	path := filepath.Join(t.TempDir(), "p.png")
	f, _ := os.Create(path)
	png.Encode(f, m)
	f.Close()
	return path
}

// shownBy gives each of as to presenter p over g (its screens take theirs, the
// rest go to the game), runs its calls and returns the page it shows.
func shownBy(g *classic.Game, p *present.Presenter, as ...front.Action) front.View {
	for _, a := range as {
		p.Decide(g.View(), a)
		for _, f := range p.Calls() {
			f(g)
		}
		for _, a := range p.Actions() {
			g.Answer(a)
		}
		g.Settle()
	}
	return p.Own(g.View())
}

// TestPagesAdapter: por's presentdata goldens (creationpng, icon, pickweapon,
// the party roster, picksaved, slots) through classic.Game alone.
func TestPagesAdapter(t *testing.T) {
	e := partyMenu(t)
	g := classic.NewGame(e.Rules(), e.Data())
	p := &present.Presenter{}
	p.SetEngine(g)
	v := shownBy(g, p, front.Action{Kind: front.ActCreate})
	if v.Creation == nil {
		t.Fatal("no creation")
	}
	golden(t, "creation", v)
	golden(t, "creationpng", shownBy(g, p, front.Action{Kind: front.ActIcon, Text: writePNG(t, 48, 24)}))
	golden(t, "pickhead", shownBy(g, p, front.Action{Kind: front.ActPicker, N: front.PickHead}))
	v = shownBy(g, p, front.Action{Kind: front.ActClose}, front.Action{Kind: front.ActIconEdit, N: -1})
	if !v.IconDrop || !v.IconEditor.Custom {
		t.Fatal("no custom icon editor")
	}
	if v = shownBy(g, p, front.Action{Kind: front.ActIconRemove}); v.IconEditor == nil || v.IconEditor.Custom {
		t.Fatal("Remove left the dropped icon")
	}
	golden(t, "icon", v)
	shownBy(g, p, front.Action{Kind: front.ActIconSize})
	shownBy(g, p, front.Action{Kind: front.ActIconSize})
	golden(t, "pickweapon", shownBy(g, p, front.Action{Kind: front.ActPicker, N: front.PickWeapon}))
}

// TestPartyAdapter: the party roster, the saved characters and the slots as por's goldens.
func TestPartyAdapter(t *testing.T) {
	e := partyMenu(t)
	e.QuickStart()
	g := classic.NewGame(e.Rules(), e.Data())
	p := &present.Presenter{}
	p.SetEngine(g)
	v := g.Settle()
	want, err := os.ReadFile(filepath.Join(porDir(t), "presentdata", "testdata", "party.txt"))
	if err != nil {
		t.Fatal(err)
	}
	var rows, got []string
	for _, l := range strings.Split(string(want), "\n") {
		var y, x int
		var fg, q string
		if n, _ := fmt.Sscanf(l, "text %d,%d %s true %q", &y, &x, &fg, &q); n == 4 && y >= 4 {
			for len(rows) <= y-4 {
				rows = append(rows, "")
			}
			rows[y-4] += " " + q
		}
	}
	for _, m := range v.Party {
		got = append(got, fmt.Sprintf(" %s %d %d", m.Name, m.AC, m.HP))
	}
	if !slices.Equal(got, rows) {
		t.Errorf("party %q, the roster showed %q", got, rows)
	}
	v = shownBy(g, p, front.Action{Kind: front.ActRemove, N: 0}, front.Action{Kind: front.ActAdd})
	if v.Picker == nil {
		t.Fatal("no saved characters")
	}
	golden(t, "picksaved", v)
	v = shownBy(g, p, front.Action{Kind: front.ActClose})
	if v.Slots = g.SlotsView(false); v.Slots == nil {
		t.Fatal("no slots")
	}
	golden(t, "slots", v)
}
