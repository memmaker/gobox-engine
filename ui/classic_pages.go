package ui

import (
	"image"
	"image/color"
	"strconv"
	"strings"

	"gobox/ega"
)

// The journal pageCut's lines (a Read's Lines): "\x00T" a title, "\x01N" a
// picture N whose rows follow as "\x02", "\x03K\x03text" a voiced paragraph's
// first line (speech clip K), else plain text.
const (
	jrWidth = 0x24 // the text's width in columns
	jrTop   = 3    // the row of a page's first line
)

// pageCut is the page the journal is laid out on: cols by rows cells, ink its colours.
type pageCut struct {
	cols, rows int
	ink        [16]color.RGBA
}

func (p pageCut) jrRows() int { return p.rows - 3 - jrTop }

// JournalPage is page n of the journal text lines under the box title title
// on a cols by rows page in colours ink (all zero: the EGA's own), its tall
// picture scrolled off pixels down, the pictures from scan, and how far that
// picture scrolls (0: none). n past the last page is the last page.
func JournalPage(title string, lines []string, n, off, cols, rows int, ink [16]color.RGBA, scan func(string) image.Image) (*Pages, int) {
	p := pageCut{cols: max(cols, 40), rows: max(rows, 25), ink: ink}
	if ink == ([16]color.RGBA{}) {
		for i := range p.ink {
			p.ink[i] = ega.RGBA(i)
		}
	}
	pages := p.pages(lines)
	n = max(min(n, len(pages)-1), 0)
	page := pages[n]
	v := &Pages{Title: p.boxTitle(title), Top: jrTop, Page: n, Pages: len(pages)}
	maxOff := 0
	for i, l := range page {
		var line []PaneRun
		switch {
		case strings.HasPrefix(l, "\x00"):
			line = []PaneRun{p.textRun(2, 14, l[1:])}
		case strings.HasPrefix(l, "\x03"):
			_, text, _ := strings.Cut(l[1:], "\x03")
			line = []PaneRun{p.textRun(2, 15, text)}
		case strings.HasPrefix(l, "\x01"):
			if sc, o := p.scan(scan, l[1:], 2*8+jrWidth*4, (jrTop+i)*8, pictureRows(page[i:])*8, off); sc != nil {
				v.Scans = append(v.Scans, *sc)
				maxOff = max(maxOff, o)
			}
		case l != "\x02":
			line = []PaneRun{p.textRun(2, 15, l)}
		}
		v.Lines = append(v.Lines, line)
	}
	more := "press a key"
	if n < len(pages)-1 {
		more = "more: " + strconv.Itoa(n+1) + "/" + strconv.Itoa(len(pages))
	}
	if maxOff > 0 {
		more = "up/down: scroll, " + more
	}
	v.Hint = p.textRun(p.cols-3-len(more), 10, more)
	return v, maxOff
}

// SpeakKeys are the speech clips of page n of the journal text lines, in
// order: its title (spaces as "_") and the paragraphs that start on it.
func SpeakKeys(lines []string, n, cols, rows int) (keys []string) {
	pages := pageCut{cols: max(cols, 40), rows: max(rows, 25)}.pages(lines)
	for _, l := range pages[max(min(n, len(pages)-1), 0)] {
		switch {
		case strings.HasPrefix(l, "\x00"):
			keys = append(keys, strings.ReplaceAll(l[1:], " ", "_"))
		case strings.HasPrefix(l, "\x03"):
			key, _, _ := strings.Cut(l[1:], "\x03")
			keys = append(keys, key)
		}
	}
	return keys
}

// textRun is s at column x in colour: on black, upcased, cut at the page's edge.
func (p pageCut) textRun(x, colour int, s string) PaneRun {
	s = s[:max(0, min(len(s), 255, p.cols-x))]
	return PaneRun{X: x, Text: s, Fg: p.ink[colour], Bg: p.ink[0], Upcase: true}
}

// boxTitle is the title centred on the top row of a box round the whole page.
func (p pageCut) boxTitle(t string) PaneRun {
	t = t[:min(len(t), 0x28)]
	return p.textRun((1+p.cols-2-len(t))/2, 15, t)
}

// pages cuts lines into pages of at most jrRows; a page does not start or
// end with a blank line or end with a title, and a picture is not split:
// one that overruns the page shrinks into the rows left when they hold half
// of it, else it starts the next page.
func (p pageCut) pages(lines []string) [][]string {
	lines = append([]string(nil), lines...)
	var pages [][]string
	for len(lines) > 0 {
		for len(lines) > 0 && lines[0] == "" {
			lines = lines[1:]
		}
		for i := 0; i < min(p.jrRows(), len(lines)); i++ {
			if !strings.HasPrefix(lines[i], "\x01") {
				continue
			}
			rows := pictureRows(lines[i:])
			if left := p.jrRows() - i; rows > left && left*2 >= rows {
				lines = append(lines[:i+left], lines[i+rows:]...) // drop the rows it no longer covers
			}
		}
		n := min(p.jrRows(), len(lines))
		if n < len(lines) && lines[n] != "" { // a voiced paragraph is not split, unless it fills the page
			k := n - 1
			for k > 0 && lines[k] != "" && !strings.HasPrefix(lines[k], "\x03") {
				k--
			}
			if k > 0 && strings.HasPrefix(lines[k], "\x03") {
				n = k
			}
		}
		for n < len(lines) && n > 1 && lines[n] == "\x02" { // back to the picture's first row
			n--
		}
		for n < len(lines) && n > 1 && (lines[n-1] == "" || strings.HasPrefix(lines[n-1], "\x00")) {
			n--
		}
		if n > 0 {
			pages = append(pages, lines[:n])
		}
		lines = lines[n:]
	}
	return pages
}

// pictureRows counts the rows of the picture starting lines.
func pictureRows(lines []string) int {
	n := 1
	for n < len(lines) && lines[n] == "\x02" {
		n++
	}
	return n
}

// pictureSize is the screen area a picture covers: the scans are 300 dpi,
// shown at about 90 screen pixels an inch, at most the text width wide and,
// unless taller than wide (those scroll), 17 rows high.
func pictureSize(im image.Image) (w, h int) {
	b := im.Bounds()
	s := min(0.3, float64(jrWidth*8)/float64(b.Dx()))
	if !tall(im) {
		s = min(s, 136/float64(b.Dy()))
	}
	return int(float64(b.Dx()) * s), int(float64(b.Dy()) * s)
}

func tall(im image.Image) bool { return im.Bounds().Dy() > im.Bounds().Dx() }

// scan places the picture name centred on x, top at y, at most maxH high:
// the scan on paper white. One taller than that shrinks to fit, or, if
// taller than wide, shows maxH of it from off pixels down; it returns how
// far it scrolls (0: not at all). nil: no such picture.
func (p pageCut) scan(get func(string) image.Image, name string, x, y, maxH, off int) (*Scan, int) {
	if get == nil {
		return nil, 0
	}
	im := get(name)
	if im == nil {
		return nil, 0
	}
	maxOff := 0
	w, h := pictureSize(im)
	src := im.Bounds()
	switch {
	case h > maxH && tall(im):
		maxOff = h - maxH
		off = min(off, maxOff)
		src.Min.Y, src.Max.Y = src.Min.Y+off*src.Dy()/h, src.Min.Y+(off+maxH)*src.Dy()/h
		h = maxH
	case h > maxH:
		w, h = w*maxH/h, maxH
	}
	return &Scan{Name: name, Img: im, Src: src, At: image.Rect(x-w/2, y, x-w/2+w, y+h), Paper: p.ink[15]}, maxOff
}
