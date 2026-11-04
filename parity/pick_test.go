package parity

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gobox/front"
	"gobox/front/classic"
	"gobox/front/present"
	"gobox/sprites"
	"por/rules"
)

var grids = [][2]int{{40, 25}, {83, 80}, {120, 30}}

// samePicker says classic's picker equals por's on every page, from the
// page it opened on, on each grid.
func samePicker(t *testing.T, name string, e *rules.Engine, kind int) {
	t.Helper()
	for _, g := range grids {
		settled(e, front.Action{Kind: front.ActResize, X: g[0], Y: g[1]})
		e.ClosePicker()
		if !e.OpenPicker(kind) {
			t.Fatalf("%s: no picker", name)
		}
		want := e.PickerView(-1)
		if kind < front.PickAdd && len(want.Pics) == 0 {
			t.Fatalf("%s: no pictures", name)
		}
		for page := -1; page < want.Pages; page++ {
			got, want := classic.Picker(e.Rules().State(), page, g[0], g[1]), e.PickerView(page)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("%s %v page %d:\n got %+v\nwant %+v", name, g, page, got, want)
			}
		}
		t.Logf("%s %v: %d pictures on %d pages", name, g, want.N, want.Pages)
	}
	e.ClosePicker()
	if classic.Picker(e.Rules().State(), -1, 40, 25) != nil {
		t.Errorf("%s: a picker after it closed", name)
	}
}

func TestPickerParity(t *testing.T) {
	e := partyMenu(t)
	if err := e.CreateCharacter(); err != nil {
		t.Fatal(err)
	}
	settled(e, front.Action{})
	samePicker(t, "heads", e, front.PickHead)
	samePicker(t, "bodies", e, front.PickBody)
	e.EndCreation(false)
	settled(e, front.Action{})

	e.QuickStart()
	settled(e, front.Action{})
	e.EditIcon(-1)
	settled(e, front.Action{})
	poolIconPick(t, e, front.PickIconHead, "combat_head_")
	poolIconPick(t, e, front.PickWeapon, "combat_body_")
	e.EndIcon(false)

	e.DropMember(0) // a character saved: the party menu's lists
	settled(e, front.Action{})
	for _, k := range []int{front.PickAdd, front.PickForget} {
		if !e.OpenPicker(k) {
			t.Fatalf("no saved characters for picker %d", k)
		}
		e.ClosePicker()
		samePicker(t, "saved", e, k)
	}
	e.OpenPicker(front.PickForget) // forgetting one: the rest
	n := e.PickerView(-1).N
	e.Pick(0)
	if e.PickerView(-1) != nil {
		if got, want := classic.Picker(e.Rules().State(), -1, 40, 25), e.PickerView(-1); !reflect.DeepEqual(got, want) || want.N != n-1 {
			t.Errorf("forgotten: got %+v\nwant %+v", got, want)
		}
	}
}

// sameIcon says classic's icon designer equals por's on each grid.
func sameIcon(t *testing.T, name string, e *rules.Engine) {
	t.Helper()
	ink := e.Rules().State().Ink
	for _, g := range grids {
		settled(e, front.Action{Kind: front.ActResize, X: g[0], Y: g[1]})
		ink = e.Rules().State().Ink
		got, want := classic.IconEditor(e.Rules().State(), &ink, g[0], g[1]), e.IconEditorView()
		for p := range want.Poses { // por leaves the poses to gobox: classic's own are built, and must be there
			if want.Poses[p] == nil {
				if got.Poses[p] == nil && name == "icon" { // (after a pick the numbers may leave the pool)
					t.Fatalf("%s %v: no pose %d built", name, g, p)
				}
				want.Poses[p] = got.Poses[p]
			}
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s %v:\n got %+v\nwant %+v", name, g, got, want)
		}
	}
}

func TestIconEditorParity(t *testing.T) {
	e := partyMenu(t)
	classic.GameID = e.Data().ID // the origins the poses are built from
	if classic.IconEditor(e.Rules().State(), new([16]color.RGBA), 40, 25) != nil {
		t.Fatal("an icon editor before it is up")
	}
	e.QuickStart()
	settled(e, front.Action{})
	e.EditIcon(-1)
	settled(e, front.Action{})
	if e.Rules().State().IconEditor == nil || e.IconEditorView() == nil {
		t.Fatal("no icon editor")
	}
	sameIcon(t, "icon", e)
	e.IconSize()
	sameIcon(t, "icon size", e)
	e.IconColour(2, 1, 9)
	e.IconColour(5, 0, 3)
	sameIcon(t, "icon colours", e)
	e.Pick(0) // no picker: nothing
	e.OpenPicker(front.PickWeapon)
	e.Pick(7)
	sameIcon(t, "icon weapon", e)

	p := filepath.Join(t.TempDir(), "p.png") // a dropped PNG: custom
	m := image.NewRGBA(image.Rect(0, 0, 48, 24))
	for i := range m.Pix {
		m.Pix[i] = byte(i * 7)
	}
	f, _ := os.Create(p)
	png.Encode(f, m)
	f.Close()
	e.IconPNG(p)
	if !e.IconEditorView().Custom {
		t.Fatal("not custom")
	}
	sameIcon(t, "icon custom", e)
	e.IconRemove()
	sameIcon(t, "icon removed", e)
	e.EndIcon(false)
	if e.Rules().State().IconEditor != nil {
		t.Error("an icon editor after EndIcon")
	}
}

// TestTitleParity: the party menu's title picture and backdrop from Data.Art,
// read live as classic.Game does, are por's (wantTitle), and the hall has none.
func TestTitleParity(t *testing.T) {
	e := partyMenu(t)
	at(t, "menu", e)
	logo, backdrop := classic.Title(nil, nil, e.Data().Art)
	if logo == nil {
		t.Fatal("no title picture")
	}
	bd := *backdrop // the file's name, not where the checkout is
	bd.File = filepath.Base(bd.File)
	if got := sum([]any{logo, bd}); got != wantTitle {
		t.Errorf("title %s, por %s", got, wantTitle)
	}
	if c := gameOf(e).View(); !reflect.DeepEqual(c.Logo, logo) || !reflect.DeepEqual(c.Backdrop, backdrop) || c.MainMenu == nil {
		t.Errorf("title through classic: logo %v backdrop %v", c.Logo != nil, c.Backdrop)
	}
	h := engine(t, true, "hall") // the hall has none
	settled(h, front.Action{})
	at(t, "hall", h)
	if c := gameOf(h).View(); c.Logo != nil || c.Backdrop != nil {
		t.Error("a title in the hall")
	}
}

// poolIconPick: the icon head and weapon pickers are gobox's pool pickers (por publishes none): each lists
// every combat name with prefix, and a pick is stored as the member's pool id.
func poolIconPick(t *testing.T, e *rules.Engine, kind int, prefix string) {
	t.Helper()
	p := &present.Presenter{}
	g := present.WithRules(gameOf(e), p)
	p.SetEngine(g)
	if !g.OpenPicker(kind) {
		t.Fatalf("icon pick %d: no pool picker", kind)
	}
	n := 0
	for _, name := range sprites.Names("pool", "combat") {
		if strings.HasPrefix(name, prefix) {
			n++
		}
	}
	if v := g.PickerView(-1); v == nil || n == 0 || v.N != n || len(v.Portraits) != n {
		t.Fatalf("icon pick %d: %d pool names, picker %+v", kind, n, v)
	}
	if !g.Pick(0) {
		t.Fatalf("icon pick %d: pick refused", kind)
	}
	if s := e.Rules().State(); !strings.HasPrefix(map[int]string{front.PickIconHead: s.Party[s.Selected].IconHeadID, front.PickWeapon: s.Party[s.Selected].IconBodyID}[kind], prefix) {
		t.Errorf("icon pick %d: pick not stored", kind)
	}
}
