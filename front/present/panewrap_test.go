package present

import (
	"testing"

	"gobox/front"
)

func TestWrapPane(t *testing.T) {
	cols := paneBox.X2 - paneBox.X1 + 1
	long := "TERRINES HAVE SPOTTED YOU AND ARE MOVING TO INTERCEPT. NEO FORCES COME TO YOUR AID."
	pn := wrapPane(&front.Pane{Lines: [][]front.PaneRun{{{Text: long}}}})
	if len(pn.Lines) < 2 {
		t.Fatalf("not wrapped: %v", pn.Lines)
	}
	for _, l := range pn.Lines {
		for _, r := range l {
			if r.X+len(r.Text) > cols || r.Text[0] == ' ' || r.Text[len(r.Text)-1] == ' ' {
				t.Errorf("bad row %+v", r)
			}
		}
	}
	if got := paneCells(pn); got.Y2-got.Y1+1 < len(pn.Lines) && len(pn.Lines) <= maxPaneRows {
		t.Errorf("box %v too small for %d rows", got, len(pn.Lines))
	}
}
