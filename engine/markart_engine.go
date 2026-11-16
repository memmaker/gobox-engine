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

// coverPicture is what lies over square q's ground: its marker's picture,
// else the tile covering it for a while (a cloud, a body: Cover), by name.
func (e *Engine) coverPicture(q rules.Square) game.Picture {
	if q.Cover > 0 {
		return game.Picture{Name: combatTileArt(e.artNames().TileOf(q.Cover - 1))}
	}
	return e.markPicture(q.Mark)
}

// overlandMark is the mark of the party's square on an overland drawn as one whole picture (mapcurs.pas DrawCursor: the cursor
// picture over the 8x8 cell of the party's square, which the picture's own cells count); nil off the overland or where the
// game names no cursor.
func (e *Engine) overlandMark() *game.PictureMark {
	c := e.artNames().OverlandCursor
	if c.File == "" || e.es.bigPic == nil || e.world.In3D != 0 || e.world.Flying != 0 {
		return nil
	}
	im := e.LoadPicMasked(c.File, c.Block, byte(c.Mask))
	if im == nil {
		return nil
	}
	sq := e.overlandSquare()
	return &game.PictureMark{Picture: framePicture(im, 0, im.Mask), At: rules.MapPos{X: sq.X, Y: sq.Y}}
}
