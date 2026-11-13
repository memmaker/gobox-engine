package present

import (
	"gobox/front"
	"image"

	"gobox/gpu"
	"gobox/ui"
)

// drawResults draws the after-combat page (Scene.Results) as a screen of
// its own: the outcome and experience, the hirelings' shares, what the
// party takes, the members harmed (name cyan, what red) and, when the party
// is destroyed, the end; its question as every page's.
func (p *Presenter) drawResults(dst *gpu.Image, v front.View, note Note) Rect {
	r := resultsOf(p.state())
	u := ui.Results{Outcome: r.Outcome, Exp: r.Exp, Shares: r.Shares, Taken: r.Taken, End: r.End}
	for _, h := range r.Harmed {
		u.Harmed = append(u.Harmed, ui.Harm(h))
	}
	p.fullPage(dst, &p.resultsPage, note, ui.ResultsTitle(u), func(m *image.RGBA, cw, ch int) {
		for _, run := range ui.ResultsRuns(u, cw, ch) {
			textAt(m, note, run.Text, run.X, run.Y, [...]int{inkGreen, inkCyan, inkRed}[run.Role])
		}
	})
	p.drawPageMenu(dst, note, v)
	b := dst.Bounds()
	return Rect{0, 0, float64(b.Dx()), float64(b.Dy())}
}
