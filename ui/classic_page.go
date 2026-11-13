package ui

import (
	"image"
	"image/color"

	"gobox/asset"
	"gobox/game"
)

// TextPage is the text page the state shows: the pane's lines, the prompt
// row, the lines over the view window (Scene.Placed) and the pictures,
// bottom first.
type TextPage struct {
	Pane   [][]game.TextRun
	Prompt []game.TextRun
	View   [][]game.TextRun
	Layers []game.Layer
}

// paneRuns lays line l out from the line's start in ink: an unset ink is
// the body text's colour (gobox's, as every font colour) on no background.
func paneRuns(l []game.TextRun, ink [16]color.RGBA) []PaneRun {
	var r []PaneRun
	x := 0
	for _, t := range l {
		fg, bg := asset.EGARGBA(10), color.RGBA{}
		if n, ok := t.Fg.EGA(); ok {
			fg = ink[n]
		}
		if n, ok := t.Bg.EGA(); ok {
			bg = ink[n]
		}
		r = append(r, PaneRun{X: x, Text: t.Text, Fg: fg, Bg: bg, Blink: t.Blink})
		x += len(t.Text)
	}
	return r
}

func paneLines(ls [][]game.TextRun, ink [16]color.RGBA) [][]PaneRun {
	if ls == nil {
		return nil
	}
	r := make([][]PaneRun, len(ls))
	for i, l := range ls {
		r[i] = paneRuns(l, ink)
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
