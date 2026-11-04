package gpu

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Timing is GOBOX_TIMING=1: a line on stderr for each start-up phase
// (Mark) and for each frame over 20 ms with what it did (Note), in ms
// since the program started. Off, Mark and Note cost a bool test.
var Timing = os.Getenv("GOBOX_TIMING") != ""

var (
	timingStart = time.Now() // package init: as good as the program's start
	frameNotes  []string
)

// Mark logs phase with the time since start (GOBOX_TIMING only).
func Mark(phase string, args ...any) {
	if Timing {
		fmt.Fprintf(os.Stderr, "timing %8.1f ms  %s\n", float64(time.Since(timingStart).Microseconds())/1000, fmt.Sprintf(phase, args...))
	}
}

// Note records what this frame did (a scene built, a texture uploaded),
// logged with the frame when it was slow (GOBOX_TIMING only).
func Note(what string, args ...any) {
	if Timing && len(frameNotes) < 24 {
		frameNotes = append(frameNotes, fmt.Sprintf(what, args...))
	}
}

// slowFrame logs a frame over 20 ms with its notes, and starts the next frame's.
func slowFrame(n int, update, draw, present time.Duration) {
	if !Timing {
		return
	}
	if t := update + draw + present; t > 20*time.Millisecond {
		Mark("slow frame %d: %.1f ms (update %.1f, draw %.1f, present %.1f) %s", n, ms(t), ms(update), ms(draw), ms(present), strings.Join(frameNotes, "; "))
	}
	frameNotes = frameNotes[:0]
}

func ms(d time.Duration) float64 { return float64(d.Microseconds()) / 1000 }
