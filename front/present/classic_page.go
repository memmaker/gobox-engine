package present

import (
	"image"
	"image/color"

	"gobox/ega"
	"gobox/front"
	"gobox/game"
)

func paneRuns(l []game.TextRun) []front.PaneRun {
	var r []front.PaneRun
	for _, t := range l {
		if t.Fg == (color.RGBA{}) { // unset: the body text's colour (gobox's, as every font colour)
			t.Fg = ega.RGBA(10)
		}
		r = append(r, front.PaneRun{X: t.X, Text: t.Text, Fg: t.Fg, Bg: t.Bg, Upcase: t.Upcase, Blink: t.Blink})
	}
	return r
}

func paneLines(ls [][]game.TextRun) [][]front.PaneRun {
	if ls == nil {
		return nil
	}
	r := make([][]front.PaneRun, len(ls))
	for i, l := range ls {
		r[i] = paneRuns(l)
	}
	return r
}

// Page is v with the text page p: the pane, the prompt row, the
// text over the view window and the pictures. v where p is nil.
func Page(v front.View, p *game.Page, portrait []game.Picture) front.View {
	if p == nil {
		return v
	}
	if p.Pane != nil {
		v.Pane = &front.Pane{Lines: paneLines(p.Pane), Rows: p.PaneRows}
	} else {
		v.Pane = nil
	}
	v.Prompt, v.ViewText = paneRuns(p.Prompt), paneLines(p.View)
	v.Pictures = make([]front.Picture, 0, len(p.Layers))
	for _, l := range p.Layers {
		f := pix(l.Picture)
		f.File, f.Block, f.Frame, f.At = l.File, l.Block, l.Frame, l.At
		f.Frames, f.FrameMs = l.Frames, l.FrameMs
		v.Pictures = append(v.Pictures, f)
	}
	if len(p.Layers) == 0 {
		v.Pictures = append(v.Pictures, story(portrait)...)
	}
	return v
}

// story is a script portrait in the view window: the HEAD at cells 3,3, the
// BODY five rows below (the Gold Box's ShowPortrait); other pictures stack on.
func story(pics []game.Picture) []front.Picture {
	var r []front.Picture
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
