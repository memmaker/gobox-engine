package ui

import (
	"strconv"
	"strings"
)

// Cell is one character cell of a Console: the character written (Upcase:
// drawn as its upcased glyph), its colour and background; Symbol cells hold
// a border/tile symbol id in Sym instead of a character.
type Cell struct {
	Set, Upcase, Symbol bool
	C, Colour, Bg       byte
	Sym                 int
}

// Frame is the symbol ids of one border style: corner, vertical, horizontal.
type Frame struct{ Corner, Vert, Horz int }

// BorderTheme is the game's border data: Border(style) gives that style's symbols.
type BorderTheme struct{ Border func(style int) Frame }

// Console is a grid of cells with a text cursor (Col, Row) for TextBlock.
// It draws nothing: the game renders Cells with its own font and palette.
type Console struct {
	Cols, Rows int
	Cells      [][]Cell
	Col, Row   int
	Theme      BorderTheme
}

func NewConsole(cols, rows int, t BorderTheme) *Console {
	c := &Console{Cols: cols, Rows: rows, Theme: t, Cells: make([][]Cell, rows)}
	for y := range c.Cells {
		c.Cells[y] = make([]Cell, cols)
	}
	return c
}

// Put writes ch at (x, y); out of the grid it does nothing.
func (c *Console) Put(x, y, colour, bg int, ch byte, upcase bool) {
	if x < 0 || x >= c.Cols || y < 0 || y >= c.Rows {
		return
	}
	c.Cells[y][x] = Cell{Set: true, Upcase: upcase, C: ch, Colour: byte(colour), Bg: byte(bg)}
}

// PutSymbol writes border symbol id at (x, y).
func (c *Console) PutSymbol(x, y, colour, id int) {
	if x < 0 || x >= c.Cols || y < 0 || y >= c.Rows {
		return
	}
	c.Cells[y][x] = Cell{Set: true, Symbol: true, Sym: id, Colour: byte(colour)}
}

// Text writes s from (x, y), each character upcased, no wrapping (at most 255).
func (c *Console) Text(x, y, colour int, s string) { c.TextBg(x, y, colour, 0, s) }

// TextBg is Text on background bg.
func (c *Console) TextBg(x, y, colour, bg int, s string) {
	for i := 0; i < len(s) && i < 255; i++ {
		c.Put(x+i, y, colour, bg, s[i], true)
	}
}

// Blank empties the cells [x1..x2] x [y1..y2].
func (c *Console) Blank(x1, y1, x2, y2 int) {
	for y := max(y1, 0); y <= y2 && y < c.Rows; y++ {
		for x := max(x1, 0); x <= x2 && x < c.Cols; x++ {
			c.Cells[y][x] = Cell{}
		}
	}
}

// Border draws style's frame around the cell rectangle (one cell outside it).
func (c *Console) Border(x1, y1, x2, y2, style, colour int) {
	f := c.Theme.Border(style)
	x1, y1, x2, y2 = x1-1, y1-1, x2+1, y2+1
	for _, p := range [][2]int{{x1, y1}, {x2, y1}, {x1, y2}, {x2, y2}} {
		c.PutSymbol(p[0], p[1], colour, f.Corner)
	}
	for x := x1 + 1; x < x2; x++ {
		c.PutSymbol(x, y1, colour, f.Horz)
		c.PutSymbol(x, y2, colour, f.Horz)
	}
	for y := y1 + 1; y < y2; y++ {
		c.PutSymbol(x1, y, colour, f.Vert)
		c.PutSymbol(x2, y, colour, f.Vert)
	}
}

// BoxTitled blanks the interior, draws the border (colour) and centres the
// title on row y1 (the interior's first row, as the games do).
func (c *Console) BoxTitled(x1, y1, x2, y2, style, colour int, title string) {
	title = title[:min(len(title), c.Cols)]
	c.Blank(x1, y1, x2, y2)
	c.Border(x1, y1, x2, y2, style, colour)
	if title != "" {
		c.Text((x1+x2-len(title))/2, y1, colour, title)
	}
}

// TextBlock writes s word wrapped in the cell box, continuing at the cursor
// when it lies inside the box (clear: blank the box first and start at its
// corner). When the box is full, full is called (the game's "press a key"),
// then the box is blanked and starts over.
func (c *Console) TextBlock(x1, y1, x2, y2, colour int, clear bool, s string, full func()) {
	if c.Col < x1 || c.Col > x2 || c.Row < y1 || c.Row > y2 {
		c.Col, c.Row = x1, y1
	}
	if clear {
		c.Blank(x1, y1, x2, y2)
		c.Col, c.Row = x1, y1
	}
	Flow(s, x1, y1, x2, y2, &c.Col, &c.Row, func(x, y int, ch byte) { c.Put(x, y, colour, 0, ch, true) }, func() {
		if full != nil {
			full()
		}
		c.Blank(x1, y1, x2, y2)
	})
}

// MenuItem is the span of one bottom-row menu item: from its hotkey to two
// cells before the next hotkey (1-based positions in the menu text).
type MenuItem struct{ First, Last int }

// IsHotkey: the characters that start a menu item ([0-9A-Z]).
func IsHotkey(c byte) bool { return c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' }

// MenuItems splits menu text into its items; every hotkey starts one.
func MenuItems(menu string) []MenuItem {
	items := []MenuItem{{}}
	for i := 1; i <= len(menu); i++ {
		if !IsHotkey(menu[i-1]) {
			continue
		}
		cur := &items[len(items)-1]
		if cur.First == 0 {
			cur.First = i
		} else {
			cur.Last = i - 2
			items = append(items, MenuItem{First: i})
		}
	}
	items[len(items)-1].Last = len(menu)
	return items
}

// MenuBar draws menu text at (x0, y): the selected item (1-based) black on
// highlight, hotkeys in highlight, the rest in normal; the row cleared to
// lastCol after it. black is the game's black.
func (c *Console) MenuBar(x0, y int, menu string, items []MenuItem, sel, normal, highlight, black, lastCol int) {
	if len(menu) == 0 {
		return
	}
	for i := 1; i <= len(menu); i++ {
		ch := menu[i-1]
		switch {
		case items[sel-1].First <= i && i <= items[sel-1].Last && highlight != 0:
			c.Put(x0+i-1, y, black, highlight, ch, false)
		case IsHotkey(ch):
			c.Put(x0+i-1, y, highlight, black, ch, false)
		default:
			c.Put(x0+i-1, y, normal, black, ch, false)
		}
	}
	for x := x0 + len(menu); x <= lastCol; x++ {
		c.Put(x, y, black, black, ' ', false)
	}
}

// Punct is the set of characters that end a word for TextBlock.
func Punct(c byte) bool { return strings.IndexByte("!,-.:;?", c) >= 0 }

// Flow lays s out word wrapped in the box x1..x2, y1..y2 from the cursor
// (*col, *row): put gets each character's cell; when the box is full, full
// is called and the box starts over.
func Flow(s string, x1, y1, x2, y2 int, col, row *int, put func(x, y int, c byte), full func()) {
	n := min(len(s), 255)
	if n == 0 {
		return
	}
	at := func(k int) byte { return s[k-1] } // 1-based
	i := 1
	printSeg := func(j int) {
		for ; i <= j; i++ {
			put(*col, *row, at(i))
			*col++
		}
	}
	for {
		j := i
		for j < n && Punct(at(j)) {
			j++
		}
		for j < n && !Punct(at(j)) && at(j) != ' ' {
			j++
		}
		if at(j) != ' ' {
			for j+1 < n && Punct(at(j+1)) {
				j++
			}
		}
		if j-i+*col > x2 {
			*col = x1
			*row++
			for i < n && at(i) == ' ' {
				i++
			}
			if *row > y2 && i < n {
				*col, *row = x1, y1
				full()
				printSeg(j)
			}
		} else {
			printSeg(j)
		}
		if i > n {
			break
		}
	}
	if *col > x2 {
		*col = x1
		*row++
	}
}

// Wrap is s word-wrapped between columns x1 and x2, from x1: the rows' text,
// each from column x1.
func Wrap(s string, x1, x2 int) []string {
	var rows [][]byte
	col, row := x1, 0
	Flow(s, x1, 0, x2, 1<<30, &col, &row, func(x, y int, c byte) {
		for len(rows) <= y {
			rows = append(rows, nil)
		}
		for len(rows[y]) < x-x1 {
			rows[y] = append(rows[y], ' ')
		}
		rows[y] = append(rows[y], c)
	}, func() {})
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = string(r)
	}
	return out
}

// PadLeft right-aligns s in width columns.
func PadLeft(s string, width int) string {
	if len(s) < width {
		s = strings.Repeat(" ", width-len(s)) + s
	}
	return s
}

// PriceLine is name left and price right-aligned in width columns (a shop
// line); the price is cut to 16 characters, the name to what fits.
func PriceLine(name, price string, width int) string {
	price = price[:min(len(price), 16)]
	line := []byte(strings.Repeat(" ", width))
	copy(line[width-len(price):], price)
	copy(line, name[:min(len(name), width)])
	return string(line)
}

// Commas is the decimal digits of n with commas every three.
func Commas(n int) string {
	s := strconv.Itoa(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}
