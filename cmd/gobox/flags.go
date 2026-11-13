package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"gobox/audio"
	"gobox/gpu"
	"gobox/ui"
)

// captureFlags registers every game's capture and test switches on the command
// line (call it before flag.Parse):
//
//	-shot FILE.png  the window after -keys, then exit (%d: every frame)
//	-keys S         the script (parseKeys with esc, the game's escapes)
//	-freeze, -quit  time stands still / the run ends after the keys
//	-slot NAME      the save slot to start from (a ui starter's)
//	-pointer x,y;x,y...  the pointer positions (window pixels)
//	-png FILE       a picture a dev scene uses
//	-size WxH       a window of that size and shape (points)
//	-fps, -passes, -novsync, -offscreen, -glitch DIR, -skip a,b  gpu.Config
//	-cpuprofile FILE  a CPU profile of the first 20 seconds
//	-stacks N       every goroutine to stderr every N seconds
//	-fight N, -freetrain, -ecltrace, -saves DIR, -at B,X,Y  ui.DevFlags
//
// The returned func, called after flag.Parse, reads them into a ui.Capture
// and starts the profile and the stack dumps.
func captureFlags(esc map[byte]ui.Key) func() ui.Capture {
	shot := flag.String("shot", "", "capture: write the window to FILE.png after -keys and exit (%d: every frame)")
	keys := flag.String("keys", "", `capture: the keys to play, each once the game is idle, a quarter second apart at least (\r Enter, \e Escape, \M \C right/left click, \P next -pointer, \w wait ...)`)
	freeze := flag.Bool("freeze", false, "capture: time stands still after the keys")
	quit := flag.Bool("quit", false, "capture: end the run after the keys")
	slot := flag.String("slot", "", "start from this save slot")
	pointer := flag.String("pointer", "", "capture: the pointer at x,y;x,y... (window pixels), \\P steps on")
	png := flag.String("png", "", "dev: a picture a dev scene uses (a portrait)")
	size := flag.String("size", "", "capture: a window of WxH points, any shape")
	fps := flag.Bool("fps", false, "measure: a line a second with the frame rate")
	passes := flag.Bool("passes", false, "measure: a frame's render passes with each -fps line")
	novsync := flag.Bool("novsync", false, "measure: frames as fast as they come")
	offscreen := flag.Bool("offscreen", false, "draw to an image of its own (as -shot does)")
	glitch := flag.String("glitch", "", "measure: glitch check frames into DIR")
	skip := flag.String("skip", "", "measure: parts of the frame left out, a,b,...")
	cpuprofile := flag.String("cpuprofile", "", "measure: a CPU profile of the first 20 seconds to FILE")
	stacks := flag.Int("stacks", 0, "debug: every goroutine to stderr every N seconds")
	fight := flag.Int("fight", 0, "debug: a fight with three of monster N (the area's MON file) at the first adventuring prompt")
	freeTrain := flag.Bool("freetrain", false, "debug: the training hall trains everyone free, experience or not")
	eclTrace := flag.Bool("ecltrace", false, "debug: each script command and its operands to stderr")
	saves := flag.String("saves", "", "the save directory: its saves/NAME are the slots (default: the game's)")
	at := flag.String("at", "", "debug: enter script block BLOCK at X,Y at the first adventuring prompt (BLOCK,X,Y)")
	return func() ui.Capture {
		c := ui.Capture{Shot: *shot, Keys: parseKeys(*keys, esc), Freeze: *freeze, Quit: *quit, Slot: *slot, PNG: *png,
			Dev: ui.DevFlags{Fight: *fight, FreeTrain: *freeTrain, ECLTrace: *eclTrace, Saves: *saves}}
		if c.Keys != nil || *shot != "" || *quit || *fps || *cpuprofile != "" || *glitch != "" {
			audio.Silent = true // a scripted or measured run is never heard
		}
		if *at != "" {
			var a ui.At
			if _, err := fmt.Sscanf(*at, "%d,%d,%d", &a.Block, &a.X, &a.Y); err != nil {
				log.Fatal("-at: ", err)
			}
			c.Dev.At = &a
		}
		for _, pt := range strings.Split(*pointer, ";") {
			var p [2]int
			if _, err := fmt.Sscanf(pt, "%d,%d", &p[0], &p[1]); err == nil {
				c.Pointer = append(c.Pointer, p)
			}
		}
		if w, h, err := parseSize(*size); err == nil {
			c.Window = [2]int{w, h}
		}
		c.GPU = gpu.Config{FPS: *fps, Passes: *passes, NoVSync: *novsync, Offscreen: c.Shot != "" || *offscreen,
			GlitchDir: *glitch, Skip: strings.Split(*skip, ",")}
		if *cpuprofile != "" {
			f, err := os.Create(*cpuprofile)
			if err == nil {
				err = pprof.StartCPUProfile(f)
			}
			if err != nil {
				log.Fatal("cpuprofile: ", err)
			}
			time.AfterFunc(20*time.Second, func() {
				pprof.StopCPUProfile()
				if err := f.Close(); err != nil {
					log.Print("cpuprofile: ", err)
				}
			})
		}
		if *stacks > 0 {
			go func() {
				for range time.Tick(time.Duration(*stacks) * time.Second) {
					buf := make([]byte, 1<<20)
					fmt.Fprintf(os.Stderr, "--- stacks %s\n", time.Now().Format(time.TimeOnly))
					os.Stderr.Write(buf[:runtime.Stack(buf, true)])
				}
			}()
		}
		return c
	}
}

// parseKeys reads a script: \M \C \P \w \B are the shell's (Right,
// Left, Next, Wait, Hold: the left button down until the next \B, so
// \B\P\B drags), and gobox's own keys are \r Enter, \e Esc, \t Tab,
// \U \D \L \R the arrows, \G \H PgDn PgUp, \F F10, \A<c> c with Alt
// held (\Aq\Aa: alt+q, alt+a); esc may add a
// game's own, and any other byte is its own code.
func parseKeys(s string, esc map[byte]ui.Key) []ui.Key {
	all := map[byte]ui.Key{'M': ui.ScriptRight, 'C': ui.ScriptLeft, 'P': ui.ScriptNext, 'w': ui.ScriptWait, 'B': ui.ScriptHold, 'S': ui.ScriptSnap,
		'r': 13, 'e': 27, 't': 9, 'U': 0x4800, 'D': 0x5000, 'L': 0x4B00, 'R': 0x4D00, 'G': 0x5100, 'H': 0x4900, 'F': 0x4400}
	for b, k := range esc {
		all[b] = k
	}
	var out []ui.Key
	for i := 0; i < len(s); i++ {
		if j := i + 2; s[i] == '\\' && i+1 < len(s) && s[i+1] == 't' && j < len(s) && s[j] >= '0' && s[j] <= '9' { // \t<ms>: a timed wait (\t alone is Tab)
			ms := 0
			for ; j < len(s) && s[j] >= '0' && s[j] <= '9'; j++ {
				ms = ms*10 + int(s[j]-'0')
			}
			out = append(out, ui.ScriptTimed|ui.Key(min(0xFFF, max(1, ms*60/1000))))
			i = j - 1
			continue
		}
		if s[i] == '\\' && i+2 < len(s) && s[i+1] == 'A' { // \A<c>: c with Alt held
			out = append(out, ui.ScriptAlt|ui.Key(s[i+2]))
			i += 2
			continue
		}
		if c, ok := all[s[min(i+1, len(s)-1)]]; s[i] == '\\' && i+1 < len(s) && ok {
			out = append(out, c)
			i++
		} else {
			out = append(out, ui.Key(s[i]))
		}
	}
	return out
}

// parseSize reads "WxH".
func parseSize(s string) (w, h int, err error) {
	_, err = fmt.Sscanf(s, "%dx%d", &w, &h)
	return
}
