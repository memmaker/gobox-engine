package classic

import (
	"fmt"
	"os"

	"gobox/game"
)

// traceOn (GOBOX_TRACE=1) logs to stderr every decision sent to the game,
// every step, and each mode change.
var traceOn = os.Getenv("GOBOX_TRACE") != ""

// trace logs what (a decision, or "step") and the mode it left, when on.
func (g *Game) trace(what string, res game.Result) {
	if !traceOn {
		return
	}
	m := g.R.State().Mode
	fmt.Fprintf(os.Stderr, "trace %s err=%v events=%d mode=%v\n", what, res.Err, len(res.Events), m)
	if m != g.traceMode {
		fmt.Fprintf(os.Stderr, "trace mode %v -> %v\n", g.traceMode, m)
		g.traceMode = m
	}
}
