package classic

import (
	"image"

	"gobox/front"
	"gobox/game"
)

// walks are the moves gobox makes on the game's Map: a step's side
// relative to the facing (forward 0, right 2, back 4, left 6), or a turn.
var walks = map[front.ActKind]struct {
	side int
	turn bool
}{
	front.ActForward: {0, false}, front.ActStrafeRight: {2, false}, front.ActBackward: {4, false}, front.ActStrafeLeft: {6, false},
	front.ActTurnRight: {2, true}, front.ActTurnLeft: {6, true},
}

// steps are a step's offset by side (0 N, 2 E, 4 S, 6 W).
var steps = map[int]image.Point{0: {0, -1}, 2: {1, 0}, 4: {0, 1}, 6: {-1, 0}}

// walkMap is the map the party walks now, nil none.
func (g *Game) walkMap() game.Map {
	if r, ok := g.R.(game.Mapped); ok {
		return r.Map()
	}
	return nil
}

// at is the party's square on the walked map: Where.X, Y (on the overland
// too, where Where.Square is only the 3D view's square).
func at(w game.Where) image.Point { return image.Pt(w.X, w.Y) }

// walk makes move k on the game's Map: a turn is Face; a step through an
// open side or a door is Arrive, a wall stops it after any script ahead. Around
// the step, RunEvent: Before when Map.Departure says a script runs ahead
// of it, after when Map.Event says one waits where the party stands.
func (g *Game) walk(k front.ActKind) error {
	w := walks[k]
	m := g.walkMap()
	if m == nil {
		return nil // nowhere to walk now
	}
	where := g.R.State().Where
	if w.turn {
		return g.do(game.Face{Facing: (where.Facing + w.side) % 8})
	}
	from, side := at(where), (where.Facing+w.side)%8
	// The script ahead of the step runs even facing a wall, as the
	// originals run the square's pre-step script on every step (SOTSB
	// O050:3DB9): it may move the party through that wall.
	if m.Departure(from, side) {
		if err := g.do(game.RunEvent{Before: true, Side: side}); err != nil || g.asked != nil {
			return err
		}
		now := g.R.State().Where
		if m = g.walkMap(); m == nil || at(now) != from || now.Facing != where.Facing || now.Area != where.Area {
			return nil
		}
	}
	if m.Edge(from, side).Kind == game.EdgeWall {
		return nil
	}
	if err := g.do(game.Arrive{From: from, To: from.Add(steps[side]), Facing: where.Facing}); err != nil || g.asked != nil {
		return err
	}
	if m := g.walkMap(); m != nil && m.Event(at(g.R.State().Where)) { // the square's event: run it
		return g.do(game.RunEvent{})
	}
	return nil
}

// openSides is the party's square's sides (N E S W) a step goes through.
func openSides(m game.Map, w game.Where) (o [4]bool) {
	for i := range o {
		o[i] = m.Edge(at(w), 2*i).Kind != game.EdgeWall
	}
	return o
}
