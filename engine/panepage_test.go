package engine

import (
	"gobox/asset"
	"strings"
	"testing"

	"gobox/game"
)

// Pane text longer than the pane is published whole, in lots gobox pages
// (Page.PaneRows), with no key wait of the game's own (BUG-CLASSES 11).
func TestPaneOverflowPaged(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	e.clearPane()
	e.paneText(asset.EGAWhite, true, strings.Repeat("word ", 46)+"LAST") // under the original's 255 characters
	p := e.statePage()
	if p.PaneRows != TextPaneRows || len(p.Pane) <= TextPaneRows {
		t.Fatalf("pane: %d lines, PaneRows %d", len(p.Pane), p.PaneRows)
	}
	if l := p.Pane[len(p.Pane)-TextPaneRows:]; !strings.Contains(joinRuns(l), "LAST") || !strings.Contains(joinRuns(p.Pane[:TextPaneRows]), "word") {
		t.Fatalf("the lots lost text: %d lines\nfirst %q\nlast %q", len(p.Pane), joinRuns(p.Pane[:TextPaneRows]), joinRuns(p.Pane[len(p.Pane)-TextPaneRows:]))
	}
	e.clearPane()
	if p := e.statePage(); p.PaneRows != 0 {
		t.Fatal("clearing the pane kept its earlier lots")
	}
}

func joinRuns(ls [][]game.TextRun) (s string) {
	for _, l := range ls {
		for _, r := range l {
			s += r.Text + " "
		}
	}
	return
}
