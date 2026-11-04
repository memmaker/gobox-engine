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

// samePicker says classic's picker holds State's pictures on every page,
// from the page it opened on, on each grid.
func samePicker(t *testing.T, name string, e *rules.Engine, kind int) {
	t.Helper()
	for _, g := range grids {
		settled(e, front.Action{Kind: front.ActResize, X: g[0], Y: g[1]})
		e.ClosePicker()
		if !e.OpenPicker(kind) {
			t.Fatalf("%s: no picker", name)
		}
		pics := len(e.Rules().State().Picker.Pics)
		if kind < front.PickAdd && pics == 0 {
			t.Fatalf("%s: no pictures", name)
		}
		first := classic.Picker(e.Rules().State(), -1, g[0], g[1])
		if first == nil || first.N != pics || first.Pages < 1 {
			t.Fatalf("%s %v: %+v for %d pictures", name, g, first, pics)
		}
		for page := -1; page < first.Pages; page++ {
			got := classic.Picker(e.Rules().State(), page, g[0], g[1])
			if got == nil || got.N != pics || got.Pages != first.Pages || len(got.Cells) == 0 || len(got.Cells) > got.PerPage {
				t.Fatalf("%s %v page %d: %+v", name, g, page, got)
			}
		}
		t.Logf("%s %v: %d pictures on %d pages", name, g, first.N, first.Pages)
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
	n := len(e.Rules().State().Picker.Pics)
	e.Pick(0)
	if p := e.Rules().State().Picker; p != nil {
		if got := classic.Picker(e.Rules().State(), -1, 40, 25); got == nil || got.N != n-1 || len(p.Pics) != n-1 {
			t.Errorf("forgotten: got %+v, %d pictures of %d", got, len(p.Pics), n)
		}
	}
}

// sameIcon says classic's icon designer shows State's on each grid.
func sameIcon(t *testing.T, name string, e *rules.Engine) {
	t.Helper()
	ink := e.Rules().State().Ink
	for _, g := range grids {
		settled(e, front.Action{Kind: front.ActResize, X: g[0], Y: g[1]})
		ink = e.Rules().State().Ink
		got, want := classic.IconEditor(e.Rules().State(), &ink, g[0], g[1]), e.Rules().State().IconEditor
		if got == nil || want == nil || got.Custom != want.Custom || got.Large != want.Large {
			t.Fatalf("%s %v:\n got %+v\nwant %+v", name, g, got, want)
		}
		for p := range got.Poses { // classic builds the poses: they must be there
			if got.Poses[p] == nil && name == "icon" && !want.Custom { // (after a pick the numbers may leave the pool)
				t.Fatalf("%s %v: no pose %d built", name, g, p)
			}
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
	if e.Rules().State().IconEditor == nil {
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
	if !e.Rules().State().IconEditor.Custom {
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
