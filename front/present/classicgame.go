package present

import (
	"gobox/front"
	"gobox/front/classic"
	"gobox/game"
)

// classicGame is a game that also gives its rules as a game.Rules: the
// screens front/classic builds from its State stand in for the game's own
// (classic.View, classic.Creation); everything else is the game's.
type classicGame struct {
	front.Game
	r game.Rules
	p *Presenter
}

func (g classicGame) View() front.View { return classic.View(g.Game.View(), g.r.State()) }

// CreationView is the sheet on the page the presenter last laid out (ActResize).
func (g classicGame) CreationView() *front.Creation {
	return classic.Creation(g.r.State(), g.p.grid[0], g.p.grid[1])
}

// WithRules is g as classicGame (p lays out its page) when it gives its rules (Rules()).
func WithRules(g front.Game, p *Presenter) front.Game {
	if r, ok := g.(interface{ Rules() game.Rules }); ok {
		return classicGame{g, r.Rules(), p}
	}
	return g
}
