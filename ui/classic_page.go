package ui

import (
	"image"
	"image/color"

	"gobox/ega"
	"gobox/game"
)

func paneRuns(l []game.TextRun) []PaneRun {
	var r []PaneRun
	for _, t := range l {
		if t.Fg == (color.RGBA{}) { // unset: the body text's colour (gobox's, as every font colour)
			t.Fg = ega.RGBA(10)
		}
		r = append(r, PaneRun{X: t.X, Text: t.Text, Fg: t.Fg, Bg: t.Bg, Upcase: t.Upcase, Blink: t.Blink})
	}
	return r
}

func paneLines(ls [][]game.TextRun) [][]PaneRun {
	if ls == nil {
		return nil
	}
	r := make([][]PaneRun, len(ls))
	for i, l := range ls {
		r[i] = paneRuns(l)
	}
	return r
}

// story is a script portrait in the view window: the HEAD at cells 3,3, the
// BODY five rows below (the Gold Box's ShowPortrait); other pictures stack on.
func story(pics []game.Picture) []Picture {
	var r []Picture
	y := 24
	for i, pc := range pics {
		h := []int{40, 48}[min(i, 1)]
		f := pix(pc)
		f.File, f.Block, f.Frame = pc.File, pc.Block, pc.Frame
		f.At = image.Rect(24, y, 24+88, y+h)
		r = append(r, f)
		y += h
	}
	return r
}
