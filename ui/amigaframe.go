package ui

import (
	"image"
	"image/draw"

	"gobox/game"
)

// Amiga frame screens: the Amiga releases draw their window frame from
// whole pieces (bars and columns) at fixed screen positions, not from 8x8
// symbols. gobox replays the program's blits onto 320x200 screens
// (game.AmigaFrame.Screens) and cuts the border cells out of them, so every
// cell is the Amiga's own pixels at that place on its screen.

const (
	amigaScreenW = 320 // the Amiga's low-resolution screen, in pixels
	amigaScreenH = 200
	frameCell    = 8 // a border cell's side, in pixels
)

// amigaFrameKey is one composed frame: the frame and its set.
type amigaFrameKey struct {
	frame *game.AmigaFrame
	set   int
}

// amigaScreens are the composed frame screens (amigaFrame); the art switch
// and a new theme drop them with the pictures they are made of.
var amigaScreens = map[amigaFrameKey][]*image.RGBA{}

// amigaFrame is a's border pieces (by place, the caps none) and edges for
// symbol set set, cut from its screens.
func amigaFrame(a *game.AmigaFrame, set int) (pieces [LeftTopCap]image.Image, edges map[string][]image.Image) {
	screens := composeFrame(a, set)
	cell := func(c game.FrameCell) image.Image {
		if c.Screen < 0 || c.Screen >= len(screens) {
			return nil
		}
		r := image.Rect(c.Col*frameCell, c.Row*frameCell, (c.Col+1)*frameCell, (c.Row+1)*frameCell)
		return screens[c.Screen].SubImage(r)
	}
	for i, c := range a.Pieces {
		pieces[i] = cell(c)
	}
	edges = map[string][]image.Image{}
	for k, cs := range a.Edges {
		for _, c := range cs {
			edges[k] = append(edges[k], cell(c))
		}
	}
	return pieces, edges
}

// composeFrame is a's screens for symbol set set: each blit drawn over the
// ones before it, as the program draws them.
func composeFrame(a *game.AmigaFrame, set int) []*image.RGBA {
	k := amigaFrameKey{a, set}
	if s, ok := amigaScreens[k]; ok {
		return s
	}
	var out []*image.RGBA
	for _, blits := range a.Screens {
		scr := image.NewRGBA(image.Rect(0, 0, amigaScreenW, amigaScreenH))
		for _, b := range blits {
			block := b.Block
			if len(a.Sets) > 0 && block == a.Sets[0] {
				block = setBlock(a.Sets, block, set)
			}
			mask := -1
			if b.Masked {
				mask = int(a.Mask)
			}
			m := daxMasked(a.File, block, b.Frame, mask)
			if m == nil {
				continue
			}
			r := m.Bounds().Sub(m.Bounds().Min).Add(image.Pt(b.X, b.Y))
			draw.Draw(scr, r, m, m.Bounds().Min, draw.Over)
		}
		out = append(out, scr)
	}
	amigaScreens[k] = out
	return out
}
