// run.go: Front, the ready-made front end over a front.Game: the shell
// (gobox/app, Config.Front) drives the presenter's screens with it, and the
// game's rules run on the screens' decisions.

package present

import (
	"gobox/ui/fonts"
	"time"

	"gobox/amiga"
	"gobox/audio"
	"gobox/blobber"
	"gobox/front"
	"gobox/game"
	"gobox/gpu"
	"gobox/ui"
)

// Front is an app.Front: the presenter's screens over Game.
type Front struct {
	Game front.Game
	P    Presenter

	// Glyph is a character's pattern in the game's font now, Font a pane
	// run's (Note.Glyph, Note.Font).
	Glyph func(byte) [8]byte
	Font  func(ch byte, upcase bool) [8]byte
	// OnStart runs as the shell starts, after the settings are loaded
	// (nil: nothing).
	OnStart func()
	ArtNew  bool // the art changed: the game publishes its page again in it (a menu waiting has no tick)

	pic     Rect   // where the picture is drawn, in device pixels
	slot    string // the save to start from (Open)
	started bool   // the game has put its first screen up (Act)
}

// NewFront is a Front over g, with the game's declarative data over what
// SetData handed in. It plays sound through snd (the game's own driver,
// one instance shared with its rules), or else through a Driver it builds
// from Data.Sounds and Data.AmigaSound, on the Tandy, and plays out itself.
func NewFront(g front.Game, data game.Data, snd ...front.Sound) *Front {
	gameData(data)
	f := &Front{}
	if font := fonts.Get(theme.Font); font != nil { // the theme's font; a game's own glyphs replace these
		f.Glyph = func(ch byte) [8]byte { g, _ := fonts.Glyph(font, ch, true); return g }
		f.Font = func(ch byte, text bool) [8]byte { g, _ := fonts.Glyph(font, ch, text); return g }
	}
	f.chooseFonts(data.Fonts)
	f.Game = WithRules(g, &f.P)
	if len(snd) > 0 {
		f.P.Snd = snd[0]
	} else if d := newDriver(); d != nil {
		f.P.Snd = d
		audio.Output(d)
	}
	if d, ok := f.P.Snd.(*audio.Driver); ok && adLib != nil { // the AdLib songs are the driver's soundtrack, its effects over them
		if a, err := audio.NewAdLib(*adLib); err == nil {
			f.P.AdLib = a
			d.Attach(nil, nil, a, nil, nil, nil)
			d.SetSoundtrack(true, 0, 0)
		}
	}
	f.P.SetEngine(f.Game)
	return f
}

// Start takes the shell's start-up save slot (Capture.Slot).
func (f *Front) Start(slot string) {
	if f.OnStart != nil {
		f.OnStart()
	}
	f.slot = slot
}

func (f *Front) Screen() string { return front.ScreenOf(f.Game.View(), f.P.Screens()) }
func (f *Front) Key(k uint16) {
	if !f.started { // the game opens (on its -slot) before it takes a key: the first one is not for the screen before
		f.Act()
	}
	f.P.Key(k, f.Game.View())
}
func (f *Front) Tick() { f.P.Tick() }

func (f *Front) Options() bool       { return f.P.Options() }
func (f *Front) UpdateOptions()      { f.P.UpdateOptions() }
func (f *Front) OptionsPad(k uint16) { f.P.OptionsPad(k) }
func (f *Front) OptionsMouse(x, y int, left, right bool, wheel float64) {
	f.P.OptionsMouse(x, y, left, right, wheel)
}

func (f *Front) PadKeys() []ui.PadKey  { return PadKeys(f.P.CombatPad()) } // the binding table's buttons (bindings.go)
func (f *Front) PadStick(x, y float64) { f.P.PadStick(x, y) }

func (f *Front) Free() bool                                { return f.P.Free(f.Game.View()) }
func (f *Front) Flow(held uint16, ready bool) (bool, bool) { return f.P.Flow(held, ready) }
func (f *Front) FreeWalk(in blobber.FreeInput)             { f.P.FreeWalk(in, f.Game.View()) }

func (f *Front) InventoryShown() bool { return f.P.InventoryShown() }
func (f *Front) InventoryMouse(x, y int, down, right bool, wheel float64) {
	f.P.InventoryMouse(x, y, down, right, wheel)
}
func (f *Front) Hover(x, y int, down bool) { f.P.Hover(x, y, down) }
func (f *Front) Click(x, y int) bool       { return f.P.Click(x, y) }
func (f *Front) ViewClick(x, y int)        { f.P.ViewClick(x, y) }
func (f *Front) Mouse(cx, cy int, in, left, right bool, wheel int) {
	f.P.Mouse(cx, cy, in, left, right, wheel, f.Game.View())
}
func (f *Front) Drop(path string, cx, cy int, in bool) {
	f.P.Drop(path, cx, cy, in, f.Game.View())
}

// Act runs the game on the screens' decisions (Answer, then Settle, which
// returns once it waits again), and lets the presenter pace what goes on
// with time (Pace). The fight's events Settle hands back are played.
func (f *Front) Act() {
	g := f.Game
	if !f.started { // the title, on the render loop's (locked) thread that runs the game
		f.started = true
		g.Open(f.slot)
		f.P.Play(g.Settle().Events)
	}
	if f.P.Playing() { // the fight's events play first: the game waits, decisions are held
		return
	}
	as := f.P.Actions() // first: the turn's decisions become rule calls
	f.P.Pace(g.View())
	calls := f.P.Calls()
	for _, c := range calls { // the rule calls, then the actions not yet rule calls
		c(g)
	}
	if len(as) > 0 || len(calls) > 0 || f.ArtNew {
		f.ArtNew = false
		for _, a := range as {
			g.Answer(a)
		}
		f.P.Play(g.Settle().Events)
	}
}

// Cell is the text cell at window pixel x, y: the picture's, or the presenter's.
func (f *Front) Cell(x, y float64) (int, int, bool) {
	p := f.pic
	if p.W <= 0 {
		return f.P.Cell(int(x), int(y))
	}
	if x < p.X || y < p.Y || x >= p.X+p.W || y >= p.Y+p.H {
		return 0, 0, false
	}
	v := f.Game.View()
	cols, rows := 40, 25
	if v.Cells.X > 0 {
		cols, rows = v.Cells.X, v.Cells.Y
	}
	return int((x - p.X) * float64(cols) / p.W), int((y - p.Y) * float64(rows) / p.H), true
}

// Draw draws the screens, the status line's toast (View.Note) on the
// bottom row while it is fresh.
func (f *Front) Draw(dst *gpu.Image) {
	v := f.Game.View()
	n := v.Note
	if f.P.fontNote != "" && time.Since(f.P.fontAt) < 2*time.Second { // Ctrl-F's font name
		n.Text, n.At = f.P.fontNote, f.P.fontAt
	}
	note := Note{Text: n.Text, Show: time.Since(n.At) < time.Second+time.Duration(len(n.Text))*80*time.Millisecond, Glyph: f.Glyph, Font: f.Font}
	if f.P.Inventory(dst, note, v) { // the inventory screen
		f.pic = Rect{}
	} else {
		f.pic = f.P.Draw(dst, note, v)
	}
}

// newDriver is the Driver over Data.Sounds (nil: none), hooked so sounds play on.
func newDriver() *audio.Driver {
	if sounds == nil {
		return nil
	}
	d, err := audio.NewDriver(sounds, 'T')
	if err != nil {
		return nil
	}
	var am audio.Amiga
	var fx []int
	if amigaMX != nil {
		am, fx = amigaMX, amigaFx
	} else if len(amigaSound) == amiga.End-amiga.Base {
		am = amiga.New(amigaSound)
	}
	d.Attach(am, fx, nil, nil, nil, nil)
	d.Wait(0)
	return d
}

var _ front.Sound = (*audio.Driver)(nil)
