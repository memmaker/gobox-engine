package app

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"gobox/gpu"
)

// Flags registers every game's capture and test switches on the command
// line (call it before flag.Parse):
//
//	-shot FILE.png  the window after -keys, then exit (%d: every frame)
//	-keys S         the script (ParseKeys with esc, the game's escapes)
//	-freeze, -quit  time stands still / the run ends after the keys
//	-slot NAME      the save slot to start from (a Starter's)
//	-pointer x,y;x,y...  the pointer positions (window pixels)
//	-png FILE       a picture a dev scene uses
//	-size WxH       a window of that size and shape (points)
//	-fps, -passes, -novsync, -offscreen, -glitch DIR, -skip a,b  gpu.Config
//	-cpuprofile FILE  a CPU profile of the first 20 seconds
//	-stacks N       every goroutine to stderr every N seconds
//
// The returned func, called after flag.Parse, reads them into a Capture
// and starts the profile and the stack dumps.
func Flags(esc map[byte]Key) func() Capture {
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
	return func() Capture {
		c := Capture{Shot: *shot, Keys: ParseKeys(*keys, esc), Freeze: *freeze, Quit: *quit, Slot: *slot, PNG: *png}
		for _, pt := range strings.Split(*pointer, ";") {
			var p [2]int
			if _, err := fmt.Sscanf(pt, "%d,%d", &p[0], &p[1]); err == nil {
				c.Pointer = append(c.Pointer, p)
			}
		}
		if w, h, err := ParseSize(*size); err == nil {
			c.Window = [2]int{w, h}
		}
		c.GPU = gpu.Config{FPS: *fps, Passes: *passes, NoVSync: *novsync, Offscreen: c.Shot != "" || *offscreen,
			GlitchDir: *glitch, Skip: strings.Split(*skip, ",")}
		if *cpuprofile != "" {
			if f, err := os.Create(*cpuprofile); err == nil {
				pprof.StartCPUProfile(f)
				time.AfterFunc(20*time.Second, func() { pprof.StopCPUProfile(); f.Close() })
			}
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
