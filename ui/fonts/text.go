package fonts

import (
	"image"
	"image/color"
	"os"
	"strings"
)

// GameFont is the Gold Box text font: 64 capitals at slots 0x20..0x7F, then
// 113 symbol tiles (tile k at slot 0x40+k).
const GameFont = "Gold_Box_8x8D1"

// Glyph is the 8x8 pattern of ch in font bits (nil or the GameFont's bytes:
// the game's own font), or false when there is none. text means ch is text:
// the game font shows a-z as capitals and has angle brackets for [ and ];
// a PC font shows the real character. Not text, ch is the raw symbol tile
// k (tiles from 177 up do not exist) and always comes from the game font.
func Glyph(bits []byte, ch byte, text bool) ([8]byte, bool) {
	var g [8]byte
	if len(bits) > 0 && text && ch >= 0x20 && ch < 0x7F {
		copy(g[:], bits[int(ch)*8:])
		return g, true
	}
	i := int(ch)
	if text {
		switch ch {
		case '[':
			ch = '<'
		case ']':
			ch = '>'
		}
		if ch >= 'a' && ch <= 'z' {
			ch -= 32
		}
		i = int(ch) % 0x40
	}
	game := Get(GameFont)
	if game == nil || i >= 177 {
		return g, false
	}
	copy(g[:], game[(0x40+i)*8:])
	return g, true
}

// DrawGlyph draws g at x, y of m: set bits fg, the rest bg (nil: left alone).
func DrawGlyph(m *image.RGBA, g [8]byte, x, y int, fg, bg *color.RGBA) {
	for r := 0; r < 8; r++ {
		for b := 0; b < 8; b++ {
			switch {
			case g[r]&(0x80>>b) != 0:
				m.SetRGBA(x+b, y+r, *fg)
			case bg != nil:
				m.SetRGBA(x+b, y+r, *bg)
			}
		}
	}
}

// DrawText draws s as text in font bits from x, y of m (8 pixels a
// character); bg nil leaves the pixels behind the glyphs alone.
func DrawText(m *image.RGBA, bits []byte, s string, x, y int, fg color.RGBA, bg *color.RGBA) {
	for i := 0; i < len(s); i++ {
		g, _ := Glyph(bits, s[i], true)
		DrawGlyph(m, g, x+i*8, y, &fg, bg)
	}
}

// Chooser is the font a game shows: the game's own font (Cur 0) or one of
// List (Cur-1). The game only says which; FONT.CFG (File) keeps the choice.
type Chooser struct {
	List []Font
	Cur  int
	File string // where the choice is kept, "" for nowhere
}

// NewChooser lists Load(dir) with first (a name) put first, the rest in
// Load's order, and picks the font named in file (if any).
func NewChooser(dir, first, file string) *Chooser {
	c := NewChooserOf(FontDir(dir), first)
	c.Keep(file)
	return c
}

// NewChooserOf lists own, then the pool, with first (a name) put first.
func NewChooserOf(own []Font, first string) *Chooser {
	c := &Chooser{}
	for _, f := range append(append([]Font{}, own...), All()...) {
		if Find(c.List, f.Name) < 0 {
			c.List = append(c.List, f)
		}
	}
	if i := Find(c.List, first); i > 0 {
		f := c.List[i]
		c.List = append([]Font{f}, append(c.List[:i], c.List[i+1:]...)...)
	}
	return c
}

// Keep keeps the choice in file from now on, and picks the font it names (if any).
func (c *Chooser) Keep(file string) {
	c.File = file
	if b, err := os.ReadFile(file); err == nil {
		if i := Find(c.List, strings.TrimSpace(string(b))); i >= 0 {
			c.Cur = i + 1
		}
	}
}

// Names are the fonts' names with the game's first, as label.
func (c *Chooser) Names(label string) []string {
	ns := []string{label}
	for _, f := range c.List {
		ns = append(ns, strings.ReplaceAll(f.Name, "_", " "))
	}
	return ns
}

// Set picks font i (0: the game's) and keeps the choice; out of range does
// nothing. The font stays picked when the choice cannot be kept: the error says so.
func (c *Chooser) Set(i int) error {
	if i < 0 || i > len(c.List) {
		return nil
	}
	c.Cur = i
	name := ""
	if i > 0 {
		name = c.List[i-1].Name
	}
	if c.File != "" {
		return os.WriteFile(c.File, []byte(name+"\n"), 0o644)
	}
	return nil
}

// Step moves to the next font (1) or the one before (-1), wrapping.
func (c *Chooser) Step(d int) {
	n := len(c.List) + 1
	c.Set((c.Cur + d + n) % n)
}

// Bits are font i's bytes (nil: the game's).
func (c *Chooser) Bits(i int) []byte {
	if i <= 0 || i > len(c.List) {
		return nil
	}
	return c.List[i-1].Bits
}

// Glyph is Glyph in font i.
func (c *Chooser) Glyph(i int, ch byte, text bool) ([8]byte, bool) { return Glyph(c.Bits(i), ch, text) }
