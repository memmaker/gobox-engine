package present

import (
	"strings"
	"time"

	"gobox/front"
)

const (
	blinkSwaps = 11                    // how many times a Blink run swaps between its ink and the highlight
	blinkMs    = 70 * time.Millisecond // the time between swaps
)

// blinked is v with the runs that Blink in the highlight colour (the
// hotkey's) on the swaps the clock is on: highlighted from the run's first
// showing, swapping blinkSwaps times to end on its own ink.
func (p *Presenter) blinked(v front.View) front.View {
	var sig strings.Builder
	each := func(f func(r *front.PaneRun)) {
		for i := range v.Prompt {
			f(&v.Prompt[i])
		}
		for _, ls := range [][][]front.PaneRun{v.ViewText, paneLinesOf(v.Pane)} {
			for _, l := range ls {
				for i := range l {
					f(&l[i])
				}
			}
		}
	}
	each(func(r *front.PaneRun) {
		if r.Blink {
			sig.WriteString(r.Text + "\x00")
		}
	})
	if sig.Len() == 0 {
		p.blinkSig = ""
		return v
	}
	if sig.String() != p.blinkSig {
		p.blinkSig, p.blinkAt = sig.String(), time.Now()
	}
	n := int(time.Since(p.blinkAt) / blinkMs)
	if n >= blinkSwaps || n%2 == 1 {
		return v
	}
	v.Prompt = append([]front.PaneRun(nil), v.Prompt...) // not the game's slices
	v.ViewText = copyLines(v.ViewText)
	if v.Pane != nil {
		c := *v.Pane
		c.Lines = copyLines(c.Lines)
		v.Pane = &c
	}
	each(func(r *front.PaneRun) {
		if r.Blink {
			r.Fg = col(inkWhite)
		}
	})
	return v
}

func paneLinesOf(p *front.Pane) [][]front.PaneRun {
	if p == nil {
		return nil
	}
	return p.Lines
}

func copyLines(ls [][]front.PaneRun) [][]front.PaneRun {
	c := make([][]front.PaneRun, len(ls))
	for i, l := range ls {
		c[i] = append([]front.PaneRun(nil), l...)
	}
	return c
}
