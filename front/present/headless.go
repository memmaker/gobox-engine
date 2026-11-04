// headless.go: what a game's cmd tests drive the Front with, so they need
// no import past gobox/front/present and /classic.

package present

import (
	"gobox/front"
	"gobox/gpu"
)

// The window's actions and boxes, as a cmd test sends and checks them.
type (
	Action   = front.Action
	CellRect = front.CellRect
)

const (
	ActResize    = front.ActResize
	ActCreate    = front.ActCreate
	PickHead     = front.PickHead
	PickBody     = front.PickBody
	PickIconHead = front.PickIconHead
	PickWeapon   = front.PickWeapon
)

// Frame runs one window frame headless (Act, then Draw onto an offscreen
// w by h picture kept across calls), as the shell does each frame.
func (f *Front) Frame(w, h int) {
	if f.off == nil || f.off.Bounds().Dx() != w || f.off.Bounds().Dy() != h {
		f.off = gpu.NewImage(w, h)
	}
	f.Act()
	f.Draw(f.off)
}
