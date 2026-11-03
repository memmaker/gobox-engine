package ui

import (
	"image"
	"testing"
)

func TestScrollBar(t *testing.T) {
	b := ScrollBar{image.Rect(0, 100, 6, 200), 40, 10, 0}
	if th := b.Thumb(); th.Min.Y != 100 || th.Max.Y != 125 {
		t.Errorf("thumb at the top = %v", th)
	}
	b.Top = 30
	if th := b.Thumb(); th.Max.Y != 200 {
		t.Errorf("thumb at the end = %v", th)
	}
	for _, c := range []struct{ y, top int }{{100, 0}, {50, 0}, {150, 15}, {200, 30}, {500, 30}} {
		if got := b.To(c.y); got != c.top {
			t.Errorf("To(%d) = %d, want %d", c.y, got, c.top)
		}
	}
}

func TestBars(t *testing.T) {
	top, sel := 0, 0
	rs := []image.Rectangle{image.Rect(0, 0, 60, 10), image.Rect(0, 10, 60, 20)}
	h, ok := MenuBar(rs, 10, 2, &top, &sel, 1)
	if !ok {
		t.Fatal("no bar")
	}
	var bs Bars
	bs.Add(h)
	if !bs.Click(55, 15) || top == 0 || sel < top {
		t.Fatalf("top %d sel %d", top, sel)
	}
	bs.Hover(0, true)
	if top != 0 {
		t.Fatal("drag did not move")
	}
	if bs.Hover(0, false) || bs.Hover(0, true) {
		t.Fatal("drag not released")
	}
}
