package ui

import (
	"fmt"
	"os"
	"time"

	"gobox/game"
)

// traceOn (GOBOX_TRACE=1) logs to stderr every decision sent to the game,
// every step, and each mode change.

// trace logs what (a decision, or "step") and the mode it left, when on.
func (g *Game) trace(what string, res game.Result) {
	if !traceOn {
		return
	}
	m := g.mode()
	turn, pace := "", ""
	if c := g.combat(); c != nil && c.Turn != nil {
		turn = fmt.Sprintf(" turn=unit%d", c.Turn.Unit)
		for _, o := range c.Turn.Options {
			turn += fmt.Sprintf(" %s/%c", o.Label, o.Key)
		}
	}
	pace = fmt.Sprintf(" pace=%v", g.pace())
	fmt.Fprintf(os.Stderr, "trace %s %s err=%v events=%d mode=%v%s%s%s\n", traceStamp(), what, res.Err, len(res.Events), m, turn, pace, traceParty(g.party()))
	for _, e := range res.Events {
		if e.Text != "" {
			fmt.Fprintf(os.Stderr, "trace %s   event %v %q\n", traceStamp(), e.Kind, e.Text)
		}
	}
	if m != g.traceMode {
		fmt.Fprintf(os.Stderr, "trace %s mode %v -> %v\n", traceStamp(), g.traceMode, m)
		g.traceMode = m
	}
}

// traceParty is each member's name, HP and status, as a trace line's tail.
func traceParty(p []game.Member) (out string) {
	for _, m := range p {
		out += fmt.Sprintf(" [%s %d/%d %s]", m.Name, m.HP, m.MaxHP, m.Status)
	}
	if out != "" {
		out = " party=" + out[1:]
	}
	return out
}

// traceStamp is a trace line's wall-clock time to the millisecond, so a
// replay's and a hold's timings can be read off the trace.
func traceStamp() string { return time.Now().Format("15:04:05.000") }
