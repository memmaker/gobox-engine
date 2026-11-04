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
	v.AreaMap = classic.AreaMap(v.AreaMap, s, gameArt)
	return v
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
