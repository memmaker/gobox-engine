package present

import (
	"image"

	"gobox/front"
)

// classicGame is the classic front with what only the presenter knows: its
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

func (g classicGame) View() front.View {
	v := g.Game.View()
	if ownMus != nil {
		ownMus.Want(musicChain(g.Game.MusicCue()))
	}
	if playSeq != "" && v.Sequence == nil { // a game.MenuItem's animation, gobox's own: the game's state never knows
		if v = Sequence(v, playSeq, sequences, filesDir); v.Sequence == nil {
			playSeq = "" // no such sequence
		}
	}
	return v
}

// pickerView is page page of the grid picker the state holds, else the front's.
func (g classicGame) pickerView(page int) *front.Picker {
	return g.Game.PickerOn(page, g.p.grid[0], g.p.grid[1])
}

// IconEditorView is the icon designer on the presenter's page.
func (g classicGame) IconEditorView() *front.IconEditor {
	ink := g.Game.View().Ink
	return g.Game.IconEditorOn(&ink, g.p.grid[0], g.p.grid[1])
}

// CreationView is the sheet on the page the presenter last laid out (ActResize).
func (g classicGame) CreationView() *front.Creation {
	return g.Game.CreationOn(g.p.grid[0], g.p.grid[1])
}

// FullMapView is the full map under cursor.
func (g classicGame) FullMapView(cursor image.Point, label *string) *front.FullMap {
	return g.Game.FullMapOf(g.Game.FullMapView(cursor, label))
}

// WithRules is g as classicGame (p lays out its page) when it is the classic front.
func WithRules(g front.Game, p *Presenter) front.Game {
	if c, ok := g.(*Game); ok {
		return classicGame{c, p}
	}
	return g
}
