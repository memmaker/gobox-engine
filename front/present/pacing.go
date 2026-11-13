package present

import (
	"gobox/front"
	"gobox/game"
)

// Pacing is v with the change count. v where p is nil.
func Pacing(v front.View, p *game.Pacing) front.View {
	if p != nil {
		v.Seq = p.Seq
	}
	return v
}
