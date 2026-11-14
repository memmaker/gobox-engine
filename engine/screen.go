package engine

import "gobox/rules"

// The adventure screen: the view window and what the screen is rebuilt with.
// Ported from 00BA:272A/27D9/290B and ovr27.

// Redraw = ovr27:016D: redraw the view window for the current view mode.
func (e *Engine) Redraw() {
	mode := e.viewMode
	if e.worldWord(rules.WordIn3D) != 0 {
		mode = 1
	}
	switch {
	case mode == 1:
		e.es.cell = e.MapCell(e.posX, e.posY)
		e.show3D() // ovr30:05FD Draw3DView: the remaster draws the view
		e.viewDrawn()
	case mode >= 2 && mode <= 4: // ovr27:0250: the area map round the party, its marker (ovr27:0000) on it
		e.DrawAreaMap()
	}
}

// DrawMainScreen = 00BA:27D9: rebuild the whole screen for DS:49F3.
func (e *Engine) DrawMainScreen() {
	switch e.beginFlag() {
	case 0:
		e.newPage()
		e.clearWindow()
	case 1:
		e.newPage()
		e.LoadPortrait(e.set, int(e.es.picHead), int(e.es.picBody))
		e.DrawPortrait()
	case 2:
		e.newPage()
		e.LoadAnimPic("PIC", 0x1D, false)
	case 4:
		e.newPage()
		e.Redraw()
	case 3:
		e.newPage()
		e.Redraw()
	case 6:
		e.newPage()
		e.LoadAnimPic("PIC", 1, false)
	}
}
