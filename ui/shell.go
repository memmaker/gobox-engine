// shell.go is the host window (Play): the frame loop, fullscreen and the
// window's shape, and a scripted capture's clock and screenshots.

package ui

import (
	"fmt"
	"gobox/asset"
	"image"
	"log"
	"os"
	"strings"
	"time"

	"gobox/gpu"
)

// shellGame is what a game gives the shell.
type shellGame interface {
	// Update is one tick of the game (TPS a second). During a scripted run
	// (Capture.Scripted) the user's input must not reach the game.
	Update()
	// Step is one frame of a scripted run, before Draw: k is the script's
	// key due this frame, 0 for none (ScriptRight, ScriptLeft, ScriptNext,
	// ScriptWait or a game key code from Capture.Keys).
	Step(k Key)
	// Idle: the game waits for input (no fade, pause, walk or rest going
	// on). A script's next key goes in only then.
	Idle() bool
	// Draw draws a frame into dst (device pixels, the whole window).
	Draw(dst *gpu.Image)
}

// Key is one key of a script (Capture.Keys): a game key code or one
// of the shell's own.
type Key = uint16

// The script's own keys: \M and \C a right and a left click at the
// pointer, \P the pointer on to the next position, \w a quarter second of
// nothing.
const ScriptRight, ScriptLeft, ScriptNext, ScriptWait, ScriptHold, ScriptSnap Key = 0xFF01, 0xFF02, 0xFF03, 0xFF04, 0xFF05, 0xFF06

// ScriptTimed is a \t<ms> wait: the low 12 bits are frames (60 Hz), taken at once
// and counted whether or not the game is idle, so a blow can be shot mid-air.
const ScriptTimed Key = 0xE000

// ScriptAlt is a \A<c> chord: the low byte c typed with Alt held.
const ScriptAlt Key = 0xD000

// Ticks is the least frames between a script's keys: a quarter second at
// 60 Hz. A key waits on past it until the game is Idle, at most MaxWait.
const Ticks, MaxWait = 15, 40 * Ticks

// shellTheme is the game's look of the window.
type shellTheme struct {
	Title string
	W, H  int // the window's size at scale 1, and its shape (16:10: the title pictures fill it)
}

// WindowMode is how the window opens.
type WindowMode struct {
	Scale      int  // the window is Theme.W x Theme.H times Scale
	Fullscreen bool // the screen at its own resolution instead of a window
	Window     bool // a window whatever Fullscreen or the Esc menu says
}

// Capture is a scripted run: keys in, frames out, the user's input ignored.
type Capture struct {
	Shot    string     // FILE.png: the window after the keys, then exit; with %d every frame while the keys play
	Keys    []Key      // the script (cmd/gobox parseKeys)
	Freeze  bool       // time stands still after the keys
	Quit    bool       // the run ends after the keys
	Window  [2]int     // W, H (points): a window of that size and any shape; 0: Theme's
	GPU     gpu.Config // gpu's measuring and capture switches
	Slot    string     // a save slot to start from instead of the game's own start ("": none); passed to a starter
	Pointer [][2]int   // with Shot: the pointer at these window pixels, one after another (Next)
	PNG     string     // a picture a dev scene uses (e.g. a portrait); the game's to read
	Dev     DevFlags   // every game's debug switches
}

// DevFlags are the debug switches every game gets (Felix 2026-11-12: the
// games' own -fight, 'J', -ecltrace, -saves and -at, once).
type DevFlags struct {
	Fight     int    // a fight with three of this monster (the area's MON file) at the first adventuring prompt (0: none)
	FreeTrain bool   // the training hall trains everyone free, experience or not (sotsb's 'J', POR's STING debug)
	ECLTrace  bool   // each script command and its operands to stderr
	Saves     string // the save directory ("": the game's)
	At        *At    // enter a script block there at the first adventuring prompt (nil: none)
}

// At is a script block and a position in it (-at BLOCK,X,Y).
type At struct{ Block, X, Y int }

// Scripted says whether the run is scripted: the user's input is ignored.
func (c Capture) Scripted() bool { return c.Shot != "" || c.Freeze || c.Quit }

// starter is a Game told before the first tick which save slot to start
// from (Capture.Slot, "" none); what that means is the game's.
type starter interface{ Start(slot string) }

// shellConfig is the app: the game and how it runs.
type shellConfig struct {
	Game     shellGame // or Front: the shell reads the input for the game's screens
	Front    frontPort
	Art      *Arts      // with Front: two arts, cross-faded (nil: one)
	FontPrev Key        // with Front: the code Ctrl-Shift-F sends (0: none)
	Menu     *shellMenu // the Esc menu's settings (nil: blobber's own)
	Theme    shellTheme
	Settings WindowMode
	Capture  Capture
	TPS      int // Update calls a second; 0: 60
}

// shell runs a game on gpu.
type shell struct {
	cfg    shellConfig
	ticks  int  // scripted frames
	fed    int  // the script's keys given so far
	gap    int  // frames since the last key
	frames int  // frames saved, Shot with %d
	snap   bool // a \S is due: this frame is saved
	until  int  // ticks: a \t<ms> wait runs until then
	snaps  int  // \S shots taken
}

// Play opens the window and runs the started game l in it until it ends:
// the one host loop every game runs in. w is how the window opens, c a
// scripted run's capture (zero: none).
func Play(l Launch, w WindowMode, c Capture) error {
	Scripted = c.Scripted()
	return run(shellConfig{
		Front:    l.Front,
		Art:      l.Art,
		FontPrev: l.FontPrev,
		Menu: &shellMenu{
			Prefs:    prefs{Dir: l.SettingsDir},
			Extra:    l.Extra,
			Rows:     l.Rows,
			Apply:    l.Apply,
			SaveLoad: l.SaveLoad,
			MainMenu: l.MainMenu,
		},
		Theme:    shellTheme{Title: l.Title, W: l.W, H: l.H},
		Settings: w,
		Capture:  c,
		TPS:      l.TPS,
	})
}

// run opens the window and runs the game until it ends.
func run(cfg shellConfig) error {
	gpu.Mark("ui.Play")
	if cfg.TPS <= 0 { // unset: the game would never tick after its first frame (a black, deaf window)
		cfg.TPS = 60
	}
	t := cfg.Theme
	Shooting = cfg.Capture.Shot != "" // -shot: the 3D view ignores the keyboard
	var fg *frontGame
	if a, ok := cfg.Front.(interface{ Arts() (bool, func(bool)) }); ok && cfg.Art == nil {
		if dos, set := a.Arts(); set != nil { // the front's own two arts (F12), DOS first
			want := dos
			cfg.Art = &Arts{Want: func() bool { return want }, Toggle: func() { want = !want }, Set: set}
		}
	}
	if cfg.Front != nil {
		fg = &frontGame{f: cfg.Front, cfg: cfg, tps: cfg.TPS}
		cfg.Game = fg
	}
	if cfg.Menu != nil {
		held := func(any) bool { return true }
		if fg != nil {
			held = func(k any) bool { return fg.latch.Down(k, true) }
		}
		if h, ok := cfg.Front.(fontHost); ok && cfg.Menu.FontNames == nil { // the front's text fonts
			h.KeepFonts(cfg.Menu.Prefs.Dir, cfg.Theme.Title)
			cfg.Menu.FontNames, cfg.Menu.FontAt, cfg.Menu.SetFont, cfg.Menu.FontGlyph = h.FontMenu()
		}
		cfg.Menu.setup(held, cfg.TPS)
		cfg.Settings.Fullscreen = cfg.Settings.Fullscreen || Settings.Fullscreen
	}
	if fg != nil && cfg.Art != nil {
		fg.art = cfg.Art.Want()
	}
	gpu.SetWindowSize(t.W*cfg.Settings.Scale, t.H*cfg.Settings.Scale) // the title pictures' shape (16:10): they fill the window whole, uncropped
	gpu.SetWindowAspect(t.W, t.H)                                     // and a resize keeps it so; the full screen has its own shape, the UI pages lay out to it and the pictures get black bars
	if w := cfg.Capture.Window; w[0] > 0 && w[1] > 0 {                // a window of that size and shape, for captures of other shapes
		gpu.SetWindowSize(w[0], w[1])
		gpu.SetWindowAspect(0, 0)
	}
	gpu.SetWindowTitle(t.Title)
	gpu.SetTPS(cfg.TPS)
	if len(cfg.Capture.Keys) > 0 { // a scripted run never takes focus, however it was launched (audio.Silent: never heard, set by cmd/gobox captureFlags before any sound starts)
		os.Setenv("GOBOX_BACKGROUND", "1")
	}
	gpu.SetFullscreen(cfg.Settings.Fullscreen && !cfg.Settings.Window && len(cfg.Capture.Keys) == 0) // a scripted run (-keys) stays a window
	gpu.SetConfig(cfg.Capture.GPU)
	if st, ok := cfg.Game.(starter); ok {
		st.Start(cfg.Capture.Slot)
	} else if st, ok := cfg.Front.(starter); ok {
		st.Start(cfg.Capture.Slot)
	}
	gpu.Mark("ui.Play: settings and fonts read, the window next")
	return gpu.RunGame(&shell{cfg: cfg})
}

func (s *shell) Update() error {
	s.cfg.Game.Update()
	return nil
}

// done says whether the script's keys have played and settled.
func (s *shell) done() bool {
	return s.fed == len(s.cfg.Capture.Keys) && s.ticks >= s.until && s.gap > 4*Ticks && (s.cfg.Game.Idle() || s.gap > MaxWait)
}

// step is a scripted run's clock, run by Draw once a frame, so a capture is
// the same run to run: the game steps once a frame, and the next key goes
// in once Ticks frames have passed and the game is Idle (MaxWait at most:
// a screen that never settles still gets its keys). \w is Ticks frames
// more. With Freeze time stands still after the keys.
func (s *shell) step() {
	c := s.cfg.Capture
	if c.Freeze && s.done() {
		return
	}
	s.ticks++
	s.gap++
	var k Key
	if s.fed < len(c.Keys) && s.ticks >= s.until {
		next := c.Keys[s.fed]
		timed := next&0xF000 == ScriptTimed
		switch {
		case timed: // at once, idle or not
			s.until, s.fed, s.gap = s.ticks+int(next&0xFFF), s.fed+1, 0
		case next == ScriptSnap: // at once: the frame as it is now
			s.snap, s.fed, s.gap = true, s.fed+1, 0
		case s.gap >= Ticks && (next == ScriptWait || s.cfg.Game.Idle() || s.gap >= MaxWait):
			k, s.fed, s.gap = next, s.fed+1, 0
			fmt.Fprintf(os.Stderr, "%s key %d/%d %#x tick %d\n", time.Now().Format("15:04:05.000"), s.fed, len(c.Keys), next, s.ticks) // each scripted key in the run's log, to tie the trace's acts to keys
		}
	}
	alt := k&0xF000 == ScriptAlt
	if alt {
		k &= 0xFF
	}
	gpu.Hold(gpu.KeyAlt, alt)
	s.cfg.Game.Step(k)
}

func (s *shell) Draw(dst *gpu.Image) {
	c := s.cfg.Capture
	if c.Scripted() {
		s.step()
	}
	s.cfg.Game.Draw(dst)
	if s.snap && c.Shot != "" { // \S: OUT-1.png, OUT-2.png ... mid-run
		s.snap, s.snaps = false, s.snaps+1
		saveShot(dst, fmt.Sprintf("%s-%d.png", strings.TrimSuffix(c.Shot, ".png"), s.snaps))
	}
	if c.Shot == "" {
		if c.Quit && s.done() { // the run ends after the keys (measuring)
			os.Exit(0)
		}
		return
	}
	if strings.Contains(c.Shot, "%d") { // every frame while the keys play, numbered (to catch a glitch of one frame)
		if s.fed > 0 && !s.done() {
			saveShot(dst, fmt.Sprintf(c.Shot, s.frames))
			s.frames++
		}
	} else if s.done() {
		saveShot(dst, c.Shot)
	}
	if s.done() {
		os.Exit(0)
	}
}

// Layout works in device pixels, so the picture fills the window exactly.
func (s *shell) Layout(w, h int) (int, int) {
	dsf := gpu.DeviceScaleFactor()
	return int(float64(w) * dsf), int(float64(h) * dsf)
}

func saveShot(dst *gpu.Image, name string) {
	b := dst.Bounds()
	img := image.NewRGBA(b)
	dst.ReadPixels(img.Pix)
	if err := asset.WritePNG(name, img); err != nil {
		log.Fatal(err)
	}
}
