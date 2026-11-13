package ui

import (
	"image"
	"image/color"
	"slices"
	"testing"
)

func TestMapMesh(t *testing.T) {
	var g [25][50]Square
	g[5][5], g[5][6] = Square{Tile: 1, Wall: true}, Square{Tile: 2, Wall: true}
	g[9][9].Tile = 4 // a dungeon piece on a walkable square: floor
	none := func(int) bool { return false }
	vs := mapMesh(&g, 25, none, none, nil, nil)
	// 1248 floors + two blocks: 2 tops, 6 outer sides (the shared side is gone)
	if got, want := len(vs), (1248+2+6)*6; got != want {
		t.Fatalf("%d vertices, want %d", got, want)
	}
	if k := kindOf(Square{Tile: 30, Wall: true, Height: 2}, 0, none, none, nil); k != stand {
		t.Errorf("outdoor sight blocker is %v, want stand", k)
	}
	if k := kindOf(Square{Tile: 30, Wall: true}, 0, none, none, nil); k != flat {
		t.Errorf("water is %v, want flat", k)
	}
	all := func(int) bool { return true }
	if k := kindOf(Square{Tile: 30, Wall: true}, 0, none, all, nil); k != stand {
		t.Errorf("a log is %v, want stand", k)
	}
	// a tree: the crown stands on the blocked square, the trunk below it joins it
	var o [25][50]Square
	o[3][4], o[4][4] = Square{Tile: 1, Wall: true, Height: 2}, Square{Tile: 2}
	tiles := make([]image.Image, 3)
	for i := range tiles {
		tiles[i] = image.NewRGBA(image.Rect(0, 0, tilePx, tilePx))
	}
	tiles[1].(*image.RGBA).Pix[3] = 255                                          // one opaque texel at the crown's top
	tiles[2].(*image.RGBA).Pix[(tilePx-1)*tiles[2].(*image.RGBA).Stride+3] = 255 // and the trunk's foot
	top := float32(0)
	for _, v := range mapMesh(&o, 0, none, all, tiles, nil) {
		top = max(top, v.Y)
	}
	if top < 1.5 {
		t.Errorf("tree top at %v, want the crown over the trunk", top)
	}
	// a sprite with clear rows below it sinks by them, so it stands on the floor
	if got, want := gap(tiles[1]), float32(tilePx-1)/tilePx; got != want {
		t.Errorf("gap %v, want %v", got, want)
	}
}

func TestStoneTexture(t *testing.T) {
	tile := image.NewRGBA(image.Rect(0, 0, tilePx, tilePx))
	for i := range tilePx * tilePx {
		tile.SetRGBA(i%tilePx, i/tilePx, color.RGBA{uint8(i), 50, 60, 255})
	}
	tile.SetRGBA(20, 1, color.RGBA{255, 255, 255, 255}) // in the bevel
	tile.SetRGBA(5, 5, color.RGBA{})                    // mortar
	mortar := color.RGBA{1, 2, 3, 255}
	got := stoneTexture([]image.Image{nil, nil, nil, tile}, 3, mortar)
	if got.RGBAAt(20, 1) != tile.RGBAAt(1, 20) {
		t.Errorf("bevel pixel %v, want the mirrored stone %v", got.RGBAAt(20, 1), tile.RGBAAt(1, 20))
	}
	if got.RGBAAt(5, 5) != mortar {
		t.Errorf("mortar %v, want the ground colour", got.RGBAAt(5, 5))
	}
}

func TestKeys(t *testing.T) {
	p := &CombatPresenter{}
	b := &Board{Turn: &Turn{Can: []Kind{CombatActAttack, CombatActWait, CombatActEnd, CombatActQuick}}}
	p.b = b
	p.bar.items = shownItems(b, false)
	p.bar.rects = make([]image.Rectangle, len(p.bar.items))
	type key struct {
		a Act
		k uint16
	}
	took := func(k key) []Decision { p.Key(k.a, k.k); return p.Actions() }
	for k, want := range map[key]Kind{{0, '1'}: CombatActAttack, {0, '2'}: CombatActWait, {EndTurn, 8}: CombatActEnd, {0, 'u'}: CombatActQuick, {0, 'p'}: CombatActWait} {
		if got := took(k); len(got) != 1 || got[0].Kind != want {
			t.Errorf("key %v: %+v, want kind %d", k, got, want)
		}
	}
	for _, k := range []key{{0, '5'}, {0, 'g'}, {CursorUp, 'w'}, {TurnLeft, 'Q'}, {Up, 0}} {
		if got := took(k); len(got) != 0 {
			t.Errorf("key %v: %+v, want nothing", k, got)
		}
	}
	if p.yawTo == 0 {
		t.Error("Q did not turn the camera")
	}
	b.Turn = &Turn{Aim: true}
	b.Targets, b.Units, b.Cursor = []int{3, 5}, []Unit{{ID: 3, X: 1, Y: 1}, {ID: 5, X: 4, Y: 2}}, image.Pt(1, 1)
	p.bar.items = shownItems(b, false)
	if len(p.bar.items) != 1 || p.bar.items[0].Text != "Cancel" {
		t.Errorf("aim bar: %+v", p.bar.items)
	}
	if got := took(key{Tab, 9}); len(got) != 0 || b.Cursor != image.Pt(4, 2) { // the aim cursor is front state: no game action
		t.Errorf("Tab aiming: %+v at %v, want the next target", got, b.Cursor)
	}
	if got := took(key{Enter, 13}); len(got) != 1 || got[0] != (Decision{Kind: CombatActAim, X: 4, Y: 2}) {
		t.Errorf("Enter aiming: %+v", got)
	}
	if got := took(key{Esc, 27}); len(got) != 1 || got[0].Kind != CombatActClose {
		t.Errorf("Esc aiming: %+v", got)
	}
}

func TestButtonRects(t *testing.T) {
	r := buttonRect(400, 3, 2, 2) // a hint wider than the buttons centres them
	if bw := face + 8; r.Dx() != bw || r.Min.X != (400-3*(bw+8)+8)/2+2*(bw+8) {
		t.Errorf("button 2 of 3 on a 400 wide bar: %v", r)
	}
}

// TestNDC: ndc is pick's inverse: a square's centre projects to a screen
// point that picks the square again, and the eye's own square is out of sight.
func TestNDC(t *testing.T) {
	c := camera{eye: vec{20, 9, 20}, at: vec{14, 0, 8}, fovy: 0.8, aspect: 4.0 / 3}
	for _, w := range []vec{{14.5, 0, 8.5}, {10.5, 0, 6.5}, {17.5, 0, 12.5}} {
		nx, ny, ok := c.ndc(w)
		x, z, picked := c.pick(nx, ny)
		if !ok || !picked || x != int(w[0]) || z != int(w[2]) {
			t.Errorf("%v → (%.2f, %.2f) %v → %d,%d %v", w, nx, ny, ok, x, z, picked)
		}
	}
	if _, _, ok := c.ndc(vec{26, 0, 32}); ok {
		t.Error("a point behind the eye is in sight")
	}
	if _, _, ok := (camera{}).ndc(vec{1, 0, 1}); ok {
		t.Error("the zero camera sees")
	}
}

func TestBarFocus(t *testing.T) {
	p := &CombatPresenter{}
	b := &Board{Turn: &Turn{Can: []Kind{CombatActAttack, CombatActWait, CombatActEnd}}}
	p.b = b
	p.bar.items = shownItems(b, false)
	p.bar.rects = make([]image.Rectangle, len(p.bar.items))
	if p.Key(Enter, 13); len(p.Actions()) != 0 {
		t.Error("Enter with no focus pressed a button")
	}
	for _, c := range []struct {
		a     Act
		focus int
	}{{Tab, 1}, {PgDn, 2}, {Tab, 3}, {Tab, 1}, {PgUp, 3}, {PgUp, 2}} {
		p.Key(c.a, 0)
		if p.bar.focus != c.focus {
			t.Fatalf("key %v: focus %d, want %d", c.a, p.bar.focus, c.focus)
		}
	}
	if p.Key(Enter, 13); !slices.Equal(p.Actions(), []Decision{{Kind: CombatActWait}}) {
		t.Error("Enter did not press the focused Wait")
	}
	p.bar.focus = 0
	p.bar.moveFocus(-1)
	if p.bar.focus != 3 {
		t.Errorf("back from none: focus %d, want the last", p.bar.focus)
	}
	b.Turn = nil
	p.bar.items = shownItems(b, true)
	for k, want := range map[uint16]Kind{'y': CombatActYes, 'n': CombatActNo} {
		if p.Key(0, k); !slices.Equal(p.Actions(), []Decision{{Kind: want}}) {
			t.Errorf("question key %#x: want kind %d", k, want)
		}
	}
}

// The square cursor (W A S D, the pad's d-pad): free, it starts on the
// active unit and Enter walks or attacks there as a click; aiming, it
// moves the aim and the shoulders cycle the targets.
func TestCursor(t *testing.T) {
	p := &CombatPresenter{}
	b := &Board{Turn: &Turn{}, Units: []Unit{{ID: 1, X: 5, Y: 5, Active: true, Party: true}, {ID: 2, X: 9, Y: 9}, {ID: 3, X: 2, Y: 2}},
		Reach: []image.Point{{5, 3}}}
	p.b = b
	if p.Key(CursorUp, 'w'); p.free == nil || *p.free != image.Pt(5, 4) { // 5,4 is no step: the cursor starts
		t.Fatalf("W: cursor %v, want 5,4 (north of the active unit)", p.free)
	}
	if p.Key(CursorUp, 'w'); p.free == nil || *p.free != image.Pt(5, 3) {
		t.Fatalf("W again: cursor %v, want 5,3", p.free)
	}
	if p.Key(Enter, 13); !slices.Equal(p.Actions(), []Decision{{Kind: CombatActWalk, X: 5, Y: 3}}) {
		t.Error("Enter on a reach square did not walk there")
	}
	p.Point(100, 100) // a moved mouse takes over
	if p.free != nil {
		t.Error("the mouse did not take over from the cursor")
	}
	b.Turn.Aim, b.Cursor, b.Targets = true, image.Pt(9, 9), []int{2, 3}
	if p.Key(CursorRight, 'd'); len(p.Actions()) != 0 || p.b.Cursor != image.Pt(10, 9) {
		t.Error("D did not move the aim east")
	}
	if p.Key(PgDn, 0); len(p.Actions()) != 0 || p.b.Cursor != image.Pt(9, 9) { // off every target: the first
		t.Error("RB did not aim at the next target")
	}
}

// A click on a Reach square walks there.
func TestClickReach(t *testing.T) {
	p := &CombatPresenter{b: &Board{Turn: &Turn{}, Reach: []image.Point{{3, 4}}}}
	if !p.clickSquare(image.Pt(3, 4)) || !slices.Equal(p.Actions(), []Decision{{Kind: CombatActWalk, X: 3, Y: 4}}) {
		t.Error("a click on a Reach square did not return Walk")
	}
}

// TestMapMeshPartial: a map whose standing sprites have no tile pictures
// (nil, missing, out of range) builds without a panic.
func TestMapMeshPartial(t *testing.T) {
	var g [25][50]Square
	g[3][4] = Square{Tile: 1, Wall: true, Height: 2}
	g[6][6] = Square{Tile: 99, Wall: true, Height: 2}
	all := func(int) bool { return true }
	for _, tiles := range [][]image.Image{nil, {}, make([]image.Image, 3)} {
		mapMesh(&g, 0, func(int) bool { return false }, all, tiles, nil)
	}
	var empty [25][50]Square
	mapMesh(&empty, 0, all, all, nil, nil)
}

// TestAimEnterCentres: Enter while aiming hands the square over and turns
// the view to the cursor: a square the game doesn't take leaves the aim up,
// looked at (the original's Center).
func TestAimEnterCentres(t *testing.T) {
	b := &Board{Units: []Unit{{ID: 1, X: 2, Y: 2, Active: true}}, Cursor: image.Pt(9, 4), CursorOn: true, Turn: &Turn{Aim: true}}
	p := &CombatPresenter{b: b}
	p.Key(Enter, 0)
	if p.tabTo == nil || *p.tabTo != (vec{9.5, 0, 4.5}) {
		t.Fatalf("view on %v", p.tabTo)
	}
	if got := p.Actions(); len(got) != 1 || got[0] != (Decision{Kind: CombatActAim, X: 9, Y: 4}) {
		t.Fatalf("%+v", got)
	}
}
