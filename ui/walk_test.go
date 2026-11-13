package ui

import (
	"image"
	"slices"
	"testing"
	"time"

	"gobox/game"
)

// yard is an open map of 4x1 unless locked or walled; a script waits on the
// squares in script, ahead of a step from those in ahead.
type yard struct {
	blocked map[image.Point]game.Edge // the east side of each square
	script  map[image.Point]bool
	ahead   map[image.Point]bool
	size    image.Point
}

func (y yard) Name() string      { return "yard" }
func (y yard) Size() image.Point { return y.size }
func (y yard) Edge(at image.Point, side int) game.Edge {
	if side == 2 {
		return y.blocked[at]
	}
	if side == 6 {
		return y.blocked[at.Add(image.Pt(-1, 0))]
	}
	return game.Edge{}
}
func (y yard) Departure(at image.Point, side int) bool { return y.ahead[at] }
func (y yard) Event(at image.Point) bool               { return y.script[at] }

// walkGame moves the party for real and says what the test tells it to.
type walkGame struct {
	x      game.Exploring
	event  map[image.Point][]game.Event // what the square's event says
	before map[image.Point][]game.Event // what the script ahead says
	onEv   map[image.Point]func(*game.Exploring)
	land   map[image.Point]image.Point // a step to the key lands on the value instead
	sent   []game.Action
}

func (g *walkGame) Version() uint64 { return fresh() }

func (g *walkGame) State() game.State { return g.x }
func (g *walkGame) Do(a game.Action) (game.Outcome, error) {
	g.sent = append(g.sent, a)
	at := image.Pt(g.x.Where.X, g.x.Where.Y)
	switch a := a.(type) {
	case game.Face:
		g.x.Where.Facing = a.Facing
	case game.Arrive:
		to := a.To
		if l, ok := g.land[to]; ok {
			to = l
		}
		g.x.Where.X, g.x.Where.Y = to.X, to.Y
	case game.RunEvent:
		if a.Before {
			return game.Outcome{Events: g.before[at]}, nil
		}
		if f := g.onEv[at]; f != nil {
			f(&g.x)
		}
		return game.Outcome{Events: g.event[at]}, nil
	}
	return game.Outcome{}, nil
}
func (g *walkGame) Save() ([]byte, error) { return nil, nil }
func (g *walkGame) Load([]byte) error     { return nil }

// walkOn walks g to at, a square a Step, to its end; g's party's square after.
func walkOn(t *testing.T, g *walkGame, at image.Point) image.Point {
	t.Helper()
	r := newRules(g, game.Data{})
	clock := time.Unix(1000, 0)
	r.now = func() time.Time { clock = clock.Add(walkPace); return clock }
	if res := r.send(tWalkTo{At: at}); res.Err != nil {
		t.Fatal(res.Err)
	}
	for i := 0; i < 20 && r.walk != nil; i++ {
		r.Step()
	}
	if r.walk != nil {
		t.Fatal("the walk never ended")
	}
	return image.Pt(g.x.Where.X, g.x.Where.Y)
}

func yardGame(y yard) *walkGame {
	if y.size == (image.Point{}) {
		y.size = image.Pt(4, 1)
	}
	g := &walkGame{}
	g.x.Map.Walk = y
	g.x.Where.Facing = 2
	return g
}

func TestWalkToTheSquare(t *testing.T) {
	if at := walkOn(t, yardGame(yard{}), image.Pt(3, 0)); at != image.Pt(3, 0) {
		t.Fatalf("at %v", at)
	}
}

// TestWalkScriptOnUnvisited: a script's text on a square not visited stops
// the walk there.
func TestWalkScriptOnUnvisited(t *testing.T) {
	sq := image.Pt(1, 0)
	g := yardGame(yard{script: map[image.Point]bool{sq: true}})
	g.event = map[image.Point][]game.Event{sq: {game.Narrated{Text: "A voice calls."}}}
	if at := walkOn(t, g, image.Pt(3, 0)); at != sq {
		t.Fatalf("at %v, want the script's square", at)
	}
}

// TestWalkTextOnVisited: on a square walked before plain text is flavour and
// the walk goes on (POR's rule); a picture still stops it.
func TestWalkTextOnVisited(t *testing.T) {
	sq := image.Pt(1, 0)
	g := yardGame(yard{script: map[image.Point]bool{sq: true}})
	g.x.Map.Automap.Walked = []image.Point{sq}
	g.event = map[image.Point][]game.Event{sq: {game.Narrated{Text: "A voice calls."}}}
	if at := walkOn(t, g, image.Pt(3, 0)); at != image.Pt(3, 0) {
		t.Fatalf("at %v, the walk stopped on visited text", at)
	}
	g = yardGame(yard{script: map[image.Point]bool{sq: true}})
	g.x.Map.Automap.Walked = []image.Point{sq}
	g.event = map[image.Point][]game.Event{sq: {game.Shown{}}}
	if at := walkOn(t, g, image.Pt(3, 0)); at != sq {
		t.Fatalf("at %v, a picture must stop it", at)
	}
}

// TestWalkBlankText: blank text only clears the window: no stop anywhere.
func TestWalkBlankText(t *testing.T) {
	sq := image.Pt(1, 0)
	g := yardGame(yard{script: map[image.Point]bool{sq: true}})
	g.event = map[image.Point][]game.Event{sq: {game.Narrated{Text: "  "}}}
	if at := walkOn(t, g, image.Pt(3, 0)); at != image.Pt(3, 0) {
		t.Fatalf("at %v", at)
	}
}

// TestWalkScriptAhead: text ahead of a step on an unvisited square: the step
// it holds is still made, then the walk stops.
func TestWalkScriptAhead(t *testing.T) {
	sq := image.Pt(1, 0)
	g := yardGame(yard{ahead: map[image.Point]bool{{}: true}})
	g.before = map[image.Point][]game.Event{{}: {game.Narrated{Text: "You approach the gate."}}}
	if at := walkOn(t, g, image.Pt(3, 0)); at != sq {
		t.Fatalf("at %v", at)
	}
}

// TestWalkScriptAheadMoves: the script ahead moves the party: no step is made.
func TestWalkScriptAheadMoves(t *testing.T) {
	g := yardGame(yard{ahead: map[image.Point]bool{{}: true}})
	g.Do(game.Face{Facing: 2})
	w := &moveAhead{walkGame: g}
	r := newRules(w, game.Data{})
	r.send(tWalkTo{At: image.Pt(3, 0)})
	r.Step()
	if r.walk != nil || slices.ContainsFunc(g.sent, func(a game.Action) bool { _, ok := a.(game.Arrive); return ok }) {
		t.Fatalf("sent %v", g.sent)
	}
}

type moveAhead struct{ *walkGame }

func (m *moveAhead) Do(a game.Action) (game.Outcome, error) {
	if _, ok := a.(game.RunEvent); ok {
		m.x.Where.Y = 3
	}
	return m.walkGame.Do(a)
}

// TestWalkLanding: a step that lands elsewhere stops the walk.
func TestWalkLanding(t *testing.T) {
	g := yardGame(yard{})
	g.land = map[image.Point]image.Point{{1, 0}: {0, 0}} // the step is refused by the rules' door, say
	if at := walkOn(t, g, image.Pt(3, 0)); at != (image.Point{}) {
		t.Fatalf("at %v", at)
	}
	if n := len(g.sent); n != 1 {
		t.Fatalf("sent %v, want the one step", g.sent)
	}
}

// TestWalkMenu: a question or a door coming up stops it, visited or not.
func TestWalkMenu(t *testing.T) {
	sq := image.Pt(1, 0)
	for _, up := range []func(*game.Exploring){
		func(x *game.Exploring) { x.Ask = &game.Question{} },
		func(x *game.Exploring) { x.Door = &game.Door{} },
	} {
		g := yardGame(yard{script: map[image.Point]bool{sq: true}})
		g.x.Map.Automap.Walked = []image.Point{sq}
		g.onEv = map[image.Point]func(*game.Exploring){sq: up}
		if at := walkOn(t, g, image.Pt(3, 0)); at != sq {
			t.Fatalf("at %v", at)
		}
	}
}

// TestWalkLockedDoor: locked doors count as walls in the search, whichever
// the kind; a plain closed door does not (the game opens it, or refuses).
func TestWalkLockedDoor(t *testing.T) {
	for _, st := range []game.DoorState{game.DoorLocked, game.DoorWizardLocked} {
		g := yardGame(yard{blocked: map[image.Point]game.Edge{{1, 0}: {Kind: game.EdgeDoor, State: st}}})
		r := newRules(g, game.Data{})
		if res := r.send(tWalkTo{At: image.Pt(3, 0)}); res.Err == nil {
			t.Fatalf("state %v: a way through a locked door", st)
		}
	}
	g := yardGame(yard{blocked: map[image.Point]game.Edge{{1, 0}: {Kind: game.EdgeDoor, State: game.DoorClosed}}})
	if at := walkOn(t, g, image.Pt(3, 0)); at != image.Pt(3, 0) {
		t.Fatalf("at %v", at)
	}
	g = yardGame(yard{size: image.Pt(4, 2), blocked: map[image.Point]game.Edge{{1, 0}: {Kind: game.EdgeDoor, State: game.DoorLocked}}})
	if at := walkOn(t, g, image.Pt(2, 0)); at.X != 2 { // round the door by the second row
		t.Fatalf("at %v", at)
	}
}

// TestWalkPace: no square before walkPace is up.
func TestWalkPace(t *testing.T) {
	g := yardGame(yard{})
	r := newRules(g, game.Data{})
	clock := time.Unix(1000, 0)
	r.now = func() time.Time { return clock }
	r.send(tWalkTo{At: image.Pt(3, 0)})
	r.Step()
	r.Step()
	if len(g.sent) != 1 {
		t.Fatalf("sent %v", g.sent)
	}
	clock = clock.Add(walkPace)
	r.Step()
	if len(g.sent) != 2 {
		t.Fatalf("sent %v", g.sent)
	}
}

// TestStopWalk: StopWalk ends the walk with nothing sent; a player's own
// action does too.
func TestStopWalk(t *testing.T) {
	g := yardGame(yard{})
	r := newRules(g, game.Data{})
	r.send(tWalkTo{At: image.Pt(3, 0)})
	r.send(tStopWalk{})
	if p := shownOf(r).Pacing; p.Pace == game.PaceWalk || len(g.sent) != 0 {
		t.Fatalf("pace %v sent %v", p.Pace, g.sent)
	}
	r.send(tWalkTo{At: image.Pt(3, 0)})
	r.send(game.Search{})
	if shownOf(r).Pacing.Pace == game.PaceWalk {
		t.Fatal("a move did not stop it")
	}
}

// TestWalkRoundWall: the way goes round a wall, the party turning first.
func TestWalkRoundWall(t *testing.T) {
	g := yardGame(yard{size: image.Pt(2, 2), blocked: map[image.Point]game.Edge{{}: {Kind: game.EdgeWall}}})
	g.x.Where.Facing = 0
	if at := walkOn(t, g, image.Pt(1, 0)); at != image.Pt(1, 0) {
		t.Fatalf("at %v", at)
	}
	if _, ok := g.sent[0].(game.Face); !ok {
		t.Fatalf("sent %v", g.sent)
	}
}

// TestWalkNoTextStops: with no text at all, what POR's table calls non-quiet
// stops the walk on an unvisited square: a menu put up, a picture put up, a
// picture told as an event. A script that only computes does not.
func TestWalkNoTextStops(t *testing.T) {
	sq := image.Pt(1, 0)
	for name, c := range map[string]struct {
		state func(*game.Exploring)
		ev    []game.Event
		want  image.Point
	}{
		"menu":    {state: func(x *game.Exploring) { x.Ask = &game.Question{Replies: []string{"Yes"}} }, want: sq},
		"picture": {state: func(x *game.Exploring) { x.Picture = &game.Picture{Name: "pic"} }, want: sq},
		"event":   {ev: []game.Event{game.Shown{}}, want: sq},
		"compute": {want: image.Pt(3, 0)},
	} {
		g := yardGame(yard{script: map[image.Point]bool{sq: true}})
		g.event = map[image.Point][]game.Event{sq: c.ev}
		if c.state != nil {
			g.onEv = map[image.Point]func(*game.Exploring){sq: c.state}
		}
		if at := walkOn(t, g, image.Pt(3, 0)); at != c.want {
			t.Errorf("%s: at %v, want %v", name, at, c.want)
		}
	}
}
