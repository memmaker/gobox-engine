package present

import (
	"image"

	"gobox/front"
	"gobox/front/classic"
	"gobox/game"
)

// classicGame is a game that also gives its rules as a game.Rules: the
// screens front/classic builds from its State stand in for the game's own
// (classic.View, classic.Creation, classic.Training); everything else is the game's.
type classicGame struct {
	front.Game
	r game.Rules
	p *Presenter
}

func (g classicGame) View() front.View {
	s := g.r.State()
	v := classic.View(g.Game.View(), s)
	if v.AreaMap != nil {
		v.AreaMap = classic.AreaMap(v.AreaMap, s, g.art())
	}
	if v.MainMenu != nil && !v.Hall && v.Powers == nil { // the title under the party menu
		v.Logo, v.Backdrop = classic.Title(v.Logo, v.Backdrop, g.art())
	}
	v.Battle = classic.Battle(v.Battle, s)
	v = classic.Page(v, s)
	v = classic.Flags(v, s)
	v = classic.Place(v, s)
	v.Quests, v.Lore = classic.Quests(s), classic.Lore(s)
	v = classic.Waits(v, s)
	v.Party, v.Purse = classic.Party(s, g.art()), classic.Purse(s)
	v = classic.Sheet(v, s)
	v = classic.Border(v, g.art())
	v = classic.Walls(v, s, g.art())
	v = classic.Intro(v, s)
	v = classic.Note(v, s)
	v = classic.Results(v, s)
	v = classic.Screen(v, s)
	v = classic.Read(v, s)
	v = classic.Pacing(v, s)
	return v
}

// art is the game's art as loaded now: it comes after NewFront, and the art set can change.
func (g classicGame) art() map[string]game.Picture {
	if d, ok := g.Game.(interface{ Data() game.Data }); ok {
		return d.Data().Art
	}
	return gameArt
}

// PickerView is page page of the grid picker State holds (the game's where it holds none).
func (g classicGame) pickerView(page int) *front.Picker {
	if p := classic.Picker(g.r.State(), page, g.p.grid[0], g.p.grid[1]); p != nil {
		return p
	}
	return g.Game.PickerView(page)
}

// IconEditorView is the icon designer State holds (the game's where it holds none).
func (g classicGame) IconEditorView() *front.IconEditor {
	ink := g.Game.View().Ink
	if e := classic.IconEditor(g.r.State(), &ink, g.p.grid[0], g.p.grid[1]); e != nil {
		return e
	}
	return g.Game.IconEditorView()
}

// CreationView is the sheet on the page the presenter last laid out (ActResize).
func (g classicGame) CreationView() *front.Creation {
	return classic.Creation(g.r.State(), g.p.grid[0], g.p.grid[1])
}

// FullMapView is the game's full map with the parts classic.FullMap takes from State.
func (g classicGame) FullMapView(cursor image.Point, label *string) *front.FullMap {
	return classic.FullMap(g.Game.FullMapView(cursor, label), g.r.State())
}

// TrainingView is the hall's offer classic.Training builds from State.
func (g classicGame) TrainingView() *front.Training { return classic.Training(g.r.State()) }

// WithRules is g as classicGame (p lays out its page) when it gives its rules (Rules()).
func WithRules(g front.Game, p *Presenter) front.Game {
	if r, ok := g.(interface{ Rules() game.Rules }); ok {
		return classicGame{g, r.Rules(), p}
	}
	return g
}
