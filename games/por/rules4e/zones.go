package rules4e

import (
	"gobox/core"
	"rules1e"
)

// Zones (PHB p. 273): an area a power leaves on the map until the end of
// its caster's next turn. Cloud of Daggers hurts who enters it or starts
// its turn in it; Icy Terrain is difficult terrain, a square in it costs
// one more.

// zone is a power's area: the squares within r of at.
type zone struct {
	at    core.BoardPos
	r     int
	by    *rules1e.Character
	dmg   int  // to who enters or starts its turn there; 0: difficult terrain
	begun bool // the caster's next turn has begun
}

// in: c stands in z.
func (z zone) in(e *rules1e.RuleState, c *rules1e.Character) bool {
	d := e.At(c).Sub(z.at)
	return max(d.X, -d.X, d.Y, -d.Y) <= z.r
}

// zoneStart is TurnStart's part: c's zones' next turn begins; a cloud c
// starts its turn in hurts it.
func (r *Rules) zoneStart(e *rules1e.RuleState, c *rules1e.Character) {
	for i := range r.zones {
		if r.zones[i].by == c {
			r.zones[i].begun = true
		}
	}
	r.zoneHurt(e, c)
}

// zoneEnd is TurnEnd's part: c's zones end at the end of its next turn.
func (r *Rules) zoneEnd(c *rules1e.Character) {
	kept := r.zones[:0]
	for _, z := range r.zones {
		if z.by != c || !z.begun {
			kept = append(kept, z)
		}
	}
	r.zones = kept
}

// zoneHurt: the clouds c stands in hurt it.
func (r *Rules) zoneHurt(e *rules1e.RuleState, c *rules1e.Character) {
	for _, z := range r.zones {
		if z.dmg > 0 && z.in(e, c) && c.Okay != 0 {
			e.Say(c, "is cut by the daggers")
			e.Hurt(c, z.dmg)
		}
	}
}

// Entered: a cloud hurts who comes into it; ice costs a move one square more.
func (r *Rules) Entered(e *rules1e.RuleState, c *rules1e.Character) {
	r.zoneHurt(e, c)
	if e.Forced() || c.Combat == nil {
		return
	}
	for _, z := range r.zones {
		if z.dmg == 0 && z.in(e, c) {
			c.Combat.Moves -= min(c.Combat.Moves, 2) // a square more (half squares)
			return
		}
	}
}
