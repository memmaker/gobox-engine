package ui

import (
	"cmp"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"gobox/asset"
	"gobox/game"
)

// The text fonts: Ctrl-F cycles from the game's own font through the
// font pool (every imported game's font, asset.PoolFonts, then gobox's),
// Ctrl-Shift-F goes back, the Esc menu's Font list picks one. The choice is
// kept in FONT.CFG beside the settings (KeepFonts).

const fontFile = "FONT.CFG"

// chooseFonts lists the font pool beside data's adventure folder, the
// game's own font (the pool's for the adventure folder's game, else the theme's) first.
func (f *Front) chooseFonts(data game.Data) {
	var pooled []asset.Font
	own, amiga := theme.Font, ""
	if data.Adventure != "" {
		root, id := filepath.Split(filepath.Clean(data.Adventure)) // the folder's name is the import's game ID
		l, name, err := asset.PoolFonts(root, id)
		if err != nil {
			panic(fmt.Sprintf("%v (run the import again)", err))
		}
		_, amiga, err = asset.PoolFonts(root, asset.PlatformID(id, asset.PlatformAmiga))
		if err != nil {
			panic(fmt.Sprintf("%v (run the import again)", err))
		}
		pooled = l
		own = cmp.Or(name, own)
	}
	c := asset.NewChooserOf(pooled, own)
	f.P.fonts = c
	game := c.Bits(asset.Find(c.List, own) + 1) // nil (not found): gobox's GameFont
	amigaGame := c.Bits(asset.Find(c.List, amiga) + 1)
	bits := func() []byte {
		if b := c.Bits(c.Cur); b != nil {
			return b
		}
		if amigaGame != nil && amigaShown() { // the game's own font follows the art (F12): the Amiga release's with its art
			return amigaGame
		}
		return game
	}
	f.Glyph = func(ch byte) [8]byte { g, _ := asset.Glyph(bits(), ch, true); return g }
	f.Font = func(ch byte, text bool) [8]byte { g, _ := asset.Glyph(bits(), ch, text); return g }
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
	bits := asset.Get(name)
	return func(ch byte, text bool) [8]byte { g, _ := asset.Glyph(bits, ch, text); return g }
}
