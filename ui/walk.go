package ui

import (
	"image"

	"gobox/game"
	"gobox/game/play"
)

// walks are the moves gobox makes on the game's Map: a step's side
// relative to the facing (forward 0, right 2, back 4, left 6), or a turn.
var walks = map[ActKind]struct {
	side int
	turn bool
}{
	ActForward: {0, false}, ActStrafeRight: {2, false}, ActBackward: {4, false}, ActStrafeLeft: {6, false},
	ActTurnRight: {2, true}, ActTurnLeft: {6, true},
}

// steps are a step's offset by side (0 N, 2 E, 4 S, 6 W).
var steps = map[int]image.Point{0: {0, -1}, 2: {1, 0}, 4: {0, 1}, 6: {-1, 0}}

// walkMap is the map the party walks now, nil none.
func (g *Game) walkMap() game.Map {
	return g.r.Map()
}

// at is the party's square on the walked map: Where.X, Y (on the overland
// too, where Where.Square is only the 3D view's square).
func at(w game.Where) image.Point { return image.Pt(w.X, w.Y) }

// walk makes move k on the game's Map: a turn is Face; a step through an
// open side or a door is Arrive, a wall stops it after any script ahead. Around
// the step, RunEvent: Before when Map.Departure says a script runs ahead
// of it, after when Map.Event says one waits where the party stands.
func (g *Game) walk(k ActKind) error {
	w := walks[k]
	m := g.walkMap()
	if m == nil {
		return nil // nowhere to walk now
	}
	where := g.Where()
	if w.turn {
		return g.do(play.Face{Facing: (where.Facing + w.side) % 8})
	}
	return g.stepThrough(m, (where.Facing+w.side)%8)
}

// goWay is ActGo on the area map: a step towards dir (0 N .. 3 W) whatever
// the facing.
func (g *Game) goWay(dir int) error {
	m := g.walkMap()
	if m == nil || dir < 0 || dir > 3 {
		return nil
	}
	return g.stepThrough(m, 2*dir)
}

// stepThrough moves the party on m through side (0 N, 2 E, 4 S, 6 W) of its square.
func (g *Game) stepThrough(m game.Map, side int) error {
	where := g.Where()
	from := at(where)
	// The script ahead of the step runs even facing a wall, as the
	// originals run the square's pre-step script on every step (SOTSB
	// O050:3DB9): it may move the party through that wall.
	if m.Departure(from, side) {
		if err := g.do(play.RunEvent{Before: true, Side: side}); err != nil || g.r.refusalAsk != nil {
			return err
		}
		now := g.Where()
		if m = g.walkMap(); m == nil || at(now) != from || now.Facing != where.Facing || now.Area != where.Area {
			return nil
		}
	}
	if m.Edge(from, side).Kind == game.EdgeWall {
		return nil
	}
	if err := g.do(play.Arrive{From: from, To: from.Add(steps[side]), Facing: where.Facing}); err != nil || g.r.refusalAsk != nil {
		return err
	}
	if m := g.walkMap(); m != nil && m.Event(at(g.Where())) { // the square's event: run it
		return g.do(play.RunEvent{})
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
