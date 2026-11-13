package engine

import (
	"gobox/game"
	"strings"
	"testing"
)

// Pane text longer than the pane is published whole, its earlier lots
// first, with no key wait of the game's own (BUG-CLASSES 11).
func TestPaneOverflowPaged(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	e.clearPane()
	e.paneText(game.InkWhite, true, strings.Repeat("word ", 46)+"LAST") // under the original's 255 characters
	p := e.statePage()
	if len(p.Pane) <= paneRows {
		t.Fatalf("pane: %d lines", len(p.Pane))
	}
	if l := p.Pane[len(p.Pane)-paneRows:]; !strings.Contains(strings.Join(l, " "), "LAST") || !strings.Contains(strings.Join(p.Pane[:paneRows], " "), "word") {
		t.Fatalf("the lots lost text: %d lines\nfirst %q\nlast %q", len(p.Pane), p.Pane[:paneRows], p.Pane[len(p.Pane)-paneRows:])
	}
	e.clearPane()
	if p := e.statePage(); len(p.Pane) > paneRows {
		t.Fatal("clearing the pane kept its earlier lots")
	}
}
