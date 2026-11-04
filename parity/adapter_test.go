package parity

import (
	"encoding/json"
	"fmt"
	"gobox/game"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gobox/front"
	"gobox/front/classic"
	"gobox/front/present"
	"por/presentdata"
)

// TestAdapter runs por through classic.Game over each dev scene and the
// same actions: the State after each is por's (wantState).
func TestAdapter(t *testing.T) {
	actions := []front.Action{{}, {Kind: front.ActResize, X: 50, Y: 30}}
	for _, scene := range []string{"", "shop", "hall", "temple", "appraise", "combat", "overland", "bestiary", "parlay", "password", "who", "effects", "journal", "inventory", "lockeddoor", "loot", "cast"} {
		t.Run("scene "+scene, func(t *testing.T) {
			b := engine(t, true, scene)
			g := classic.NewGame(b.Rules(), b.Data())
			defer g.Answer(front.Action{Kind: front.ActResize, X: 40, Y: 25}) // the page size is global in por: put it back
			for _, act := range actions {
				g.Answer(act)
				g.Settle()
				at(t, "after", b)
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

// TestPagesAdapter: por's presentdata goldens (moved here from por's
// TestCreationPage, unchanged, with its icon editor clicks) (creationpng, icon, pickweapon,
// the party roster, picksaved, slots) through classic.Game alone.
func TestPagesAdapter(t *testing.T) {
	present.Host = &presentdata.Host.Options
	present.SetData(presentdata.Data("")) // the presenter's game data: the icon editor's part names
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
	ed := v.IconEditor
	for _, k := range []struct {
		x, y int
		want front.Action
	}{
		{ed.Head.X1, ed.Head.Y1, front.Action{Kind: front.ActPicker, N: front.PickIconHead}},
		{ed.PoseAt[1].X1, ed.PoseAt[1].Y1 + 5, front.Action{Kind: front.ActPicker, N: front.PickWeapon}},
		{ed.Size.X2, ed.Size.Y1, front.Action{Kind: front.ActIconSize}},
		{ed.Parts.X1 + 3, ed.Parts.Y1 + 2, front.Action{Kind: front.ActIconColour, N: 2, X: 1, Y: -1}},
		{ed.Palette.X1 + 3, ed.Palette.Y1 + 1, front.Action{Kind: front.ActIconColour, N: ed.Part, X: ed.Shade, Y: ed.Colours[9]}},
	} {
		if a, ok := present.IconClick(ed, k.x, k.y); !ok || a != k.want {
			t.Errorf("click %d,%d: %+v, want %+v", k.x, k.y, a, k.want)
		}
	}
	if v = shownBy(g, p, front.Action{Kind: front.ActIconSize}); v.IconEditor == nil || v.IconEditor.Large == ed.Large {
		t.Error("Size not taken")
	}
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

// TestMapAdapter: the typed data behind por's recorded map pngs (map_*.json)
// rebuilt through classic.Map from the State.Map it implies equals itself, so
// present.MapImage gives the same pixels as por's TestMapImage.
func TestMapAdapter(t *testing.T) {
	files, _ := filepath.Glob(filepath.Join(porDir(t), "presentdata", "testdata", "map_*.json"))
	if len(files) == 0 {
		t.Fatal("no recorded maps")
	}
	kinds := [5]string{"", "shop", "temple", "inn", "civic"}
	inks := [5]int{15, 10, 13, 14, 11} // por's label inks by kind
	for _, f := range files {
		var rec struct {
			Squares [16][16]struct {
				Seen   front.Seen
				Walls  [4]front.MapWall
				Ink    int
				Letter byte
			}
			Party, Ahead image.Point
		}
		b, _ := os.ReadFile(f)
		if err := json.Unmarshal(b, &rec); err != nil {
			t.Fatal(err)
		}
		a := game.Automap{Size: image.Pt(16, 16)}
		var have [26]bool
		maxL := byte('A')
		for y, row := range rec.Squares {
			for x, q := range row {
				at := image.Pt(x, y)
				if q.Seen >= front.SeenFar {
					a.Known = append(a.Known, at)
				}
				if q.Seen == front.Visited {
					a.Walked = append(a.Walked, at)
				}
				for i, w := range q.Walls {
					switch w {
					case front.Wall:
						a.Walls = append(a.Walls, game.Wall{At: at, Side: i * 2})
					case front.Door:
						a.Doors = append(a.Doors, game.Wall{At: at, Side: i * 2, State: game.DoorClosed})
					case front.LockedDoor:
						a.Doors = append(a.Doors, game.Wall{At: at, Side: i * 2, State: game.DoorLocked})
					}
				}
				if q.Letter != 0 {
					a.Labels = append(a.Labels, game.Label{At: at, Text: string(q.Letter), Kind: kinds[slices.Index(inks[:], q.Ink)]})
					have[q.Letter-'A'], maxL = true, max(maxL, q.Letter)
				}
			}
		}
		for l := byte('A'); l < maxL; l++ { // the letters before the last are other labels' (off the map)
			if !have[l-'A'] {
				a.Labels = append(a.Labels, game.Label{At: image.Pt(-1, -1), Text: string(l)})
			}
		}
		facing := slices.Index([]image.Point{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}, rec.Ahead) * 2
		s := game.State{Map: a, Where: game.Where{X: rec.Party.X, Y: rec.Party.Y, Facing: facing}}
		got := classic.Map(&front.MapSquares{}, s)
		for y := range got.Squares {
			for x, q := range got.Squares[y] {
				w := rec.Squares[y][x]
				if q.Seen != w.Seen || q.Walls != w.Walls || q.Letter != w.Letter || (w.Letter != 0 && inks[q.Kind] != w.Ink) {
					t.Fatalf("%s: square %d,%d is %+v, want %+v", filepath.Base(f), x, y, q, w)
				}
			}
		}
		if got.Party != rec.Party || got.Ahead != rec.Ahead {
			t.Errorf("%s: party %v ahead %v, want %v %v", filepath.Base(f), got.Party, got.Ahead, rec.Party, rec.Ahead)
		}
	}
}
