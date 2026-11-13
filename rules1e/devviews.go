package rules1e

import (
	"gobox/core"
	"gobox/core/goldbox"
)

// DevPlace sets the fight up for a live check before its first turn (a
// dev scene's setup, as the placement): melee puts a foe on a free square
// next to each member (moving the member first when it is boxed in), else
// each member goes to the nearest free square beside the map's edge.
func (e *RuleState) DevPlace(melee bool) {
	free := func(p core.BoardPos) bool {
		o, g := e.MapAt(p.X, p.Y)
		return o == 0 && g > 0 && e.GroundInfo(g)[0] < 0xFF
	}
	near := func(p core.BoardPos, ok func(core.BoardPos) bool) (core.BoardPos, bool) {
		best, bd := p, 1<<30
		for y := 0; y < 25; y++ {
			for x := 0; x < 50; x++ {
				q := core.BoardPos{X: x, Y: y}
				if d := max(goldbox.Abs(x-p.X), goldbox.Abs(y-p.Y)); d < bd && ok(q) {
					best, bd = q, d
				}
			}
		}
		return best, bd < 1<<30
	}
	beside := func(p core.BoardPos, ok func(core.BoardPos) bool) (core.BoardPos, bool) {
		for dir := 0; dir < 8; dir++ {
			if q := p.Add(core.BoardPos{X: DirDX[dir], Y: DirDY[dir]}); ok(q) {
				return q, true
			}
		}
		return p, false
	}
	put := func(i int, p core.BoardPos) {
		e.Cpos[i].X, e.Cpos[i].Y = byte(p.X), byte(p.Y)
		e.CombatRedrawMap()
	}
	var foes []int
	for i := 1; i <= e.Ncomb; i++ {
		if c := e.CombatantSlots[i-1]; c != nil && c.Side != e.Party[0].Side && e.Cpos[i].Size == 1 {
			foes = append(foes, i)
		}
	}
	offMap := func(p core.BoardPos) bool { _, g := e.MapAt(p.X, p.Y); return g == 0 }
	for _, c := range e.Party {
		m := e.CombatIndex(c)
		if m == 0 {
			continue
		}
		at := core.BoardPos{X: int(e.Cpos[m].X), Y: int(e.Cpos[m].Y)}
		if !melee {
			if _, ok := beside(at, offMap); !ok {
				if q, ok := near(at, func(q core.BoardPos) bool { _, ok := beside(q, offMap); return ok && free(q) }); ok {
					put(m, q)
				}
			}
			continue
		}
		if len(foes) == 0 {
			return
		}
		if _, ok := beside(at, free); !ok {
			q, ok := near(at, func(q core.BoardPos) bool { _, ok := beside(q, free); return ok && free(q) })
			if !ok {
				continue
			}
			put(m, q)
			at = q
		}
		f, _ := beside(at, free)
		put(foes[0], f)
		foes = foes[1:]
	}
}
