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

// Sequence is v with the animation s.Sequence names among seqs, its files in dir.
func Sequence(v front.View, s game.State, seqs []game.Sequence, dir string) front.View {
	for _, q := range seqs {
		if s.Sequence != "" && q.Name == s.Sequence {
			v.Sequence = &front.Sequence{Name: q.Name, Anim: q.Anim, Dir: dir}
		}
	}
	return v
}
