package present

import (
	"testing"

	"gobox/front"
)

// A pane of 5 lines with Rows 2 shows 2, 2, then the last; a key moves on.
func TestPaneRows(t *testing.T) {
	var ls [][]front.PaneRun
	for _, s := range []string{"a", "b", "c", "d", "e"} {
		ls = append(ls, []front.PaneRun{{Text: s}})
	}
	p := &Presenter{}
	var got []string
	for i := 0; i < 3; i++ {
		v := p.paged(front.View{Pane: &front.Pane{Lines: ls, Rows: 2}})
		s := ""
		for _, l := range v.Pane.Lines {
			s += l[0].Text
		}
		got = append(got, s)
		if p.paneMore {
			p.Key(' ', front.View{Pane: &front.Pane{Lines: ls, Rows: 2}})
		}
	}
	if got[0] != "ab" || got[1] != "cd" || got[2] != "e" && got[2] != "de" || p.paneMore {
		t.Fatalf("%q more %v", got, p.paneMore)
	}
}

// A picture with Frames and FrameMs shows one of them; without, its own Frame.
func TestLayerFrames(t *testing.T) {
	p := &Presenter{}
	v := p.paged(front.View{Pictures: []front.Picture{{Frame: 9}, {Frame: 9, Frames: []int{3, 4}, FrameMs: 100}}})
	if f := v.Pictures[1].Frame; v.Pictures[0].Frame != 9 || f != 3 && f != 4 {
		t.Fatalf("%+v", v.Pictures)
	}
}
