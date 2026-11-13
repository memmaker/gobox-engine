package engine

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"gobox/adventure"
	"gobox/game"
)

// Text fonts (extra): the font files in data/assets/fonts are published as
// game.Data.Fonts; gobox picks, draws and keeps the choice (Ctrl-F cycles
// from the game's own font, Ctrl-Shift-F goes back). A PC font draws text
// in its real case (the game's font has capitals only); the game's symbols
// (borders, cursors) keep its own glyphs.

// fontFaces are the 8x8 fonts (.f8, 256 glyphs of 8 bytes) in data/assets/fonts;
// never nil, so gobox offers the font menu.
func (e *Engine) fontFaces() []game.Typeface {
	faces := []game.Typeface{}
	ents, _ := os.ReadDir(filepath.Join(e.data, adventure.AssetsDir, "fonts"))
	for _, d := range ents {
		if !strings.EqualFold(filepath.Ext(d.Name()), ".f8") {
			continue
		}
		if b, err := os.ReadFile(filepath.Join(e.data, adventure.AssetsDir, "fonts", d.Name())); err == nil && len(b) == 256*8 {
			faces = append(faces, game.Typeface{Name: strings.TrimSuffix(d.Name(), filepath.Ext(d.Name())), Bits: b})
		}
	}
	return faces
}

// Note is a short message for the window to show over the picture (the
// font name after Ctrl-F) and when it was set.
func (e *Engine) Note() (string, time.Time) {
	e.present()
	if e.restLine && time.Since(e.ext.noteAt) > 2*time.Second { // extra: the rest's time left, live in the flash's place
		r := e.restUnit
		return "Rest time " + twoDigits(r(4)) + ":" + twoDigits(r(3)) + ":" + twoDigits(10*r(2)+r(1)), time.Now()
	}
	return e.ext.note, e.ext.noteAt
}
