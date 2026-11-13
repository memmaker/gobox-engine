package ui

import (
	"bytes"
	"gobox/ui/fonts"
	"image"
	"image/color"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

// Oracles: verbatim copies of the games' code (sotsb-go DisplayTextBlock,
// DrawBorder, drawMenuBar, list menu cursor) writing to a plain grid.
type oracle struct {
	cells          [][]TextCell
	msgCol, msgRow int
	fulls          int
}

func (o *oracle) char(x, y, colour int, ch byte) {
	if x > 0x27 || y > 0x18 {
		return
	}
	o.cells[y][x] = TextCell{Set: true, Upcase: true, C: ch, Colour: byte(colour)}
}

func (o *oracle) block(x1, y1, x2, y2, colour int, clear bool, s string) {
	if o.msgCol < x1 || o.msgCol > x2 || o.msgRow < y1 || o.msgRow > y2 {
		o.msgCol, o.msgRow = x1, y1
	}
	if clear {
		o.clear(x1, y1, x2, y2)
		o.msgCol, o.msgRow = x1, y1
	}
	n := len(s)
	if n > 255 {
		n = 255
	}
	if n == 0 {
		return
	}
	at := func(k int) byte { return s[k-1] }
	i := 1
	printSeg := func(j int) {
		for ; i <= j; i++ {
			o.char(o.msgCol, o.msgRow, colour, at(i))
			o.msgCol++
		}
	}
	for {
		j := i
		for j < n && punct(at(j)) {
			j++
		}
		for j < n && !punct(at(j)) && at(j) != ' ' {
			j++
		}
		if at(j) != ' ' {
			for j+1 < n && punct(at(j+1)) {
				j++
			}
		}
		if j-i+o.msgCol > x2 {
			o.msgCol = x1
			o.msgRow++
			for i < n && at(i) == ' ' {
				i++
			}
			if o.msgRow > y2 && i < n {
				o.msgCol, o.msgRow = x1, y1
				o.fulls++
				o.clear(x1, y1, x2, y2)
				printSeg(j)
			}
		} else {
			printSeg(j)
		}
		if i > n {
			break
		}
	}
	if o.msgCol > x2 {
		o.msgCol = x1
		o.msgRow++
	}
}

func punct(c byte) bool { return strings.IndexByte("!,-.:;?", c) >= 0 }

func (o *oracle) clear(x1, y1, x2, y2 int) {
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			o.cells[y][x] = TextCell{}
		}
	}
}

func newOracle() *oracle {
	o := &oracle{cells: make([][]TextCell, 25)}
	for y := range o.cells {
		o.cells[y] = make([]TextCell, 40)
	}
	return o
}

func TestTextBlockGolden(t *testing.T) {
	words := []string{"the", "Dragon,", "of", "fire-breathing", "gold!", "a", "treasure:", "x", "supercalifragilistic", "ok?", "  ", "."}
	r := rand.New(rand.NewSource(1))
	o, c := newOracle(), TextNewConsole(40, 25, TextTheme{})
	for n := 0; n < 300; n++ {
		var s string
		for k := r.Intn(40); k >= 0; k-- {
			s += words[r.Intn(len(words))] + " "
		}
		x1, y1 := r.Intn(10), r.Intn(8)
		x2, y2 := x1+5+r.Intn(25), y1+1+r.Intn(10)
		clear := r.Intn(3) == 0
		fulls := 0
		o.block(x1, y1, x2, y2, 7, clear, s)
		c.TextBlock(x1, y1, x2, y2, 7, clear, s, func() { fulls++ })
		if !reflect.DeepEqual(o.cells, c.Cells) || o.msgCol != c.Col || o.msgRow != c.Row {
			t.Fatalf("case %d (%q) differs", n, s)
		}
		if fulls != o.fulls {
			t.Fatalf("case %d: full %d times, want %d", n, fulls, o.fulls)
		}
		o.fulls = 0
	}
}

func TestWrap(t *testing.T) {
	got := TextWrap("The quick brown fox, jumps over the lazy dog.", 2, 15)
	want := []string{"The quick ", "brown fox, ", "jumps over ", "the lazy dog."}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%q", got)
	}
}

func TestBorderGolden(t *testing.T) {
	th := TextTheme{Border: func(s int) TextFrame { b := 0x114 + s*3; return TextFrame{b, b + 1, b + 2} }}
	c := TextNewConsole(40, 25, th)
	c.BoxTitled(3, 3, 20, 9, 1, 5, "Items")
	want := TextNewConsole(40, 25, th)
	b := 0x114 + 3
	x1, y1, x2, y2 := 2, 2, 21, 10 // DrawBorder's copy
	sym := func(x, y, id int) { want.Cells[y][x] = TextCell{Set: true, Symbol: true, Sym: id, Colour: 5} }
	for _, p := range [][2]int{{x1, y1}, {x2, y1}, {x1, y2}, {x2, y2}} {
		sym(p[0], p[1], b)
	}
	for x := x1 + 1; x <= x2-1; x++ {
		sym(x, y1, b+2)
		sym(x, y2, b+2)
	}
	for y := y1 + 1; y <= y2-1; y++ {
		sym(x1, y, b+1)
		sym(x2, y, b+1)
	}
	for i := 0; i < 5; i++ {
		want.Cells[3][(3+20-5)/2+i] = TextCell{Set: true, Upcase: true, C: "Items"[i], Colour: 5}
	}
	if !reflect.DeepEqual(c.Cells, want.Cells) {
		t.Fatal("border differs")
	}
}

func TestMenuBarGolden(t *testing.T) {
	menu := "Buy Sell Next Exit"
	items := TextMenuItems(menu)
	if len(items) != 4 || items[1] != (TextMenuItem{5, 8}) {
		t.Fatalf("%v", items)
	}
	c := TextNewConsole(40, 25, TextTheme{})
	c.MenuBar(7, 24, menu, items, 2, 2, 15, 0, 39)
	// verbatim drawMenuBar rule, cell by cell
	for i := 1; i <= len(menu); i++ {
		ch := menu[i-1]
		var w TextCell
		switch {
		case items[1].First <= i && i <= items[1].Last:
			w = TextCell{Set: true, C: ch, Colour: 0, Bg: 15}
		case TextIsHotkey(ch):
			w = TextCell{Set: true, C: ch, Colour: 15}
		default:
			w = TextCell{Set: true, C: ch, Colour: 2}
		}
		if c.Cells[24][7+i-1] != w {
			t.Fatalf("cell %d", i)
		}
	}
	if !c.Cells[24][39].Set || c.Cells[24][7+len(menu)-1+1].C != ' ' {
		t.Fatal("row not cleared")
	}
}

func TestListGolden(t *testing.T) {
	// sotsb lmStep/lmSkipHeaders/lmPageScroll copies, over a list with headers.
	skip := func(i int) bool { return i%7 == 0 }
	var top, idx int
	rows, total := 5, 23
	step := func(f bool) {
		if f {
			idx++
			if idx > top+rows-1 {
				idx = top
			}
			if idx > total-1 {
				idx = top
			}
		} else {
			idx--
			if idx < top {
				idx = top + rows - 1
			}
			if idx > total-1 {
				idx = total - 1
			}
		}
	}
	skipH := func(f bool) {
		for k := 0; k < rows; k++ {
			if idx >= 0 && idx < total && !skip(idx) {
				break
			}
			step(f)
		}
	}
	page := func(f bool) {
		rel := idx - top
		if f {
			top += rows
			if top > total-rows {
				top = total - rows
			}
		} else {
			top -= rows
		}
		if top < 0 {
			top = 0
		}
		idx = top + rel
		skipH(f)
	}
	l := &TextList{Rows: rows, Total: total, Skip: skip, Index: 1}
	idx = 1
	r := rand.New(rand.NewSource(2))
	for n := 0; n < 500; n++ {
		f := r.Intn(2) == 0
		switch r.Intn(3) {
		case 0:
			step(f)
			skipH(f)
			l.Move(f)
		case 1:
			page(f)
			l.Page(f)
		default:
			step(f)
			l.Step(f)
		}
		if l.Top != top || l.Index != idx {
			t.Fatalf("step %d: %d/%d want %d/%d", n, l.Top, l.Index, top, idx)
		}
	}
}

func TestLineHelpers(t *testing.T) {
	if got := TextPriceLine("Long Sword", "1,500", 20); got != "Long Sword     1,500" {
		t.Fatalf("%q", got)
	}
	if TextPadLeft("42", 5) != "   42" || TextCommas(1234567) != "1,234,567" || TextCommas(999) != "999" {
		t.Fatal("pad/commas")
	}
}

// glyphAtOracle is por-go present/remaster glyphAt, drawing one run character.
func glyphAtOracle(m *image.RGBA, g [8]byte, x, y int, fg, bg color.RGBA) {
	for r := 0; r < 8; r++ {
		for b := 0; b < 8; b++ {
			c := bg
			if g[r]&(0x80>>b) != 0 {
				c = fg
			}
			m.SetRGBA(x+b, y+r, c)
		}
	}
}

func TestConsoleDraw(t *testing.T) {
	pal := func(i byte) color.RGBA { return color.RGBA{i * 16, 255 - i, i, 255} }
	c := TextNewConsole(6, 2, TextTheme{})
	c.Text(0, 0, 10, "Hi [x]")
	c.Put(1, 1, 3, 4, 'q', false)
	got := image.NewRGBA(image.Rect(0, 0, 48, 16))
	want := image.NewRGBA(got.Rect)
	c.Draw(got, nil, pal, nil)
	for y, row := range c.Cells {
		for x, cl := range row {
			if cl.Set {
				g, _ := fonts.Glyph(nil, cl.C, cl.Upcase)
				glyphAtOracle(want, g, x*8, y*8, pal(cl.Colour), pal(cl.Bg))
			}
		}
	}
	if !bytes.Equal(got.Pix, want.Pix) {
		t.Fatal("pixels differ")
	}
}
