package ui

import (
	"image"
	"image/color"

	"gobox/gpu"
	"gobox/gpu/vector"
)

// ScrollBar is every scrolled list's bar: N lines, Rows of them shown from
// Top, in R: a groove with a bevelled thumb.
type ScrollBar struct {
	R            image.Rectangle
	N, Rows, Top int
}

// BarColors are the bar's four colours.
type BarColors struct{ Groove, Thumb, Light, Dark color.RGBA }

// Thumb is the part of R the shown lines take.
func (s ScrollBar) Thumb() image.Rectangle {
	h := s.R.Dy()
	return image.Rect(s.R.Min.X, s.R.Min.Y+h*s.Top/s.N, s.R.Max.X, s.R.Min.Y+h*(s.Top+s.Rows)/s.N)
}

// To is the top that puts the thumb's middle at y, kept in range.
func (s ScrollBar) To(y int) int {
	return min(max(0, (y-s.R.Min.Y)*s.N/max(1, s.R.Dy())-s.Rows/2), max(0, s.N-s.Rows))
}

// Draw paints the bar with fill: the groove in R's middle third, the thumb
// with a light top and a dark bottom edge.
func (s ScrollBar) Draw(fill func(image.Rectangle, color.RGBA), c BarColors) {
	w, e := s.R.Dx()/3, max(1, s.R.Dx()/6)
	fill(image.Rect(s.R.Min.X+w, s.R.Min.Y, s.R.Max.X-w, s.R.Max.Y), c.Groove)
	t := s.Thumb()
	fill(t, c.Thumb)
	fill(image.Rect(t.Min.X, t.Min.Y, t.Max.X, t.Min.Y+e), c.Light)
	fill(image.Rect(t.Min.X, t.Max.Y-e, t.Max.X, t.Max.Y), c.Dark)
}

// DrawOn paints the bar on dst.
func (s ScrollBar) DrawOn(dst *gpu.Image, c BarColors) {
	s.Draw(func(r image.Rectangle, c color.RGBA) {
		vector.DrawFilledRect(dst, float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy()), c)
	}, c)
}

// BarHit is a bar drawn this frame and what moving its thumb does.
type BarHit struct {
	Bar ScrollBar
	Set func(top int)
}

// MenuBar is the bar of a menu at scale k whose first rows lines (rs, its
// rows' rectangles) show n lines from *top, in its right blank column. Its
// Set moves *top and keeps *sel (nil: none) among the lines shown. ok is
// false when no bar is needed.
func MenuBar(rs []image.Rectangle, n, rows int, top, sel *int, k float64) (h BarHit, ok bool) {
	if n <= rows || len(rs) < rows || rows == 0 {
		return h, false
	}
	r0, r1 := rs[0], rs[rows-1]
	b := ScrollBar{image.Rect(r0.Max.X-int(7*k), r0.Min.Y, r0.Max.X-int(k), r1.Max.Y), n, rows, *top}
	return BarHit{b, func(t int) {
		*top = t
		if sel != nil {
			*sel = min(max(*sel, t), t+rows-1)
		}
	}}, true
}

// Bars are the bars drawn this frame (Reset each frame, Add each) and the
// drag in progress.
type Bars struct {
	hits []BarHit
	drag *BarHit
}

func (b *Bars) Reset()       { b.hits = b.hits[:0] }
func (b *Bars) Add(h BarHit) { b.hits = append(b.hits, h) }

// Click starts a drag on a bar under (x, y): the thumb jumps there.
func (b *Bars) Click(x, y int) bool {
	for _, h := range b.hits {
		if image.Pt(x, y).In(h.Bar.R) {
			h.Set(h.Bar.To(y))
			b.drag = &h
			return true
		}
	}
	return false
}

// Hover drags the thumb taken by Click while the button is down.
func (b *Bars) Hover(y int, down bool) bool {
	h := b.drag
	if h == nil {
		return false
	}
	if !down {
		b.drag = nil
		return false
	}
	h.Set(h.Bar.To(y))
	return true
}
