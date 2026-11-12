package pool

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Explicit interactions (extra, the Esc menu's Gameplay "Explicit
// Interactions", Engine.Explicit): a step onto a square whose event is a
// harmless one (a description, a shop, a talk, a question) stops the
// square's script where that event starts. The HUD says so
// (View.Interact) and RunEvent runs it. Dangerous events
// (ambushes, traps) and the ones not tied to a square (wandering
// monsters) run on stepping as ever. data/assets/scripts/interactions.txt
// names the events' labels, per area.

// interactions reads the table in dir: per area, the labels a step stops at.
func interactions(dir string) map[int][]string {
	b, err := os.ReadFile(filepath.Join(dir, "interactions.txt"))
	if err != nil {
		return nil
	}
	out := map[int][]string{}
	for _, l := range strings.Split(string(b), "\n") {
		l, _, _ = strings.Cut(l, "#")
		f := strings.Fields(l)
		if len(f) < 2 {
			continue
		}
		if a, err := strconv.Atoi(f[0]); err == nil {
			out[a] = append(out[a], f[1])
		}
	}
	return out
}

// stepEvents runs the square's events after a step or an area change:
// with Explicit on, it stops at a harmless event (ext.stopAt), which then
// waits for RunEvent.
func (e *Engine) stepEvents() {
	e.ext.interact = nil
	e.ext.stepping = e.Explicit.Load()
	e.push(cont{kind: kStepEvents})
	e.RunECL(e.es.events[1])
}

// stopsHere: a step's events have come to a harmless event, which waits
// for RunEvent.
func (e *Engine) stopsHere() bool {
	if !e.ext.stepping || !e.ext.stopAt[int(e.vm.PC)] {
		return false
	}
	e.ext.interact = &[2]int{e.posX, e.posY}
	e.vm.Calls = nil
	return true
}

// stepOwed: Arrive moved the party here and its step events wait (RunEvent).
func (e *Engine) stepOwed() bool {
	return e.ext.owed != nil && *e.ext.owed == [2]int{e.posX, e.posY}
}

// interactHere: the party stands on the square whose event waits.
func (e *Engine) interactHere() bool {
	return e.ext.interact != nil && *e.ext.interact == [2]int{e.posX, e.posY}
}

// interactNow runs the waiting event (RunEvent): the square's events,
// all of them; then the next command.
func (e *Engine) interactNow() {
	e.ext.interact = nil
	e.es.picShown, e.es.picAnimating = 0, 1
	e.push(cont{kind: kStepEnd}, cont{kind: kAreaChanged})
	e.RunECL(e.es.events[1])
}
