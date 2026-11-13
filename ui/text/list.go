package text

// TextList is the cursor of a scrolling list of Total rows, Rows visible from
// Top, Index the selected row. Skip(i) marks rows the cursor passes over
// (headers, greyed lines). Same rules as the sotsb list menu.
type TextList struct {
	Top, Index, Rows, Total int
	Skip                    func(i int) bool
}

func (l *TextList) skipped(i int) bool {
	return i >= 0 && i < l.Total && l.Skip != nil && l.Skip(i)
}

// Step moves the cursor one row within the page, wrapping, without the
// header skip.
func (l *TextList) Step(forward bool) {
	if forward {
		l.Index++
		if l.Index > l.Top+l.Rows-1 {
			l.Index = l.Top
		}
		if l.Index > l.Total-1 {
			l.Index = l.Top
		}
	} else {
		l.Index--
		if l.Index < l.Top {
			l.Index = l.Top + l.Rows - 1
		}
		if l.Index > l.Total-1 {
			l.Index = l.Total - 1
		}
	}
}

// SkipRows steps past skipped rows, at most Rows times.
func (l *TextList) SkipRows(forward bool) {
	for k := 0; k < l.Rows; k++ {
		if l.Index < 0 || l.Index >= l.Total || !l.skipped(l.Index) {
			break
		}
		l.Step(forward)
	}
}

// Move is Step then SkipRows.
func (l *TextList) Move(forward bool) {
	l.Step(forward)
	l.SkipRows(forward)
}

// Page scrolls a page, keeping the cursor's row within the page.
func (l *TextList) Page(forward bool) {
	rel := l.Index - l.Top
	if forward {
		l.Top = min(l.Top+l.Rows, l.Total-l.Rows)
	} else {
		l.Top -= l.Rows
	}
	l.Top = max(l.Top, 0)
	l.Index = l.Top + rel
	l.SkipRows(forward)
}

// HasNext and HasPrev: there is a page after / before (leadingHdr: header
// rows that start the list, which Prev does not return to).
func (l *TextList) HasNext() bool               { return l.Total-l.Rows > l.Top }
func (l *TextList) HasPrev(leadingHdr int) bool { return l.Top > leadingHdr }
