// headless.go: what a game's cmd tests drive the Front with, so they need
// no import past gobox/front/present and /classic.

package ui

import (
	"gobox/gpu"
)

// TitleDrawn: the last frame drew the main menu's screen, over the title's
// picture (Data.TitleBackdrop) when the game has one: never during play.
func (f *Front) TitleDrawn() bool { return f.P.mm.titleDrawn }

// Frame runs one window frame headless (Act, then Draw onto an offscreen
// w by h picture kept across calls), as the shell does each frame.
func (f *Front) Frame(w, h int) {
	if f.off == nil || f.off.Bounds().Dx() != w || f.off.Bounds().Dy() != h {
		f.off = gpu.NewImage(w, h)
	}
	f.Act()
	f.Draw(f.off)
}
