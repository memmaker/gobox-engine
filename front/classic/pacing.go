package classic

import (
	"gobox/front"
	"gobox/game"
)

// Pacing is v with what goes on with time s.Pacing holds. v where s has none.
func Pacing(v front.View, s game.State) front.View {
	if p := s.Pacing; p != nil {
		v.Pace, v.Resting, v.Seq = front.Pace(p.Pace), p.Resting, p.Seq
	}
	return v
}
