package pool

import (
	"io/fs"

	"gobox/core"
)

// portraitPNG (extra): the PNG at path is the selected member's portrait, only
// on the sheet, the inventory and the creation: a stray drop elsewhere must never
// change a face. The original changes it only at creation and training.
func (e *Engine) portraitPNG(path string) {
	if !e.view.on && !e.snap.invOn && !e.ext.creating {
		return
	}
	b, ok := e.droppedPNG(path)
	if e.creation != nil { // the creation's draft takes it, the sheet is drawn from it
		e.creationEdit(func(cr *creation) error {
			if ok {
				cr.d.D.Portrait = b
			}
			return nil
		})
		return
	}
	if ok {
		e.do(core.SetPortrait{Who: e.selected(), PNG: b})
	}
	if e.view.on && !e.snap.invOn {
		e.SheetPortrait()
	}
}

// droppedPNG is the PNG file dropped at path, or false after saying why it
// is none.
func (e *Engine) droppedPNG(path string) ([]byte, bool) {
	b, err := readPNG(path)
	if err == nil && b == nil {
		err = fs.ErrNotExist
	}
	if err != nil {
		e.say("Not a PNG")
		return nil, false
	}
	return b, true
}
