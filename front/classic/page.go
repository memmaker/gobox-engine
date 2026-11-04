package classic

import (
	"gobox/front"
	"gobox/game"
)

func paneRuns(l []game.TextRun) []front.PaneRun {
	var r []front.PaneRun
	for _, t := range l {
		r = append(r, front.PaneRun{X: t.X, Text: t.Text, Fg: t.Fg, Bg: t.Bg, Upcase: t.Upcase})
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

// Page is v with the text page s.Page holds: the pane, the prompt row, the
// text over the view window and the pictures. v where s has no page.
func Page(v front.View, s game.State) front.View {
	p := s.Page
	if p == nil {
		return v
	}
	if p.Pane != nil {
		v.Pane = &front.Pane{Lines: paneLines(p.Pane)}
	} else {
		v.Pane = nil
	}
	v.Prompt, v.ViewText = paneRuns(p.Prompt), paneLines(p.View)
	v.Pictures = make([]front.Picture, 0, len(p.Layers))
	for _, l := range p.Layers {
		f := pix(l.Picture)
		f.File, f.Block, f.Frame, f.At = l.File, l.Block, l.Frame, l.At
		v.Pictures = append(v.Pictures, f)
	}
	return v
}
