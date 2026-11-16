package engine

import (
	"gobox/game"
	"gobox/rules"
)

// markPicture is the picture of marker m over a square (the game's Marks: a combat sprite's frame); none for no marker.
func (e *Engine) markPicture(m rules.SquareMark) game.Picture {
	if m == 0 {
		return game.Picture{}
	}
	for _, a := range e.artNames().Marks {
		if a.Mark == m {
			if im := e.icons[a.Slot][a.Frame]; im != nil {
				return gamePicture(e.picPix(im, 0, false, nil))
			}
		}
	}
	return game.Picture{}
}
