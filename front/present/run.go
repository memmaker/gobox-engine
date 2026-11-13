// run.go: Front, the ready-made front end over a *classicGame: the shell
// (gobox/app, Config.Front) drives the presenter's screens with it, and the
// game's rules run on the screens' decisions.

package present

import (
	"fmt"
	"gobox/input"
	"gobox/ui/fonts"
	"log"
	"path/filepath"
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
	Game *classicGame
	P    Presenter

	// Glyph is a character's pattern in the game's font now, Font a pane
	// run's (Note.Glyph, Note.Font).
	Glyph func(byte) [8]byte
	Font  func(ch byte, upcase bool) [8]byte
	// OnStart runs as the shell starts, after the settings are loaded
	// (nil: nothing).
	OnStart func()
	ArtNew  bool // the art changed: the game publishes its page again in it (a menu waiting has no tick)

	pic     Rect       // where the picture is drawn, in device pixels
	slot    string     // the save to start from (Open)
	started bool       // the game has put its first screen up (Act)
	off     *gpu.Image // Frame's offscreen picture
	v       front.View // the Game's View, kept while fresh (view)
	fresh   bool
	drawn   string   // screenSig of the View last drawn
	held    []uint16 // keys pressed on a screen not drawn yet: they act once it is (Draw)

	viewed     bool   // GOBOX_TIMING: a View was logged
	lastScreen string // and what it showed
}

// NewFront is a Front over g, with the game's declarative data over what
// SetData handed in. It plays sound through snd (the game's own driver,
// one instance shared with its rules), or else through a Driver it builds
// from Data.Sounds and Data.AmigaSound, on the Tandy, and plays out itself.
func NewFront(g *Game, data game.Data, snd ...audio.Sound) *Front {
	gpu.Mark("NewFront")
	gameData(data)
	f := &Front{}
	if font := fonts.Get(theme.Font); font != nil { // the theme's font; a game's own glyphs replace these
		f.Glyph = func(ch byte) [8]byte { g, _ := fonts.Glyph(font, ch, true); return g }
		f.Font = func(ch byte, text bool) [8]byte { g, _ := fonts.Glyph(font, ch, text); return g }
	}
	f.chooseFonts(data.Fonts)
	gpu.Mark("NewFront: fonts")
	f.Game = &classicGame{g, &f.P}
	if len(snd) > 0 {
		f.P.Snd = snd[0]
	} else if d := newDriver(); d != nil {
		f.P.Snd = d
		audio.Output(d)
	}
	if d, ok := f.P.Snd.(*audio.Driver); ok && len(snd) > 0 && amigaMX != nil { // the game's Music-X set on the driver it was handed (newDriver attaches its own)
		d.SetAmiga(amigaMX, amigaFx)
	}
	if d, ok := f.P.Snd.(*audio.Driver); ok && adLib != nil { // the AdLib songs are the driver's soundtrack, its effects over them
		if a, err := audio.NewAdLib(*adLib); err == nil {
			f.P.AdLib = a
			d.Attach(nil, nil, a, nil, nil, nil)
			d.SetSoundtrack(true, 0, 0)
		}
	}
	if own := ownMusicDir(data); own != "" && len(snd) == 0 { // the player's own music: data/music/<trigger>
		var err error
		if ownMus, err = newOwnMusic(own, data.MusicPlaces); err != nil {
			log.Printf("own music: %v", err) // NewFront has no error to return; the music found plays
		}
		audio.Output(ownMus)
		if d, ok := f.P.Snd.(*audio.Driver); ok {
			d.SetOwnMusic(ownMus.Playing)
		}
	}
	f.P.SetEngine(f.Game)
	gpu.Mark("NewFront: sound out, done")
	return f
}

// Start takes the shell's start-up save slot (Capture.Slot).
func (f *Front) Start(slot string) {
	if f.OnStart != nil {
		f.OnStart()
	}
	f.slot = slot
}

// view is the Game's View, built once and kept until something may change
// it (stale): the shell asks for it several times a frame (Screen, Free,
// Act, Draw), and each build is the rules' whole State.
func (f *Front) view() front.View {
	if !f.fresh {
		if gpu.Timing {
			defer f.timedView(time.Now())
		}
		f.v, f.fresh = f.Game.View(), true
	}
	return f.v
}

// timedView logs the first View and each change of screen (GOBOX_TIMING).
func (f *Front) timedView(at time.Time) {
	v := f.v
	s := fmt.Sprintf("exploring %v, fight %v, party menu %v, intro %v, page %v, area %d", v.Exploring, v.Battle != nil, v.MainMenu != nil, v.Intro != nil, v.Pane != nil, v.Area)
	if !f.viewed || s != f.lastScreen {
		gpu.Mark("View (%.1f ms to build): %s", float64(time.Since(at).Microseconds())/1000, s)
	}
	f.viewed, f.lastScreen = true, s
}

// stale drops the kept View: every call that may change the game or the screens makes it.
func (f *Front) stale() { f.fresh = false }

func (f *Front) Screen() string { return front.ScreenOf(f.view(), f.P.Screens()) }
func (f *Front) Key(k uint16) {
	if !f.started { // the game opens (on its -slot) before it takes a key: the first one is not for the screen before
		f.Act()
	}
	defer f.stale()
	if f.stopWalk() {
		return
	}
	if a := actOf(k); f.P.Playing() && a != input.Space && a != input.Enter && a != input.Esc && !f.P.CameraKey(k) { // a key while the fight's events play: kept for the screen after, never dropped (Space, Enter, Esc cut the moment short)
		f.held = append(f.held, k)
		return
	}
	if v := f.view(); f.drawn != "" && screenSig(v) != f.drawn { // the screen changed and is not on show yet: the key waits for it, never acting on the one before
		f.held = append(f.held, k)
		return
	}
	f.P.Key(k, f.view())
}

// screenSig is what screen v is, as far as keys go: a key acts on the one drawn.
func screenSig(v front.View) string {
	ask := front.AskKind(-1)
	if v.Ask != nil {
		ask = v.Ask.Kind
	}
	turn := v.Battle != nil && v.Battle.Turn != nil
	return fmt.Sprint(v.Battle != nil, turn, ask, v.Inventory != nil, v.Sheet != nil, v.Camp != nil, v.Temple != nil, v.Shop != nil, v.MainMenu != nil, v.Exploring, v.Training != nil)
}
func (f *Front) Tick() { Scripted = true; f.P.Tick() } // only a script ticks: its clock is the frames

func (f *Front) Options() bool       { return f.P.Options() }
func (f *Front) UpdateOptions()      { f.stale(); f.P.UpdateOptions() }
func (f *Front) OptionsPad(k uint16) { f.stale(); f.P.OptionsPad(k) }
func (f *Front) OptionsMouse(x, y int, left, right bool, wheel float64) {
	f.stale()
	f.P.OptionsMouse(x, y, left, right, wheel)
}

func (f *Front) PadKeys() []ui.PadKey  { return PadKeys(f.P.CombatPad()) } // the binding table's buttons (bindings.go)
func (f *Front) PadStick(x, y float64) { f.stale(); f.P.PadStick(x, y) }

// Idle: the front waits for input; nothing goes on with time (a paced
// fight, rest, story or walk, a timed pause) and the 3D view stands still.
// A scripted run (-keys) gives its next key only then.
func (f *Front) Idle() bool {
	if f.P.Playing() { // the fight's events still play: a script's next key (or its end) waits for them
		return false
	}
	v := f.view()
	if v.Pace != front.PaceNone && v.Pace != front.PaceAnim || v.Ask != nil && v.Ask.Kind == front.AskPause {
		return false
	}
	return f.P.view == nil || !f.P.view.Moving()
}

func (f *Front) Free() bool { return f.P.Free(f.view()) }
func (f *Front) Flow(held uint16, ready bool) (bool, bool) {
	return f.P.Flow(held, ready)
}
func (f *Front) FreeWalk(in blobber.FreeInput) { f.P.FreeWalk(in, f.view()) }

func (f *Front) InventoryShown() bool { return f.P.InventoryShown() }
func (f *Front) InventoryMouse(x, y int, down, right bool, wheel float64) {
	f.stale()
	f.P.InventoryMouse(x, y, down, right, wheel)
}
func (f *Front) Hover(x, y int, down bool) { f.stale(); f.P.Hover(x, y, down) }

// Click is a left click at window pixel x, y. On the title or a sequence it
// is the key Enter is, on the view the front has now (the presenter's
// shownView lags before the first draw, so early clicks were lost).
func (f *Front) Click(x, y int) bool {
	if !f.started {
		f.Act()
	}
	if v := f.P.own(f.view()); v.Intro != nil || v.Sequence != nil {
		f.Key('\r')
		return true
	}
	f.stale()
	if f.stopWalk() {
		return true
	}
	return f.P.Click(x, y)
}
func (f *Front) ViewClick(x, y int) {
	f.stale()
	if !f.stopWalk() {
		f.P.ViewClick(x, y)
	}
}

// stopWalk stops the auto-walk under way (any key or click does; that input
// acts on nothing else); false: none is.
func (f *Front) stopWalk() bool {
	if f.view().Pace != front.PaceWalk {
		return false
	}
	f.P.call(func(e *classicGame) { e.StopWalk() })
	return true
}
func (f *Front) Mouse(cx, cy int, in, left, right bool, wheel int) {
	defer f.stale()
	f.P.Mouse(cx, cy, in, left, right, wheel, f.view())
}
func (f *Front) Drop(path string, cx, cy int, in bool) {
	defer f.stale()
	f.P.Drop(path, cx, cy, in, f.view())
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
		f.stale()
	}
	if f.P.Playing() { // the fight's events play first: the game waits, decisions are held
		return
	}
	as := f.P.Actions() // first: the turn's decisions become rule calls
	f.P.Pace(f.view())
	calls := f.P.Calls()
	for _, c := range calls { // the rule calls, then the actions not yet rule calls
		c(g)
	}
	if len(as) > 0 || len(calls) > 0 || f.ArtNew {
		f.stale() // the game runs: Draw builds the View again
		if f.ArtNew {
			f.P.ran++ // the picker and icon editor lay their pages out again, in the new art
		}
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
	v := f.view()
	cols, rows := 40, 25
	if v.Cells.X > 0 {
		cols, rows = v.Cells.X, v.Cells.Y
	}
	return int((x - p.X) * float64(cols) / p.W), int((y - p.Y) * float64(rows) / p.H), true
}

// Draw draws the screens, the status line's toast (View.Note) on the
// bottom row while it is fresh.
func (f *Front) Draw(dst *gpu.Image) {
	v := f.view()
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
	f.drawn = screenSig(v)
	if held := f.held; len(held) > 0 { // the keys pressed before this screen showed, now on it
		f.held = nil
		for _, k := range held {
			f.Key(k)
		}
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

var _ audio.Sound = (*audio.Driver)(nil)

// ownMusicDir is the player's own music folder: Data.OwnMusic, else beside
// Data.Files ("": none).
func ownMusicDir(d game.Data) string {
	switch {
	case d.OwnMusic != "":
		return d.OwnMusic
	case d.Files != "":
		return filepath.Join(filepath.Dir(d.Files), "music")
	}
	return ""
}
