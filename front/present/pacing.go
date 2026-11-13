package present

import (
	"gobox/front"
	"gobox/game"
)

// Pacing is v with what goes on with time. v where p is nil.
func Pacing(v front.View, p *game.Pacing) front.View {
	if p != nil {
		v.Pace, v.Resting, v.Seq = front.Pace(p.Pace), p.Resting, p.Seq
	}
	return v
}
