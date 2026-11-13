package present

import (
	"gobox/front"
	"image"
	"os"
	"slices"
	"strings"
	"testing"
)

// The "edges" credits: one frame closing on the page's edges, the original's
// rows spread over the taller page, the block centred, the remaster's lines
// in the rows above the bottom frame.
func TestCreditEdges(t *testing.T) {
	old, oldBoxes := credits, creditBoxes
	defer func() { credits, creditBoxes = old, oldBoxes }()
	credits, creditBoxes = []Credit{{X: 2, Y: 1, Colour: 14, Text: "TITLE"}, {X: 2, Y: 3, Colour: 15, Text: "LEFT"}, {X: 21, Y: 3, Colour: 15, Text: "RIGHT COLUMN"}, {X: 2, Y: 23, Colour: 10, Text: "LAST"}}, nil
	c := screenPage(front.View{Cells: image.Pt(60, 40), Credits: true})
	if !slices.Contains(c.boxes, front.CellRect{X1: 1, Y1: 1, X2: 58, Y2: 38}) {
		t.Fatalf("no frame on the page's edges: %v", c.boxes)
	}
	var dump []string
	at := map[string]image.Point{}
	for y, l := range c.lines() {
		row := []byte(strings.Repeat(" ", c.cols))
		for _, r := range l {
			copy(row[r.X:], r.Text)
			at[strings.TrimSpace(r.Text)] = image.Pt(r.X, y)
		}
		dump = append(dump, string(row))
	}
	if d := os.Getenv("CREDIT_DUMP"); d != "" {
		os.WriteFile(d, []byte(strings.Join(dump, "\n")), 0o644)
	}
	// The block is x 2..33 (31 wide): centred on 60 it starts at 14.
	if at["TITLE"] != image.Pt(14, 1) || at["LAST"] != image.Pt(14, 34) || at["RIGHT COLUMN"].X != 33 {
		t.Errorf("lines at %v", at)
	}
	if at[remasterLines[0]] != image.Pt(20, 36) || at[remasterLines[1]].Y != 37 { // 60 cells: too narrow for one row
		t.Errorf("remaster lines at %v", at)
	}
}
