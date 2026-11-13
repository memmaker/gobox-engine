package ui

import (
	"image"
	"image/color"
)

// *classicGame is the classic front with what only the presenter knows: its
// page's size, the music and a menu item's animation.
type classicGame struct {
	*Game
	p *Presenter
}

// playSeq is the game.MenuItem sequence playing ("": none); its end comes back here, not to the game.
var playSeq string

// SequenceDone ends a menu item's animation here; the game's own go on to it.
func (g classicGame) SequenceDone(name string) {
	if playSeq != "" && name == playSeq {
		playSeq = ""
		return
	}
	g.Game.SequenceDone(name)
}

// pickerView is page page of the grid picker the state holds, else the front's.
func (g classicGame) pickerView(page int) *FrontPicker {
	return g.Game.PickerOn(page, g.r.grid.X, g.r.grid.Y)
}

// IconEditorView is the icon designer on the presenter's page.
func (g classicGame) IconEditorView(ink [16]color.RGBA) *FrontIconEditor {
	return g.Game.IconEditorOn(&ink, g.r.grid.X, g.r.grid.Y)
}

// CreationView is the sheet on the page the presenter last laid out (ActResize).
func (g classicGame) CreationView() *FrontCreation {
	return g.Game.CreationOn(g.r.grid.X, g.r.grid.Y)
}

// FullMapView is the full map under cursor.
func (g classicGame) FullMapView(cursor image.Point, label *string) *FrontFullMap {
	return g.Game.FullMapOf(g.Game.FullMapView(cursor, label))
}
