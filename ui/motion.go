package ui

import (
	"image"
	"math"
	"time"
)

// How the figures move between the boards the game hands over: never a
// jump. A unit walks to its new square cell by cell at a steady pace (the
// path the board offered for it, else the straight line of squares), a new
// one rises onto the board, one the board dropped (fallen, fled, removed)
// finishes its walk and sinks away.
const (
	DefaultStepTime = 0.18                   // seconds a step takes when the host gives no StepTime
	appearTime      = 0.35                   // seconds a unit takes to rise onto the board
	departTime      = 0.45                   // seconds one the board dropped takes to sink away
	catchUpSteps    = 3                      // a walk further behind than this many steps speeds up...
	maxStepsAFrame  = 1.0                    // ...but no figure ever moves more than a step in one frame
	unseenGap       = 250 * time.Millisecond // a board not drawn this long was off the screen: shown again, it starts afresh
)

// walker is how a unit is shown: where its figure stands now and the
// squares (centres) it still walks to.
type walker struct {
	at, end [2]float64 // shown centre; where the walk ends (the board's last)
	path    [][2]float64
	grow    float64 // 0..1: risen onto the board
	gone    float64 // seconds since the board dropped it (departing)
	unit    Unit    // as the last board had it (a departure draws it)
}

// shownUnit is a figure to draw this frame: the unit, where, how big (0..1).
type shownUnit struct {
	u     Unit
	x, z  float32
	scale float32
}

// moveUnits brings every figure a frame of dt seconds closer to b's
// board and lists what to draw, the departing ones after b's units.
func (p *CombatPresenter) moveUnits(b *Board, dt float64) []shownUnit {
	if p.units == nil {
		p.units = map[int]*walker{}
	}
	step := p.StepTime
	if step <= 0 {
		step = DefaultStepTime
	}
	first := !p.started
	p.started = true
	var out []shownUnit
	here := map[int]bool{}
	for _, u := range b.Units {
		if u.Icon == nil {
			continue
		}
		here[u.ID] = true
		c := unitCentre(u)
		w, ok := p.units[u.ID]
		if !ok || w.gone > 0 { // new, or back before it sank away
			if !ok {
				w = &walker{at: c, end: c}
				if first {
					w.grow = 1
				}
				p.units[u.ID] = w
			}
			w.gone = 0
		}
		if c != w.end {
			w.path = append(w.path, p.cells(w, u, c)...)
			w.end = c
		}
		w.unit = u
		w.advance(dt, step)
		w.grow = min(1, w.grow+dt/appearTime)
		out = append(out, shownUnit{u, float32(w.at[0]), float32(w.at[1]), float32(w.grow)})
	}
	for id, w := range p.units {
		if here[id] {
			continue
		}
		w.advance(dt, step)
		if len(w.path) == 0 { // the walk off the board done: it sinks away
			w.gone += dt
		}
		if w.gone >= departTime {
			delete(p.units, id)
			continue
		}
		out = append(out, shownUnit{w.unit, float32(w.at[0]), float32(w.at[1]), float32(w.grow * (1 - w.gone/departTime))})
	}
	p.paths = b.Paths
	return out
}

// cells are the centres u walks through from where w's walk ends to c:
// the last board's path to u's square where it offered one, else the
// straight line of squares.
func (p *CombatPresenter) cells(w *walker, u Unit, c [2]float64) [][2]float64 {
	off := [2]float64{c[0] - float64(u.X), c[1] - float64(u.Y)} // the centre's place on the unit's anchor square
	if steps := p.paths[image.Pt(u.X, u.Y)]; len(steps) > 1 && centre(steps[0].Point, off) == w.end {
		out := make([][2]float64, 0, len(steps)-1)
		for _, s := range steps[1:] {
			out = append(out, centre(s.Point, off))
		}
		return out
	}
	d := [2]float64{c[0] - w.end[0], c[1] - w.end[1]}
	n := int(math.Ceil(max(math.Abs(d[0]), math.Abs(d[1]))))
	if n == 0 {
		return [][2]float64{c}
	}
	out := make([][2]float64, 0, n)
	for i := 1; i <= n; i++ {
		t := float64(i) / float64(n)
		out = append(out, [2]float64{w.end[0] + math.Round(d[0]*t), w.end[1] + math.Round(d[1]*t)})
	}
	out[len(out)-1] = c
	return out
}

func centre(at image.Point, off [2]float64) [2]float64 {
	return [2]float64{float64(at.X) + off[0], float64(at.Y) + off[1]}
}

// advance walks w dt seconds on its path, a step every step seconds (faster
// when far behind, never more than maxStepsAFrame).
func (w *walker) advance(dt, step float64) {
	budget := dt / step
	if n := len(w.path); n > catchUpSteps {
		budget *= float64(n) / catchUpSteps
	}
	budget = min(budget, maxStepsAFrame)
	for budget > 0 && len(w.path) > 0 {
		to := w.path[0]
		d := [2]float64{to[0] - w.at[0], to[1] - w.at[1]}
		n := max(math.Abs(d[0]), math.Abs(d[1])) // a diagonal is one step, as the rules count it
		if n <= budget {
			w.at, w.path, budget = to, w.path[1:], budget-n
			continue
		}
		w.at = [2]float64{w.at[0] + d[0]*budget/n, w.at[1] + d[1]*budget/n}
		budget = 0
	}
}
