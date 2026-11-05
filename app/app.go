// Package app is the app shell: the window, the frame loop,
// fullscreen and the window's shape, and a scripted capture's clock and
// screenshots. The game supplies its ticks and frames (Game); Run runs it.
package app

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"strings"

	"gobox/blobber"
	"gobox/gpu"
)

// Game is what a game gives the shell.
type Game interface {
	// Update is one tick of the game (TPS a second). During a scripted run
	// (Capture.Scripted) the user's input must not reach the game.
	Update()
	// Step is one frame of a scripted run, before Draw: k is the script's
	// key due this frame, 0 for none (ScriptRight, ScriptLeft, ScriptNext,
	// ScriptWait or a game key code from Capture.Keys).
	Step(k Key)
	// Draw draws a frame into dst (device pixels, the whole window).
	Draw(dst *gpu.Image)
}

// Key is one key of a script (Capture.Keys): a game key code or one
// of the shell's own.
type Key = uint16

// The script's own keys: \M and \C a right and a left click at the
// pointer, \P the pointer on to the next position, \w a quarter second of
// nothing.
const Right, Left, Next, Wait, Hold Key = 0xFF01, 0xFF02, 0xFF03, 0xFF04, 0xFF05

// Ticks is the frames between a script's keys: a quarter second at 60 Hz.
const Ticks = 15

// Theme is the game's look of the window.
type Theme struct {
	Title string
	W, H  int // the window's size at scale 1, and its shape (16:10: the title pictures fill it)
}

// Settings is how the window opens.
type Settings struct {
	Scale      int  // the window is Theme.W x Theme.H times Scale
	Fullscreen bool // the screen at its own resolution instead of a window
	Window     bool // a window whatever Fullscreen or the Esc menu says
}

// Capture is a scripted run: keys in, frames out, the user's input ignored.
type Capture struct {
	Shot    string     // FILE.png: the window after the keys, then exit; with %d every frame while the keys play
	Keys    []Key      // the script (ParseKeys)
	Freeze  bool       // time stands still after the keys
	Quit    bool       // the run ends after the keys
	Window  [2]int     // W, H (points): a window of that size and any shape; 0: Theme's
	GPU     gpu.Config // gpu's measuring and capture switches
	Slot    string     // a save slot to start from instead of the game's own start ("": none); passed to a Starter
	Pointer [][2]int   // with Shot: the pointer at these window pixels, one after another (Next)
	PNG     string     // a picture a dev scene uses (e.g. a portrait); the game's to read
}

// Scripted says whether the run is scripted: the user's input is ignored.
func (c Capture) Scripted() bool { return c.Shot != "" || c.Freeze || c.Quit }

// Starter is a Game told before the first tick which save slot to start
// from (Capture.Slot, "" none); what that means is the game's.
type Starter interface{ Start(slot string) }

// Config is the app: the game and how it runs.
type Config struct {
	Game     Game // or Front: the shell reads the input for the game's screens
	Front    Front
	Art      *Art  // with Front: two arts, cross-faded (nil: one)
	FontPrev Key   // with Front: the code Ctrl-Shift-F sends (0: none)
	Menu     *Menu // the Esc menu's settings (nil: blobber's own)
	Theme    Theme
	Settings Settings
	Capture  Capture
	TPS      int // Update calls a second; 0: 60
}

// ParseKeys reads a script: \M \C \P \w \B are the shell's (Right,
// Left, Next, Wait, Hold: the left button down until the next \B, so
// \B\P\B drags), and gobox's own keys are \r Enter, \e Esc, \t Tab,
// \U \D \L \R the arrows, \G \H PgDn PgUp, \F F10; esc may add a
// game's own, and any other byte is its own code.
func ParseKeys(s string, esc map[byte]Key) []Key {
	all := map[byte]Key{'M': Right, 'C': Left, 'P': Next, 'w': Wait, 'B': Hold,
		'r': 13, 'e': 27, 't': 9, 'U': 0x4800, 'D': 0x5000, 'L': 0x4B00, 'R': 0x4D00, 'G': 0x5100, 'H': 0x4900, 'F': 0x4400}
	for b, k := range esc {
		all[b] = k
	}
	var out []Key
	for i := 0; i < len(s); i++ {
		if c, ok := all[s[min(i+1, len(s)-1)]]; s[i] == '\\' && i+1 < len(s) && ok {
			out = append(out, c)
			i++
		} else {
			out = append(out, Key(s[i]))
		}
	}
	return out
}

// ParseSize reads "WxH".
func ParseSize(s string) (w, h int, err error) {
	_, err = fmt.Sscanf(s, "%dx%d", &w, &h)
	return
}

// shell runs a game on gpu.
type shell struct {
	cfg    Config
	ticks  int // scripted frames
	frames int // frames saved, Shot with %d
}

// Run opens the window and runs the game until it ends.
func Run(cfg Config) error {
	gpu.Mark("app.Run")
	if cfg.TPS <= 0 { // unset: the game would never tick after its first frame (a black, deaf window)
		cfg.TPS = 60
	}
	t := cfg.Theme
	blobber.Shooting = cfg.Capture.Shot != "" // -shot: the 3D view ignores the keyboard
	var fg *frontGame
	if a, ok := cfg.Front.(interface{ Arts() (bool, func(bool)) }); ok && cfg.Art == nil {
		if dos, set := a.Arts(); set != nil { // the front's own two arts (F12), DOS first
			want := dos
			cfg.Art = &Art{Want: func() bool { return want }, Toggle: func() { want = !want }, Set: set}
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
		cfg.Settings.Fullscreen = cfg.Settings.Fullscreen || blobber.Settings.Fullscreen
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
	if len(cfg.Capture.Keys) > 0 { // a scripted run is never heard and never takes focus, however it was launched
		os.Setenv("GOBOX_NOSOUND", "1")
		os.Setenv("GOBOX_BACKGROUND", "1")
	}
	gpu.SetFullscreen(cfg.Settings.Fullscreen && !cfg.Settings.Window && len(cfg.Capture.Keys) == 0) // a scripted run (-keys) stays a window
	gpu.SetConfig(cfg.Capture.GPU)
	if st, ok := cfg.Game.(Starter); ok {
		st.Start(cfg.Capture.Slot)
	} else if st, ok := cfg.Front.(Starter); ok {
		st.Start(cfg.Capture.Slot)
	}
	gpu.Mark("app.Run: settings and fonts read, the window next")
	return gpu.RunGame(&shell{cfg: cfg})
}

func (s *shell) Update() error {
	s.cfg.Game.Update()
	return nil
}

// done says whether the script's keys have played and settled.
func (s *shell) done() bool { return s.ticks > Ticks*(len(s.cfg.Capture.Keys)+4) }

// step is a scripted run's clock, run by Draw once a frame, so a capture is
// the same frame for frame run to run (ticks and frames interleave
// differently each run): the game steps once a frame, a key goes in every
// Ticks frames, and with Freeze time stands still after the keys.
func (s *shell) step() {
	c := s.cfg.Capture
	if c.Freeze && s.done() {
		return
	}
	s.ticks++
	var k Key
	if n := s.ticks / Ticks; s.ticks%Ticks == 0 && n <= len(c.Keys) {
		k = c.Keys[n-1]
	}
	s.cfg.Game.Step(k)
}

func (s *shell) Draw(dst *gpu.Image) {
	c := s.cfg.Capture
	if c.Scripted() {
		s.step()
	}
	s.cfg.Game.Draw(dst)
	if c.Shot == "" {
		if c.Quit && s.done() { // the run ends after the keys (measuring)
			os.Exit(0)
		}
		return
	}
	if strings.Contains(c.Shot, "%d") { // every frame while the keys play, numbered (to catch a glitch of one frame)
		if s.ticks > Ticks && s.ticks <= Ticks*(len(c.Keys)+1) {
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
	f, err := os.Create(name)
	if err == nil {
		err = (&png.Encoder{CompressionLevel: png.BestSpeed}).Encode(f, img)
		f.Close()
	}
	if err != nil {
		log.Fatal(err)
	}
}
