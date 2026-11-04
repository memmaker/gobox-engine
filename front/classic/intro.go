package classic

import (
	"gobox/front"
	"gobox/game"
)

// Intro is v with the title s.Intro holds, waiting to be played. v where s
// has none.
func Intro(v front.View, s game.State) front.View {
	if in := s.Intro; in != nil {
		v.Intro = &front.Intro{Full: in.Full, Pictures: in.Pictures, Lightning: in.Lightning}
	}
	return v
}
