package ui

import (
	"log"
	"path/filepath"
	"time"

	"gobox/game"
	"gobox/ui/fonts"
)

// The text fonts (Data.Fonts): Ctrl-F cycles from the game's own font
// through the game's fonts and the pool, Ctrl-Shift-F goes back, the Esc
// menu's Font list picks one. The choice is kept in FONT.CFG beside the
// settings (KeepFonts).

const fontFile = "FONT.CFG"

// chooseFonts lists the game's fonts (nil Data.Fonts: the game does its own).
func (f *Front) chooseFonts(own []game.Typeface) {
	if own == nil {
		return
	}
	var l []fonts.Font
	for _, t := range own {
		l = append(l, fonts.Font{Name: t.Name, Bits: t.Bits})
	}
	first := ""
	if len(l) > 0 {
		first = l[0].Name
	}
	c := fonts.NewChooserOf(l, first)
	f.P.fonts = c
	game := fonts.Get(theme.Font) // the game's own font: the theme's, else gobox's
	bits := func() []byte {
		if b := c.Bits(c.Cur); b != nil {
			return b
		}
		return game
	}
	f.Glyph = func(ch byte) [8]byte { g, _ := fonts.Glyph(bits(), ch, true); return g }
	f.Font = func(ch byte, text bool) [8]byte { g, _ := fonts.Glyph(bits(), ch, text); return g }
}

// KeepFonts keeps the font choice in dir (app calls it with the settings'
// folder) and names the game's own font title.
func (f *Front) KeepFonts(dir, title string) {
	if f.P.fonts != nil {
		f.P.fonts.Keep(filepath.Join(dir, fontFile))
		f.P.fontLabel = title
	}
}

// FontMenu are the Esc menu's Font list: the names, the one in use, a pick, a font's glyph for the preview.
func (f *Front) FontMenu() (names func() []string, at func() int, set func(int), glyph func(int, byte) [8]byte) {
	c := f.P.fonts
	if c == nil {
		return nil, nil, nil, nil
	}
	// the menu has no place for an error: the choice stands, the failed keeping is logged
	set = func(i int) {
		if err := c.Set(i); err != nil {
			log.Printf("font choice not kept: %v", err)
		}
	}
	return func() []string { return c.Names(f.P.fontLabel) }, func() int { return c.Cur }, set,
		func(i int, ch byte) [8]byte { g, _ := c.Glyph(i, ch, true); return g }
}

// cycleFont moves to the next font (d 1) or the one before (-1) and shows
// its name for a moment; false: the game has no font list here.
func (p *Presenter) cycleFont(d int) bool {
	if p.fonts == nil {
		return false
	}
	p.fonts.Step(d)
	p.fontNote, p.fontAt = p.fonts.Names(p.fontLabel)[p.fonts.Cur], time.Now()
	return true
}

// PoolGlyph is the glyph lookup of the pool font called name (MapImage's glyph).
func PoolGlyph(name string) func(ch byte, text bool) [8]byte {
	bits := fonts.Get(name)
	return func(ch byte, text bool) [8]byte { g, _ := fonts.Glyph(bits, ch, text); return g }
}
