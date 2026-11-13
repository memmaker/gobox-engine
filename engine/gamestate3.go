package engine

import (
	"gobox/game"
)

// state3 fills the State fields of the title, the after-combat page, the
// journal text, the page's size and colours and the pacing from the engine.
func (e *Engine) state3(s *facts) {
	if e.ext.intro {
		s.Intro = &game.Intro{} // the title is up: State's Scene is the adventure's Start (SetStart)
	}
	s.Sequence = e.sequence
	s.Read = e.ext.read // the engine's own: the same pointer while the text is up
	r := e.snap.results
	if r != nil {
		g := &game.Results{Shares: r.Shares, Outcome: r.Outcome, Exp: r.Exp, Taken: r.Taken, End: r.End}
		for _, h := range r.Harmed {
			g.Harmed = append(g.Harmed, game.Harm{Name: h.Name, What: h.What})
		}
		s.Results = g
	}
}
