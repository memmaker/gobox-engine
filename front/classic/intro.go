package classic

import (
	"gobox/front"
	"gobox/game"
)

// Intro is v with the title in, waiting to be played. v where in is nil.
func Intro(v front.View, in *game.Intro) front.View {
	if in != nil {
		v.Intro = &front.Intro{Full: in.Full, Pictures: in.Pictures, Lightning: in.Lightning}
	}
	return v
}

// Sequence is v with the animation name names among seqs, its files in dir.
func Sequence(v front.View, name string, seqs []game.Sequence, dir string) front.View {
	for _, q := range seqs {
		if name != "" && q.Name == name {
			v.Sequence = &front.Sequence{Name: q.Name, Anim: q.Anim, Dir: dir}
		}
	}
	return v
}
