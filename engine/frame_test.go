package engine

import (
	"testing"
)

// The 3D screen's text pane keeps the original's place on a page larger
// than 40x25 (its pop-up the pane's size).
func TestTextPaneFollowsPage(t *testing.T) {
	e := testEngine(t)
	defer setPageSize(minCols, minRows)
	setPageSize(43, 27)
	e.state().BeginFlag = 4
	e.clearPane()
	if e.pagePane() == nil || TextPane != (CellRect{X1: 1, Y1: msgTop, X2: 0x26, Y2: 0x16}) {
		t.Fatalf("frame %v pane %v", e.pagePane() != nil, TextPane)
	}
}
