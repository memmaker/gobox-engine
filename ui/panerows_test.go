package ui

import (
	"testing"

	"gobox/game"
)

// pagePlayed is a Presenter playing a moment that holds pg.
func pagePlayed(pg *game.Page) *Presenter {
	p, _ := presenterOn(game.Scene{})
	p.played = &frame{Moment: Moment{Report: game.Report{Page: pg}}}
	return p
}

// A pane of 5 lines with Rows 2 shows 2, 2, then the last; a key moves on.
func TestPaneRows(t *testing.T) {
	var ls [][]game.TextRun
	for _, s := range []string{"a", "b", "c", "d", "e"} {
		ls = append(ls, []game.TextRun{{Text: s}})
	}
	p := pagePlayed(&game.Page{Pane: ls, PaneRows: 2})
	var got []string
	for i := 0; i < 3; i++ {
		s := ""
		for _, l := range p.pane().Lines {
			s += l[0].Text
		}
		got = append(got, s)
		if p.panePaged() {
			p.Key(' ')
		}
	}
	if got[0] != "ab" || got[1] != "cd" || got[2] != "e" && got[2] != "de" || p.panePaged() {
		t.Fatalf("%q more %v", got, p.panePaged())
	}
}

// A picture with Frames and FrameMs shows one of them; without, its own Frame.
func TestLayerFrames(t *testing.T) {
	p := pagePlayed(&game.Page{Layers: []game.Layer{{Picture: game.Picture{Frame: 9}}, {Picture: game.Picture{Frame: 9}, Frames: []int{3, 4}, FrameMs: 100}}})
	ps := p.pictures()
	if f := ps[1].Frame; ps[0].Frame != 9 || f != 3 && f != 4 {
		t.Fatalf("%+v", ps)
	}
}
