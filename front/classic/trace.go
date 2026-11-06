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
	s := g.R.State()
	m := s.Mode
	turn, pace := "", ""
	if c := s.Combat; c != nil && c.Turn != nil {
		turn = fmt.Sprintf(" turn=unit%d", c.Turn.Unit)
	}
	if s.Pacing != nil {
		pace = fmt.Sprintf(" pace=%v", s.Pacing.Pace)
	}
	fmt.Fprintf(os.Stderr, "trace %s err=%v events=%d mode=%v%s%s\n", what, res.Err, len(res.Events), m, turn, pace)
	if m != g.traceMode {
		fmt.Fprintf(os.Stderr, "trace mode %v -> %v\n", g.traceMode, m)
		g.traceMode = m
	}
}
