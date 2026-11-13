package present

import "gobox/front"

// maxPaneRows is the most rows the text box grows to (about half the screen).
const maxPaneRows = 12

// paneCells is paneBox grown to hold pn's lines (up to maxPaneRows).
func paneCells(pn *front.Pane) front.FrontCellRect {
	r := paneBox
	if pn != nil {
		r.Y2 = max(r.Y2, r.Y1+min(len(pn.Lines), maxPaneRows)-1)
	}
	return r
}

// wrapPane is pn with every line broken at spaces to fit the box's columns
// (games hand the text over unwrapped); pn itself when all fit.
func wrapPane(pn *front.Pane) *front.Pane {
	cols := paneBox.X2 - paneBox.X1 + 1
	var out [][]front.PaneRun
	changed := false
	for _, l := range pn.Lines {
		w := wrapLine(l, cols)
		changed = changed || len(w) != 1
		out = append(out, w...)
	}
	if !changed {
		return pn
	}
	c := *pn
	c.Lines = out
	return &c
}

type paneCh struct {
	x   int
	c   byte
	run front.PaneRun // the char's run (Text and X unused)
}

// wrapLine breaks line l into rows of cols columns, at the last space that
// fits (else mid-word); later rows start at the first run's column.
func wrapLine(l []front.PaneRun, cols int) [][]front.PaneRun {
	var cs []paneCh
	for _, r := range l {
		for i := 0; i < len(r.Text); i++ {
			cs = append(cs, paneCh{r.X + i, r.Text[i], r})
		}
	}
	if len(cs) == 0 || cs[len(cs)-1].x < cols {
		return [][]front.PaneRun{l}
	}
	indent := cs[0].x
	var rows [][]front.PaneRun
	for len(cs) > 0 {
		x0, e := cs[0].x, 0
		for e < len(cs) && cs[e].x-x0+indent < cols {
			e++
		}
		n := e
		if e < len(cs) {
			for i := e; i > 0; i-- { // the last space that fits
				if cs[i].c == ' ' {
					n = i
					break
				}
			}
		}
		n = max(n, 1)
		var row []front.PaneRun
		for _, ch := range cs[:n] {
			if k := len(row) - 1; k >= 0 && row[k].X+len(row[k].Text) == ch.x-x0+indent && sameStyle(row[k], ch.run) {
				row[k].Text += string(ch.c)
				continue
			}
			r := ch.run
			r.X, r.Text = ch.x-x0+indent, string(ch.c)
			row = append(row, r)
		}
		rows = append(rows, row)
		cs = cs[n:]
		for len(cs) > 0 && cs[0].c == ' ' { // the spaces at the break
			cs = cs[1:]
		}
	}
	return rows
}

func sameStyle(a, b front.PaneRun) bool {
	return a.Fg == b.Fg && a.Bg == b.Bg && a.Upcase == b.Upcase && a.Blink == b.Blink
}
