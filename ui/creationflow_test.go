package ui

import (
	"image"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"gobox/engine"
	"gobox/game"
	"gobox/importer"
	"gobox/ui/input"
)

const flowW, flowH = 1280, 800

// flowFront is cotab's Front on the imported data (skipped when not
// imported), its Party submenu up; roster is its charlist.txt.
func flowFront(t *testing.T, roster string) (*Front, *engine.Engine) {
	t.Helper()
	const adv = "../data/cotab"
	fo, err := importer.Read(adv)
	if err != nil {
		t.Skipf("cotab: not imported: %v", err)
	}
	eng, err := engine.NewPool(adv, adv+"/sources")
	if err != nil {
		t.Fatal(err)
	}
	if err := eng.SetStart(fo.Manifest.Start); err != nil {
		t.Fatal(err)
	}
	eng.Cfg.SaveDir = t.TempDir()
	if roster != "" {
		if err := os.WriteFile(filepath.Join(eng.Cfg.SaveDir, "CHARLIST.TXT"), []byte(roster), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	seed := uint32(1)
	eng.FixedSeed = &seed
	data := eng.Data()
	f := NewFront(NewGame(eng, data), data)
	f.Start("")
	for i := range 400 { // Enter through the title to the party menu
		f.Frame(flowW, flowH)
		if _, ok := eng.State().(game.Assembling); ok && i > 50 {
			break
		}
		if i%40 == 0 {
			f.Key('\r')
		}
	}
	for i := 0; i < 5 && !f.P.mm.sub; i++ {
		flowKey(f, 'p')
	}
	if !f.P.mm.sub {
		t.Fatal("no Party submenu")
	}
	return f, eng
}

// flowKey is key k and the frames that show what it did.
func flowKey(f *Front, k uint16) {
	f.Key(k)
	for range 3 {
		f.Frame(flowW, flowH)
	}
}

// flowClick is the shell's left click on the middle of page cell r.
func flowClick(f *Front, r CellRect) {
	q := f.P.cellRect(r)
	x, y := (q.Min.X+q.Max.X)/2, (q.Min.Y+q.Max.Y)/2
	f.Hover(x, y, false)
	if !f.Click(x, y) {
		cx, cy, in := f.Cell(float64(x), float64(y))
		f.Mouse(cx, cy, in, true, false, 0)
	}
	for range 3 {
		f.Frame(flowW, flowH)
	}
}

// drawnPortraits checks the page drawn shows the laid out page's portraits,
// each in its cell: the picture under the pointer is the one a click takes.
func drawnPortraits(t *testing.T, p *Presenter, what string) {
	t.Helper()
	pk := p.picker(p.ink())
	if pk == nil {
		t.Fatalf("%s: no picker", what)
	}
	for k, r := range pk.Cells {
		want := pk.Portraits[pk.Page*pk.PerPage+k]
		if want == nil {
			t.Fatalf("%s: cell %d has no portrait", what, k)
		}
		if !slices.ContainsFunc(p.pg.overs, func(o overlay) bool { return o.Img == want && o.At.In(r.Pixels()) }) {
			t.Fatalf("%s: page %d cell %d: the page drawn does not show its portrait", what, pk.Page, k)
		}
	}
}

// The creation sheet's head picker: no refusal toast for opening it, the
// page drawn is the page the pointer and the keys act on after every page
// turn (class: a page drawn from the presenter's own state that pageKey
// does not name stays stale), and a pick reaches the draft.
func TestCreationHeadPicker(t *testing.T) {
	f, eng := flowFront(t, "")
	p := &f.P
	flowKey(f, 'c')
	flowKey(f, 'r') // Roll
	cr := p.creation()
	if cr == nil {
		t.Fatal("no creation sheet")
	}
	flowClick(f, cr.Head)
	if p.gridPick == nil {
		t.Fatal("no head picker")
	}
	if n := p.note().Text; n != "" {
		t.Errorf("opening the head picker said %q", n)
	}
	drawnPortraits(t, p, "opened")
	pages := p.picker(p.ink()).Pages
	if pages < 3 {
		t.Fatalf("%d pages: too few to turn", pages)
	}
	for i := 1; i < pages; i++ {
		flowClick(f, p.picker(p.ink()).Next)
		if pg := p.picker(p.ink()).Page; pg != i {
			t.Fatalf("Next: page %d, want %d", pg, i)
		}
		drawnPortraits(t, p, "next")
	}
	for i := pages - 2; i >= 0; i-- {
		flowClick(f, p.picker(p.ink()).Prev)
		if pg := p.picker(p.ink()).Page; pg != i {
			t.Fatalf("Prev: page %d, want %d", pg, i)
		}
		drawnPortraits(t, p, "prev")
	}
	flowKey(f, keyOf(input.PgDn))
	drawnPortraits(t, p, "PgDn")
	pk := p.picker(p.ink())
	want := p.eng.p.pool.names[pk.Page*pk.PerPage+1]
	flowClick(f, pk.Cells[1])
	if p.gridPick != nil {
		t.Fatal("the picker stayed up after a pick")
	}
	if got := f.Game.Subject().PortraitHeadID; got != want {
		t.Errorf("draft head %q, want the cell clicked %q", got, want)
	}
	keep := func() {
		flowKey(f, 'k')
		for _, c := range "Zed\r" {
			flowKey(f, uint16(c))
		}
		if l := eng.BuildCharList(); !slices.Equal(l, []string{"Zed"}) {
			t.Fatalf("roster %q after Keep, want Zed once", l)
		}
	}
	keep()
	flowKey(f, 'f') // Forget's card: the head picked
	if pk = p.picker(p.ink()); pk == nil || len(pk.Portraits) != 1 || pk.Portraits[0] != poolPicture("portraits", want) {
		t.Fatalf("Forget's card does not show Zed's head %q: %+v", want, pk)
	}
	flowKey(f, keyOf(input.Esc))
	flowKey(f, 'c') // kept again under the same name, no head picked: still one roster character
	flowKey(f, 'r')
	keep()
	flowKey(f, 'a') // Add's card: the face the party's card then shows
	if pk = p.picker(p.ink()); pk == nil || len(pk.Portraits) != 1 {
		t.Fatalf("no Add picker of one: %+v", pk)
	}
	card := pk.Portraits[0]
	flowKey(f, '\r')
	if len(p.party()) != 1 {
		t.Fatalf("party of %d after Add", len(p.party()))
	}
	if face := poolPicture("portraits", defaultPortrait(p.party()[0]).PortraitHeadID); card == nil || card != face {
		t.Errorf("Add's card and the party's card show different faces")
	}
}

// The creation sheet shows the draft's combat icon, and its icon editor
// comes up (EditIcon is offered while drafting), takes a colour and a pool
// head, and Keep leaves both on the draft.
func TestCreationIconEditor(t *testing.T) {
	f, eng := flowFront(t, "")
	p := &f.P
	flowKey(f, 'c')
	flowKey(f, 'r')
	cr := p.creation()
	if cr == nil {
		t.Fatal("no creation sheet")
	}
	icons := 0
	for _, pc := range p.screenPage().pics {
		if !pc.Fill && pc.Img != nil && pc.At.In(cr.Icon.Pixels()) {
			icons++
		}
	}
	if icons != 2 {
		t.Errorf("%d combat icon poses drawn on the sheet, want 2", icons)
	}
	flowKey(f, 'i')
	ed := p.iconEditor(p.ink())
	if ed == nil || p.screenName() != "icon" {
		t.Fatalf("no icon editor: screen %q, note %q", p.screenName(), p.note().Text)
	}
	if ed.Poses[0] == nil || ed.Poses[1] == nil {
		t.Error("the editor shows no poses")
	}
	p.sendA(Action{Kind: ActIconColour, N: 2, X: 1, Y: 5})
	f.Frame(flowW, flowH)
	flowClick(f, ed.Head)
	if p.gridPick == nil || p.eng.p.pool == nil {
		t.Fatal("no icon head picker")
	}
	drawnPortraits(t, p, "icon heads")
	head := p.eng.p.pool.names[2]
	flowClick(f, p.picker(p.ink()).Cells[2])
	if ed = p.iconEditor(p.ink()); ed == nil {
		t.Fatal("the editor went with the head picker")
	}
	if ed.Colour[2][1] != 5 {
		t.Errorf("part 2's high colour %d after the head pick, want 5", ed.Colour[2][1])
	}
	flowKey(f, 'k')
	if p.icon != nil || p.creation() == nil {
		t.Fatalf("Keep: editor up %v, sheet up %v", p.icon != nil, p.creation() != nil)
	}
	m := f.Game.Subject()
	if h, _ := memberIcon(subject(m)); h != head {
		t.Errorf("draft icon head %q, want %q", h, head)
	}
	if len(m.IconColors) < 3 || m.IconColors[2]>>4 != 5 {
		t.Errorf("draft icon colours %v, want part 2's high colour 5", m.IconColors)
	}
	if _, ok := eng.State().(game.Assembling); !ok {
		t.Errorf("state %T after the editor", eng.State())
	}
}

// The icon head picker lists only heads that compose with the draft's body:
// ready poses of its size class, in pictures of its palette (the other
// class's sit off the neck, another art's do not show at all); a head picked
// leaves the body as it was, and the action pose is a picture of its own.
func TestIconHeadPickerFitsBody(t *testing.T) {
	f, _ := flowFront(t, "")
	p := &f.P
	defer switchArt(daxArt.amiga)
	switchArt(os.Getenv("FLOW_AMIGA") != "") // the other art: FLOW_AMIGA=1
	flowKey(f, 'c')
	flowKey(f, 'r')
	head, body := memberIcon(subject(f.Game.Subject()))
	ready, _ := composeIcon(head, body, nil, 0, false)
	action, ok := composeIcon(head, body, nil, 1, false)
	if !ok || slices.Equal(ready.Pix, action.Pix) {
		t.Errorf("%s over %s: the action pose is the ready one", head, body)
	}
	flowKey(f, 'i')
	flowClick(f, p.iconEditor(p.ink()).Head)
	if p.eng.p.pool == nil {
		t.Fatal("no icon head picker")
	}
	names := p.eng.p.pool.names
	if !slices.Contains(names, poolEntryID(head)) {
		t.Fatalf("the draft's own head %s is not among %v", head, names)
	}
	bp := poolPicture("combat", body).(*image.Paletted)
	for _, n := range names {
		im := poolPicture("combat", n).(*image.Paletted)
		if part, ok := poolParts()[n]; !ok || part.small != poolParts()[poolEntryID(body)].small {
			t.Errorf("%s: not a ready head of %s's size class", n, body)
		}
		if len(im.Palette) != len(bp.Palette) {
			t.Errorf("%s: %d colours, the draft's body %d", n, len(im.Palette), len(bp.Palette))
		}
	}
	i := slices.IndexFunc(names, func(n string) bool { return n != poolEntryID(head) })
	pk := p.picker(p.ink())
	if i < 0 || i >= len(pk.Cells) {
		t.Fatalf("no other head on the first page: %v", names)
	}
	flowClick(f, pk.Cells[i])
	if h, b := memberIcon(subject(f.Game.Subject())); poolEntryID(h) != names[i] || b != body {
		t.Errorf("after picking %s: icon %s over %s, want it over %s", names[i], h, b, body)
	}
}

// Every pool head composes over every pool body in both arts (class: palette
// indices OR'd as bit planes outside the sixteen colours leave the palette).
func TestPoolIconComposesInBothArts(t *testing.T) {
	flowFront(t, "")
	defer switchArt(daxArt.amiga)
	for _, amiga := range []bool{false, true} {
		switchArt(amiga)
		heads, bodies := poolEntries("combat", "combat_head_"), poolEntries("combat", "combat_body_")
		if len(heads) == 0 || len(bodies) == 0 {
			t.Fatalf("amiga %v: %d heads, %d bodies", amiga, len(heads), len(bodies))
		}
		for _, h := range heads {
			for _, b := range bodies {
				for pose := range 2 {
					pic, ok := composeIcon(h, b, nil, pose, false)
					if n := len(pic.Palette); ok && n > 0 && slices.ContainsFunc(pic.Pix, func(v uint8) bool { return int(v) >= n }) {
						t.Fatalf("amiga %v: %s over %s leaves its palette of %d", amiga, h, b, n)
					}
				}
			}
		}
	}
}

// The Forget picker's cards each show a portrait and its pages turn on the
// page drawn; charlist.txt's duplicates of a name are one roster character.
func TestForgetPicker(t *testing.T) {
	names := ""
	for _, n := range []string{"Aldric", "Brann", "Corwin", "Elowen", "Ilvan", "Maren", "Osric", "Pip", "Sylvaine", "Tilly", "Pip", "Pip"} {
		names += n + "\r\n"
	}
	f, eng := flowFront(t, names)
	p := &f.P
	if l := eng.BuildCharList(); len(l) != 10 {
		t.Fatalf("roster %q, want Pip once", l)
	}
	flowKey(f, 'f')
	pk := p.picker(p.ink())
	if pk == nil || pk.Pages < 2 {
		t.Fatalf("no paged Forget picker: %+v", pk)
	}
	drawnPortraits(t, p, "opened")
	flowClick(f, pk.Next)
	if pg := p.picker(p.ink()).Page; pg != 1 {
		t.Fatalf("Next: page %d", pg)
	}
	drawnPortraits(t, p, "next")
	flowClick(f, p.picker(p.ink()).Prev)
	drawnPortraits(t, p, "prev")
}
