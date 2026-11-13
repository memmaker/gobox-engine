// control.go holds the combat map's controls: keys, the square cursor, targets, the camera's zoom, pan and chase, and picking units under the pointer.

package ui

import (
	"cmp"
	"gobox/gpu"
	"image"
	"math"
	"slices"
)

// Zoom moves the camera in (wheel up) or out by one wheel step: 15 percent
// of its distance at speed 1 (the Zoom speed option).
func (p *CombatPresenter) Zoom(in bool, speed float64) {
	p.initCamera()
	f := 1 + 0.15*speed
	if in {
		p.distTo = max(minDist, p.distTo/f)
	} else {
		p.distTo = min(maxDist, p.distTo*f)
	}
}

// Key is the XCOM 2 scheme: W A S D pan (held: see pan), the arrows move the cursor, Q and E turn,
// Home brings the camera back, 1..9 and 0 pick the bar's
// buttons, the bar's letters (hotkeys) theirs, Tab and Shift-Tab cycle the
// targets while aiming (else they and PgUp/PgDn move the bar's focus,
// which Enter presses), Enter takes the aimed one, Esc gives up (No) and
// Backspace ends the turn. It reports whether k was taken.
//
// a is what the binding table makes of the
// key, k the key itself: the digits and the bar's letters are raw input.
func (p *CombatPresenter) Key(a Act, k uint16) bool {
	p.initCamera()
	if k < 128 && k >= 'A' && k <= 'Z' {
		k += 'a' - 'A'
	}
	b := p.b
	aiming := b != nil && b.Turn != nil && b.Turn.Aim
	cycle := a == Tab || a == PgUp || a == PgDn
	back := a == PgUp || a == Tab && p.Shift
	keyed := combatBar{}
	if b != nil && !p.Picking {
		keyed = keyedItems(b, p.Asking) // the buttonless acts' keys (M, V ...), after the bar's
	}
	either := func(f func(*combatBar) *barItem) *barItem { return cmp.Or(f(&p.bar), f(&keyed)) }
	chord := either(func(b *combatBar) *barItem { return b.chordKey(k, p.Alt, p.Shift) })
	game := either(func(b *combatBar) *barItem { return b.gameKey(k) })
	table := either(func(b *combatBar) *barItem { return b.tableKey(k) })
	switch {
	case chord != nil: // the game's own chords (Alt-Q, Space) first
		p.press(*chord)
	case k == 'u' && p.Shift && p.bar.hotkey(a, k) != nil: // Shift-U: Quick for the whole party
		p.act(Decision{Kind: DecideQuick, N: 1})
	case game != nil: // the game's own hotkey wins over the camera's (Q, E, W ...)
		p.press(*game)
	case table != nil: // the remaster's letter for the same act (F Attack, U Quick), where the game gave its own
		p.press(*table)
	case p.Camera(a):
	case a >= CursorUp && a <= CursorRight && p.stepUnit(cursorDirs[a]): // the arrows step the turn's unit, as the original's
	case a >= CursorUp && a <= CursorRight: // the square cursor (the pad's d-pad and left stick send these too)
		p.moveCursor(cursorDirs[a])
	case a == Up || a == Down || a == InputLeft || a == InputRight: // W A S D pan
	case k >= '0' && k <= '9':
		if i := (int(k) - '0' + 9) % 10; i < len(p.bar.rects) { // 1 is the first button, 0 the tenth
			p.press(p.bar.items[i])
		}
	case either(func(b *combatBar) *barItem { return b.hotkey(a, k) }) != nil:
		p.press(*either(func(b *combatBar) *barItem { return b.hotkey(a, k) }))
	case cycle && aiming && len(b.Targets) > 0: // Tab, Shift-Tab, the pad's shoulders: the next, previous target
		p.cycleTarget(back)
	case a == Space:
		return p.Key(Enter, k) // Space picks, as Enter
	case a == Enter && aiming: // a square the game doesn't take leaves the aim up, the view centred on the cursor (the original's Center)
		t := vec{float64(b.Cursor.X) + .5, 0, float64(b.Cursor.Y) + .5}
		p.tabTo = &t
		p.act(Decision{Kind: DecideAim, X: b.Cursor.X, Y: b.Cursor.Y})
	case cycle && len(p.bar.rects) > 0: // Tab, Shift-Tab, the pad's shoulders (PgUp, PgDn): the focus
		d := 1
		if back {
			d = -1
		}
		p.bar.moveFocus(d)
	case a == Enter && p.bar.focus > 0 && p.bar.focus <= len(p.bar.rects):
		p.press(p.bar.items[p.bar.focus-1])
	case a == Enter && p.free != nil: // the square cursor: as a click there
		return p.clickSquare(*p.free)
	case a == Enter && p.lookCursor(): // no cursor yet: it starts on the turn's unit (Move had it), the arrows then move it
	case a == Esc && aiming:
		p.act(Decision{Kind: DecideClose})
	default:
		return false
	}
	return true
}

// Camera takes the camera's own acts (Q and E turn, Home brings it back),
// whatever the fight is doing: its events playing too. It reports whether a
// was one.
func (p *CombatPresenter) Camera(a Act) bool {
	if !CameraAct(a) {
		return false
	}
	p.initCamera()
	switch a {
	case TurnLeft:
		p.yawTo -= math.Pi / 4
	case TurnRight:
		p.yawTo += math.Pi / 4
	case Home:
		p.offset = vec{}
	}
	return true
}

// CameraAct is whether a is one of the camera's own acts (Camera).
func CameraAct(a Act) bool {
	return a == TurnLeft || a == TurnRight || a == Home
}

// cycleTarget moves the aim to the next of the turn's targets (back: the
// previous), the camera along.
func (p *CombatPresenter) cycleTarget(back bool) {
	b := p.b
	i := 0
	for j, id := range b.Targets {
		for _, u := range b.Units {
			if u.ID == id && u.X == b.Cursor.X && u.Y == b.Cursor.Y {
				i = j + 1
				if back {
					i = j - 1 + len(b.Targets)
				}
			}
		}
	}
	for _, u := range b.Units {
		if u.ID == b.Targets[i%len(b.Targets)] {
			c := p.cursor
			p.tabFrom = &c
			p.aimTo(image.Pt(u.X, u.Y))
		}
	}
}

// moveCursor moves the square cursor a square forward (d[0]) or right
// (d[1]) as the camera looks: aiming, the aim (ActPoint); else the free
// cursor, from the active unit's square, kept on the map.
func (p *CombatPresenter) moveCursor(d [2]float64) {
	b := p.b
	if b == nil || b.Turn == nil {
		return
	}
	step := p.screenStep(d)
	at := b.Cursor
	if !b.Turn.Aim {
		if p.free == nil {
			p.free = &image.Point{}
			for _, u := range b.Units {
				if u.Active {
					*p.free = image.Pt(u.X, u.Y)
				}
			}
		}
		at = *p.free
	}
	at = at.Add(step)
	if u, ok := p.activeUnit(); ok && !b.Turn.Aim && *p.free == image.Pt(u.X, u.Y) && !onBoard(at) { // the cursor on the unit, pushed off the board's edge: a step off, the flee
		p.act(Decision{Kind: DecideFlee})
		return
	}
	at = image.Pt(min(max(at.X, 0), len(p.ground[0])-1), min(max(at.Y, 0), len(p.ground)-1))
	if b.Turn.Aim {
		p.aimTo(at)
		return
	}
	*p.free = at
}

// cursorDirs are the arrows' square offsets on screen (moveCursor's d).
var cursorDirs = map[Act][2]float64{CursorUp: {1, 0}, CursorDown: {-1, 0}, CursorRight: {0, 1}, CursorLeft: {0, -1}}

// screenStep is the board offset of screen direction d as the camera turns.
func (p *CombatPresenter) screenStep(d [2]float64) image.Point {
	sin, cos := math.Sin(p.yaw), math.Cos(p.yaw)
	return image.Pt(int(math.Round(-sin*d[0]+cos*d[1])), int(math.Round(-cos*d[0]-sin*d[1])))
}

// stepUnit walks the turn's unit one square in screen direction d when the
// game lets it go there (Reach), outside an aim; it reports whether it did. A blocked step starts the look
// cursor instead, which the arrows then move (Enter walks there).
func (p *CombatPresenter) stepUnit(d [2]float64) bool {
	b := p.b
	if b == nil || b.Turn == nil || b.Turn.Aim || p.Asking || p.Picking {
		return false
	}
	for _, u := range b.Units {
		if u.Active {
			if p.free != nil && (*p.free != image.Pt(u.X, u.Y) || p.moving != u.ID) { // the look cursor (Enter started it, or it moved off the unit): the arrows move it
				return false
			}
			to := image.Pt(u.X, u.Y).Add(p.screenStep(d))
			if !onBoard(to) { // a step off the board's edge: the flee (the game asks, or refuses away from its edge)
				p.free = nil
				p.act(Decision{Kind: DecideFlee})
				return true
			}
			if p.moving == u.ID || slices.Contains(b.Reach, to) {
				p.free = nil // Move's cursor, still on the unit: the arrows step it, as the original's
				p.act(Decision{Kind: DecideWalk, X: to.X, Y: to.Y})
				return true
			}
		}
	}
	return false
}

// onBoard is whether sq is on the board.
func onBoard(sq image.Point) bool {
	return sq.In(image.Rect(0, 0, len(Board{}.Ground[0]), len(Board{}.Ground)))
}

// activeUnit is the unit whose turn it is (ok false: none).
func (p *CombatPresenter) activeUnit() (Unit, bool) {
	if p.b != nil {
		for _, u := range p.b.Units {
			if u.Active {
				return u, true
			}
		}
	}
	return Unit{}, false
}

// lookCursor starts the square cursor on the turn's unit, outside an aim,
// a question and a pick; the arrows then move it and Enter takes the
// square it is on (clickSquare). It reports whether it started.
func (p *CombatPresenter) lookCursor() bool {
	u, ok := p.activeUnit()
	if !ok || p.b.Turn == nil || p.b.Turn.Aim || p.Asking || p.Picking {
		return false
	}
	p.free, p.moving = &image.Point{u.X, u.Y}, 0
	return true
}

// targetAt: a foe a click attacks stands on sq.
func targetAt(b *Board, sq image.Point) bool {
	for _, u := range b.Units {
		if slices.Contains(b.Targets, u.ID) && slices.Contains(u.Squares, sq) {
			return true
		}
	}
	return false
}

// pan moves the camera's centre by the held W A S D or the right stick, relative to
// its heading, and reports whether any is held. A pan starts from where
// the camera is (not from the focus it was easing toward), so the first
// frames never swing the other way.
func (p *CombatPresenter) pan(dt float64, focus vec) bool {
	f, r := -p.Stick[1], p.Stick[0] // the pad's right stick
	for _, h := range [4]struct {
		k    gpu.Key
		f, r float64
	}{{gpu.KeyW, 1, 0}, {gpu.KeyS, -1, 0}, {gpu.KeyD, 0, 1}, {gpu.KeyA, 0, -1}} {
		if gpu.IsKeyPressed(h.k) {
			f, r = f+h.f, r+h.r
		}
	}
	if f == 0 && r == 0 {
		return false
	}
	if p.offset == (vec{}) {
		p.offset = p.at.sub(focus)
	}
	v := dt * p.dist * 1.2 // squares a second, faster zoomed out
	sin, cos := math.Sin(p.yaw), math.Cos(p.yaw)
	p.offset = p.offset.add(vec{-sin*f + cos*r, 0, -cos*f - sin*r}.mul(v))
	return true
}

// chase is how much of the easing toward focus this frame takes (0..1):
// all of it when the focus is out of sight, none while
// it sits inside the view's comfort box (the eye does not follow what is
// in plain sight: less motion the player did not ask for), and rising
// smoothly as it leaves the box for the edge.
func (p *CombatPresenter) chase(focus vec) float64 {
	const side, top, bottom = 0.6, 0.65, -0.3 // the box in -1..1 (the bar covers the bottom)
	nx, ny, ok := p.cam.ndc(focus)
	if !ok {
		return 1
	}
	out := max(0, math.Abs(nx)-side, ny-top, bottom-ny)
	return min(1, out/0.3)
}

func (p *CombatPresenter) initCamera() {
	if p.distTo == 0 {
		p.dist, p.distTo = 12, 12
	}
}

// Pick is the map square under screen pixel (x, y), as last drawn.
func (p *CombatPresenter) Pick(x, y int) (int, int, bool) {
	sq, hit := p.pickPlane(x, y)
	return sq.X, sq.Y, hit && onBoard(sq)
}

// pickPlane is the square under screen pixel (x, y) on the board's plane,
// on the board or past its edge (hit false: the pixel shows no ground).
func (p *CombatPresenter) pickPlane(x, y int) (image.Point, bool) {
	r := p.screen
	if r.Empty() || !image.Pt(x, y).In(r) {
		return image.Point{}, false
	}
	sx, sz, hit := p.cam.plane(2*float64(x-r.Min.X)/float64(r.Dx())-1, 1-2*float64(y-r.Min.Y)/float64(r.Dy()))
	return image.Pt(sx, sz), hit
}

// unitCentre is the middle of the squares u stands on (its footprint).
func unitCentre(u Unit) [2]float64 {
	if len(u.Squares) == 0 {
		return [2]float64{float64(u.X) + .5, float64(u.Y) + .5}
	}
	var c [2]float64
	for _, s := range u.Squares {
		c[0] += float64(s.X) + .5
		c[1] += float64(s.Y) + .5
	}
	return [2]float64{c[0] / float64(len(u.Squares)), c[1] / float64(len(u.Squares))}
}

// footprint is how many squares u's footprint spans across (x) and deep (z).
func footprint(u Unit) (w, d float32) {
	if len(u.Squares) == 0 {
		return 1, 1
	}
	r := image.Rectangle{u.Squares[0], u.Squares[0].Add(image.Pt(1, 1))}
	for _, s := range u.Squares {
		r = r.Union(image.Rectangle{s, s.Add(image.Pt(1, 1))})
	}
	return float32(r.Dx()), float32(r.Dy())
}

// iconSize is the size the figure's footprint follows: the built icon's
// (IconSize), else the picture's own.
func iconSize(u Unit) image.Point {
	if u.IconSize != (image.Point{}) {
		return u.IconSize
	}
	return u.Icon.Bounds().Size()
}

// aimTo puts the aim cursor on square at: front state, the game is not
// told (Enter or a click sends the ActAim).
func (p *CombatPresenter) aimTo(at image.Point) {
	p.aimAt = &at
	if p.b != nil {
		p.b.Cursor, p.b.CursorOn, p.b.CursorAt = at, true, []image.Point{at}
		for _, u := range p.b.Units {
			if slices.Contains(u.Squares, at) {
				p.b.CursorAt = u.Squares
			}
		}
	}
}
