package present

import (
	"gobox/front"
	"gobox/game"
)

// Waits is v with the locked door's ways through and the power a member
// must choose.
func Waits(v front.View, d *game.Door, p *game.PowerPick) front.View {
	v.Door, v.Powers = nil, nil
	if d != nil {
		v.Door = &front.DoorOpts{Bash: d.Bash, Pick: d.Pick, Knock: d.Knock}
	}
	if p != nil {
		v.Powers = &front.PowerPick{Name: p.Name, What: p.What}
		for _, m := range p.Options {
			v.Powers.Options = append(v.Powers.Options, front.Maneuver{ID: m.ID, Name: m.Name, Tip: m.Tip, Aim: front.Aim(m.Aim), Range: m.Range, Area: m.Area, Move: m.Move})
		}
	}
	return v
}
