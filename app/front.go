package app

import (
	"math"
	"slices"
	"strings"
	"time"

	"gobox/blobber"
	"gobox/front"
	"gobox/gpu"
	"gobox/input"
	"gobox/ui"
)

// Front is a game's screens as the shell's input reader drives them
// (Config.Front): the shell reads the keyboard, the mouse and the pads,
// passes them through the latch (front.Latch) and hands the screens their
// keys, clicks and No Grid walk. Key codes are the BIOS's (ui.BIOSKey).
type Front interface {
	Screen() string // names the screen shown: a change latches the inputs held (front.Latch)
	Key(k Key)      // a key code for the screens
	Act()           // a tick's end: the game runs on the screens' decisions
	Tick()          // a scripted frame: the screens' own clock
	Idle() bool     // the screens wait for input (a script's next key goes in)
	Draw(dst *gpu.Image)

	// the options page (the renderer's): it has the keyboard and the pads while open
	Options() bool
	UpdateOptions()
	OptionsPad(k Key)
	OptionsMouse(x, y int, left, right bool, wheel float64)

	PadKeys() []ui.PadKey  // the pad's buttons now (the binding table's)
	PadStick(x, y float64) // the right stick

	Free() bool                                   // No Grid's view is up: W A S D and the arrows walk
	Flow(held Key, ready bool) (send, paced bool) // fluid movement: the held move key, ask whether to send it
	FreeWalk(in blobber.FreeInput)

	InventoryShown() bool
	InventoryMouse(x, y int, down, right bool, wheel float64)
	Hover(x, y int, down bool)
	Click(x, y int) bool                     // a click the screens took (a button, a menu item)
	ViewClick(x, y int)                      // No Grid: a click on the view
	Cell(x, y float64) (cx, cy int, in bool) // the text cell at window pixel x, y
	Mouse(cx, cy int, in, left, right bool, wheel int)
	Drop(path string, cx, cy int, in bool)
}

// Art is a game with two arts (e.g. two ports' pictures), cross-faded when
// the one asked for changes (F12 toggles it).
type Art struct {
	Want   func() bool // the art asked for
	Toggle func()      // F12: ask for the other
	Set    func(bool)  // the game now shows this art
}

// artFade is how long the pictures take to cross-fade from one art to the other.
const artFade = 600 * time.Millisecond

// frontGame is the Game the shell makes of a Front.
type frontGame struct {
	f   Front
	cfg Config
	tps int

	pads    []gpu.GamepadID // input buffers, reused each tick
	pressed []gpu.Key
	chars   []rune

	stickDir  uint16 // left stick direction being held
	stickHeld int    // ticks it has been held

	wheel     float64     // wheel movement not yet sent as a step
	rightOnce bool        // \M: the right button pressed this tick
	leftOnce  bool        // \C: the left button
	leftHold  bool        // \B: the left button held (a drag)
	pointer   int         // Capture.Pointer's position in use
	latch     front.Latch // the inputs held through a screen change: ignored until let go
	look      [2]float64  // No Grid: the turn and pitch the mouse and the right stick gave this tick
	drag      *drag       // No Grid: the left button held to look, nil while up
	mlook     bool        // No Grid: mouselook is on (right click): the cursor captured, the mouse looks
	art       bool        // the art shown (Art.Set)
	old       *gpu.Image  // the last frame in the other art, faded out over the new one
	fade      time.Time   // when the art changed
}

func (g *frontGame) scripted() bool { return g.cfg.Capture.Scripted() }

func (g *frontGame) Update() {
	if g.scripted() { // the user's input never reaches a capture or a measurement, only the scripted keys (Step)
		if len(g.cfg.Capture.Pointer) > 0 || g.leftOnce || g.rightOnce || g.leftHold { // a scripted click needs no -pointer: on the title any spot steps it
			g.mouse()
		}
		return
	}
	g.latch.See(g.f.Screen(), g.held())
	if g.f.Options() { // the renderer's options page has the keyboard and the pads
		for _, id := range gpu.AppendGamepadIDs(g.pads[:0]) {
			g.gamepad(id, g.f.OptionsPad)
		}
		g.f.UpdateOptions()
		if g.f.Options() {
			mx, my := gpu.CursorPosition()
			_, dy := gpu.Wheel()
			g.f.OptionsMouse(mx, my, gpu.IsMouseButtonJustPressed(gpu.MouseButtonLeft), gpu.IsMouseButtonJustPressed(gpu.MouseButtonRight), dy)
		}
		return
	}
	g.keys()
	for _, id := range gpu.AppendGamepadIDs(g.pads[:0]) {
		g.gamepad(id, g.f.Key)
		x, y := gpu.StandardGamepadAxisValue(id, gpu.StandardGamepadAxisRightStickHorizontal), gpu.StandardGamepadAxisValue(id, gpu.StandardGamepadAxisRightStickVertical)
		g.f.PadStick(x, y)
		if math.Hypot(x, y) > 0.25 { // No Grid: the right stick looks
			g.look[0] += x * freeStick
			g.look[1] -= y * freeStick
		}
	}
	g.mouse()
	g.free()
	for _, d := range gpu.AppendDroppedFiles(nil) {
		cx, cy, in := g.f.Cell(float64(d.X), float64(d.Y))
		g.f.Drop(d.Path, cx, cy, in)
	}
	g.f.Act()
}

// Step is a scripted frame: the screens tick once a frame, and k is the script's key due now.
func (g *frontGame) Idle() bool { return g.f.Idle() }

func (g *frontGame) Step(k Key) {
	g.f.Tick()
	switch k {
	case 0:
	case Right, Left: // \M, \C: a click at the pointer
		g.rightOnce, g.leftOnce = k == Right, k == Left
	case Next: // \P: the pointer moves on to the next of Capture.Pointer
		g.pointer++
	case Hold: // \B: the left button down, or up again (\B\P\B drags)
		g.leftHold = !g.leftHold
	case Wait: // \w
	default:
		if g.f.Options() { // the Esc menu reads its keys as the pad's, as Update hands them
			g.f.OptionsPad(k)
			g.f.UpdateOptions()
			break
		}
		g.f.Key(k)
	}
	g.f.Act() // the screens' own decisions too: a paced pause ends
}

// Held arrow keys repeat like the PC keyboard did (typematic), but only
// while the game waits for input, so walking stops when the key is let go.
var repeatKeys = []gpu.Key{
	gpu.KeyArrowUp, gpu.KeyArrowDown, gpu.KeyArrowLeft, gpu.KeyArrowRight,
	gpu.KeyKP8, gpu.KeyKP2, gpu.KeyKP4, gpu.KeyKP6,
	gpu.KeyPageUp, gpu.KeyPageDown,
}

func (g *frontGame) repeats(d int) bool { return input.Repeat(d, g.tps) }

// keys passes the keyboard to the screens.
func (g *frontGame) keys() {
	free := g.f.Free() // No Grid: the arrows and W A S D walk (free)
	paced := free
	// fluid movement: the view asks for the held move as the last one ends.
	// A key counts as held only after holdTicks (200 ms): a tap (even one a
	// little longer than half a turn) is one move, never two
	var held uint16
	for _, k := range repeatKeys[:8] {
		if d := gpu.KeyPressDuration(k); g.latch.Down(k, d > 0) && d >= g.tps/5 {
			held, _ = ui.BIOSKey(k)
		}
	}
	if !free {
		var send bool
		send, paced = g.f.Flow(held, true)
		if send {
			g.f.Key(held)
		}
	}
	for i, k := range repeatKeys {
		if i < 8 && paced {
			continue
		}
		if d := gpu.KeyPressDuration(k); g.latch.Down(k, d > 1) && g.repeats(d) {
			code, _ := ui.BIOSKey(k)
			g.f.Key(code)
		}
	}
	for _, k := range gpu.AppendJustPressedKeys(g.pressed[:0]) {
		if free && slices.Contains(repeatKeys[:8], k) { // No Grid walks and turns with them (free)
			continue
		}
		if code, ok := ui.BIOSKey(k); ok {
			g.f.Key(code)
		} else if k == gpu.KeyF12 && gpu.IsKeyPressed(gpu.KeyShift) {
			g.f.Key(ui.ShiftF12) // the wall style
		} else if k == gpu.KeyF12 && g.cfg.Art != nil { // the other art (Draw fades it in)
			g.cfg.Art.Toggle()
			if blobber.OnSettings != nil {
				blobber.OnSettings(blobber.Settings)
			}
		} else if k == gpu.KeyF && gpu.IsKeyPressed(gpu.KeyControl) && gpu.IsKeyPressed(gpu.KeyShift) {
			if g.cfg.FontPrev != 0 {
				g.f.Key(g.cfg.FontPrev) // Ctrl-F next font, Ctrl-Shift-F the one before
			}
		} else if k >= gpu.KeyA && k <= gpu.KeyZ && gpu.IsKeyPressed(gpu.KeyControl) {
			g.f.Key(uint16(k-gpu.KeyA) + 1) // Ctrl-letter: ASCII 1..26
		}
	}
	// a Ctrl combination types no letter of its own (Ctrl+Alt is AltGr: it does)
	if gpu.IsKeyPressed(gpu.KeyControl) && !gpu.IsKeyPressed(gpu.KeyAlt) {
		return
	}
	for _, r := range gpu.AppendInputChars(g.chars[:0]) {
		if free && strings.ContainsRune("wasdWASD", r) {
			continue
		}
		if r < 128 {
			g.f.Key(uint16(r))
		}
	}
}

// drag is the left button held on No Grid's view, and how far the mouse
// has moved since (a click moves less than 6 pixels).
type drag struct{ moved float64 }

// No Grid's look speeds: radians per pointer pixel dragged, per tick at the stick's full tilt.
const freeMouse, freeStick = 0.005, 0.05

// free gives No Grid's camera this tick: W A S D and the arrows walk and
// turn, the look the mouse and the stick gave turns and pitches.
func (g *frontGame) free() {
	var in blobber.FreeInput
	if g.f.Free() {
		key := func(ks ...gpu.Key) float64 {
			for _, k := range ks {
				if g.latch.Down(k, gpu.IsKeyPressed(k)) {
					return 1
				}
			}
			return 0
		}
		in.Forward = key(gpu.KeyW, gpu.KeyArrowUp, gpu.KeyKP8) - key(gpu.KeyS, gpu.KeyArrowDown, gpu.KeyKP2)
		in.Right = key(gpu.KeyD) - key(gpu.KeyA)
		in.Turn = (key(gpu.KeyArrowRight, gpu.KeyKP6) - key(gpu.KeyArrowLeft, gpu.KeyKP4)) * 2.6 / float64(g.tps)
		in.Turn, in.Pitch = in.Turn+g.look[0], g.look[1]
	}
	g.look = [2]float64{}
	g.f.FreeWalk(in)
}

// wheelDamp is how much wheel movement makes one step after a scroll's first.
const wheelDamp = 4

// mouse passes the pointer's text cell, clicks and wheel steps to the screens.
func (g *frontGame) mouse() {
	mx, my := gpu.CursorPosition()
	if pts := g.cfg.Capture.Pointer; g.cfg.Capture.Shot != "" && len(pts) > 0 { // a capture's pointer (window pixels)
		p := pts[min(g.pointer, len(pts)-1)]
		mx, my = gpu.ScreenPos(p[0], p[1]) // as the cursor's: a HiDPI window's screen has more pixels than its points
	}
	leftHeld := g.latch.Down(gpu.MouseButtonLeft, gpu.IsMouseButtonPressed(gpu.MouseButtonLeft)) || g.leftHold
	if g.f.InventoryShown() { // its own clicks and wheel; keys still go to the game (a capture's \C is down for one tick: a click)
		_, dy := gpu.Wheel()
		g.f.InventoryMouse(mx, my, leftHeld || g.leftOnce, gpu.IsMouseButtonJustPressed(gpu.MouseButtonRight) || g.rightOnce, dy)
		g.leftOnce, g.rightOnce = false, false
		return
	}
	g.f.Hover(mx, my, leftHeld)
	left := gpu.IsMouseButtonJustPressed(gpu.MouseButtonLeft) || g.leftOnce
	g.leftOnce = false
	if left && g.f.Click(mx, my) { // an action bar button, a menu item, the combat map
		return
	}
	cx, cy, in := g.f.Cell(float64(mx), float64(my))
	_, dy := gpu.Wheel()
	switch {
	case dy == 0:
		g.wheel = 0 // idle: the next scroll starts afresh
	case g.wheel == 0: // a scroll's first tick steps at once, so a mouse wheel notch is one step
		g.wheel = math.Copysign(1, dy) + dy/wheelDamp
	default: // a trackpad streams many small deltas a gesture: damp them
		g.wheel += dy / wheelDamp
	}
	steps := int(g.wheel)
	g.wheel -= float64(steps)
	if g.wheel == 0 && dy != 0 {
		g.wheel = math.Copysign(1e-9, dy) // still mid-scroll
	}
	right := gpu.IsMouseButtonJustPressed(gpu.MouseButtonRight) || g.rightOnce
	g.rightOnce = false
	free := g.f.Free()
	if g.mlook && !free { // a menu or another screen came up: the pointer is back
		g.mlook = false
		gpu.SetCursorCaptured(false)
	}
	if free && right && !g.scripted() { // No Grid: a right click turns mouselook on or off (the command menu has Options)
		g.mlook, right = !g.mlook, false
		gpu.SetCursorCaptured(g.mlook)
	}
	if g.mlook { // the cursor hidden and still: the mouse looks; a left click opens the command menu
		dx, dy := gpu.MouseDelta()
		g.look[0] += dx * freeMouse
		g.look[1] -= dy * freeMouse
		if left {
			g.f.ViewClick(mx, my)
		}
		return
	}
	if free || g.drag != nil { // No Grid: a left drag looks too; a click without one is still a click, on release
		left = left && g.scripted()
		switch held := g.latch.Down(gpu.MouseButtonLeft, gpu.IsMouseButtonPressed(gpu.MouseButtonLeft)); {
		case held && g.drag == nil:
			g.drag = &drag{}
		case held:
			dx, dy := gpu.MouseDelta() // the mouse's own movement: the captured cursor stands still
			g.look[0] += dx * freeMouse
			g.look[1] -= dy * freeMouse
			g.drag.moved += math.Abs(dx) + math.Abs(dy)
			if g.drag.moved >= 6 { // a drag, not a click: the cursor hides and stays put, so the look never stops at the screen's edge
				gpu.SetCursorCaptured(true)
			}
		case g.drag != nil:
			gpu.SetCursorCaptured(false)
			if g.drag.moved < 6 {
				g.f.ViewClick(mx, my)
			}
			g.drag = nil
		}
	}
	g.f.Mouse(cx, cy, in, left, right, steps)
}

// gamepad sends the pad's keys (the binding table's, with repeat) to send.
func (g *frontGame) gamepad(id gpu.GamepadID, send func(Key)) {
	if !gpu.IsStandardGamepadLayoutAvailable(id) {
		return
	}
	pad := g.f.PadKeys()
	for _, p := range pad {
		d := gpu.StandardGamepadButtonPressDuration(id, p.B)
		if !g.latch.Down(padButton{id, p.B}, d > 0) {
			continue
		}
		if d == 1 || p.Repeat && d > 1 && g.repeats(d) {
			send(p.Key)
		}
	}
	// Left stick as a D-pad.
	x := gpu.StandardGamepadAxisValue(id, gpu.StandardGamepadAxisLeftStickHorizontal)
	y := gpu.StandardGamepadAxisValue(id, gpu.StandardGamepadAxisLeftStickVertical)
	var b gpu.StandardGamepadButton = -1
	switch {
	case y < -0.6:
		b = gpu.StandardGamepadButtonLeftTop
	case y > 0.6:
		b = gpu.StandardGamepadButtonLeftBottom
	case x < -0.6:
		b = gpu.StandardGamepadButtonLeftLeft
	case x > 0.6:
		b = gpu.StandardGamepadButtonLeftRight
	}
	dir := uint16(0) // the stick sends what the d-pad's button does
	for _, p := range pad {
		if p.B == b {
			dir = p.Key
		}
	}
	if dir != g.stickDir {
		g.stickDir, g.stickHeld = dir, 0
	}
	if dir != 0 && g.latch.Down(padStick{id, dir}, true) {
		g.stickHeld++
		if g.repeats(g.stickHeld) {
			send(dir)
		}
	}
}

func (g *frontGame) Draw(dst *gpu.Image) {
	g.f.Draw(dst)
	g.crossFade(dst)
}

// crossFade switches the game's art when another is asked for (F12, the
// Graphics page), keeping this frame to fade out over the next ones: only
// the pictures differ, so only they blend from one art to the other.
func (g *frontGame) crossFade(dst *gpu.Image) {
	a := g.cfg.Art
	if a == nil {
		return
	}
	if t := time.Since(g.fade); g.old != nil && t < artFade {
		var op gpu.DrawImageOptions
		op.ColorScale.ScaleAlpha(1 - float32(t)/float32(artFade))
		dst.DrawImage(g.old, &op)
	}
	if g.art == a.Want() {
		return
	}
	g.art = a.Want()
	w, h := dst.Size()
	gpu.Sized(&g.old, w, h)
	g.old.Clear()
	g.old.DrawImage(dst, nil)
	g.fade = time.Now()
	a.Set(g.art)
}

// padButton and padStick are a pad's held button and stick direction, as
// the latch knows them.
type (
	padButton struct {
		id gpu.GamepadID
		b  gpu.StandardGamepadButton
	}
	padStick struct {
		id  gpu.GamepadID
		dir uint16
	}
)

// held is every input held now: keys, mouse buttons, pad buttons and stick
// directions (the latch's).
func (g *frontGame) held() []any {
	var h []any
	for _, k := range gpu.AppendPressedKeys(nil) {
		h = append(h, k)
	}
	for _, b := range []gpu.MouseButton{gpu.MouseButtonLeft, gpu.MouseButtonRight, gpu.MouseButtonMiddle} {
		if gpu.IsMouseButtonPressed(b) {
			h = append(h, b)
		}
	}
	for _, id := range gpu.AppendGamepadIDs(nil) {
		for b := gpu.StandardGamepadButton(0); b <= gpu.StandardGamepadButtonCenterCenter; b++ {
			if gpu.IsStandardGamepadButtonPressed(id, b) {
				h = append(h, padButton{id, b})
			}
		}
	}
	if g.stickDir != 0 {
		for _, id := range gpu.AppendGamepadIDs(nil) {
			h = append(h, padStick{id, g.stickDir})
		}
	}
	return h
}
