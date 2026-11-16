package engine

import (
	"reflect"
	"testing"

	"gobox/game"
	"gobox/rules"
)

// The rocket's tests mirror buckrogers-go TestRocketStep (the oracle) against buckroger-re explore.pas MoveRocket.

// rocketMoveText is the test block's move entry: it copies the ship's column and the first and fifth bodies' squares into cells the
// test reads (Word4BA5, Word4BAB, Word4BAA are scripts' words with no other meaning in a block of the test's own).
const rocketMoveText = "Word4BA5 = RocketX\n    Word4BAB = BodySquares[0]\n    Word4BAA = BodySquares[9]"

// br1Flying is a begun BR1 game flying the rocket in a block of the test's own (view kind 2), the menu up.
func br1Flying(t *testing.T) *Engine {
	t.Helper()
	e := br1Begun(t)
	br1OwnBlockMove(t, e, 201, "AreaKind(2, 255)", rocketMoveText)
	return e
}

// rocketAsk is the question the flying party waits on.
func rocketAsk(t *testing.T, e *Engine) *game.Question {
	t.Helper()
	q := commonOf(e.build()).Ask
	if q == nil {
		t.Fatalf("no question on %T", e.build())
	}
	return q
}

// rocketMove makes a move of the front's heading dir from the menu.
func rocketMove(t *testing.T, e *Engine, dir int) {
	t.Helper()
	if q := rocketAsk(t, e); !q.Heading {
		if len(q.Replies) != 1 || q.Replies[0] != "Move" {
			t.Fatalf("the menu %+v", q)
		}
		if _, err := e.doNow(game.Reply{N: 0}); err != nil {
			t.Fatal(err)
		}
	}
	sc, ok := e.build().(game.Scene)
	if !ok || sc.Kind != game.SceneHeading || len(sc.Text) != 1 || sc.Text[0] != "En route. Fuel = "+rocketFuel(t, e)+" " {
		t.Fatalf("%T: want a heading scene with the fuel: %+v", e.build(), sc)
	}
	if q := rocketAsk(t, e); !q.Heading || len(q.Replies) != 1 || q.Replies[0] != "Done" {
		t.Fatalf("the heading question %+v", q)
	}
	if _, err := e.doNow(game.Heading{Dir: dir}); err != nil {
		t.Fatal(err)
	}
}

func rocketFuel(t *testing.T, e *Engine) string {
	t.Helper()
	return itoa(int64(e.fact(rules.NoChar, shipCells["ShipFuel"]).(rules.Number)))
}

// TestBR1RocketStart: the flight starts on orbit 2's body (9,8), heading 87's north, the counters clear (gameinit.pas
// InitShipAndOrbit); the front gets the system map as an overland with the ship as the party, and a menu of Move.
func TestBR1RocketStart(t *testing.T) {
	e := br1Flying(t)
	w := e.world
	if w.Flying != 1 || w.In3D != 0 || w.Scripted != 0 || w.RocketX != 9 || w.RocketY != 8 || w.RocketHeading != 0 || w.OrbitTick != 0 || w.OrbitStep != 0 {
		t.Fatalf("world %+v", w)
	}
	c := commonOf(e.build()) // a question is up: the Scene carries the map and where the party is
	am, where := c.Automap, c.Where
	if am.Size.W != 22 || am.Size.H != 22 || len(am.Terrain) != 22*22 || am.Marker != "ship" || where.Mode != "space" || where.X != 9 || where.Y != 8 || where.Facing != 0 {
		t.Errorf("map %+v where %+v", am.Size, where)
	}
	sun, bodies := 0, 0
	for _, tile := range am.Terrain {
		if tile == 296 {
			sun++
		}
		if tile != 0 {
			bodies++
		}
	}
	if am.Terrain[10*22+11] != 296 || sun < 1 || bodies != 14 {
		t.Errorf("sun %d, %d tiles drawn (the sun and 13 bodies)", sun, bodies)
	}
	if q := rocketAsk(t, e); q.Heading || len(q.Replies) != 1 || q.Replies[0] != "Move" {
		t.Errorf("menu %+v", q)
	}
}

// TestBR1RocketStep mirrors TestRocketStep: a move east into the sun is undone, the counter rises by 3 (30 to 33, wrapped to 1), the
// orbit steps down from 0 to 223, three days pass (a hit point a day) and the bodies stand where step 223 puts them, which the script
// reads through BodySquares.
func TestBR1RocketStep(t *testing.T) {
	e := br1Flying(t)
	e.world.RocketX, e.world.RocketY, e.world.OrbitTick = 10, 10, 30
	hp := e.selFact(rules.FactHP)
	e.setSelFact(rules.FactHP, hp-10)
	rocketMove(t, e, 2) // east
	w := e.world
	if w.RocketX != 10 || w.RocketY != 10 || w.RocketHeading != 6 || w.OrbitTick != 1 || w.OrbitStep != 223 {
		t.Fatalf("world %+v", w)
	}
	if got := e.selFact(rules.FactHP); got != hp-10+3 {
		t.Errorf("hit points %d, want %d: three days rested", got, hp-7)
	}
	s := e.spaceMap()
	if b := s.Body(0, 223).At; int(br1Cell(t, e, "Word4BAB")) != b.X {
		t.Errorf("the script reads body 0 at x %d, want %d", br1Cell(t, e, "Word4BAB"), b.X)
	}
	if b := s.Body(4, 223).At; int(br1Cell(t, e, "Word4BAA")) != b.Y {
		t.Errorf("the script reads body 4's y %d, want %d", br1Cell(t, e, "Word4BAA"), b.Y)
	}
	if int(br1Cell(t, e, "Word4BA5")) != 10 {
		t.Errorf("the script reads the ship's column %d", br1Cell(t, e, "Word4BA5"))
	}
	if !rocketAsk(t, e).Heading {
		t.Errorf("after a move the heading is asked again (the Pascal's move mode)")
	}
	if _, err := e.doNow(game.Heading{Dir: 0}); err != nil { // north, from the sun's edge
		t.Fatal(err)
	}
	if e.world.RocketY != 9 || e.world.RocketHeading != 0 {
		t.Errorf("north: %d,%d heading %d", e.world.RocketX, e.world.RocketY, e.world.RocketHeading)
	}
	if _, err := e.doNow(game.Leave{}); err != nil { // Done: back to the menu
		t.Fatal(err)
	}
	if q := rocketAsk(t, e); q.Heading || len(q.Replies) != 1 {
		t.Errorf("Done leaves the move mode: %+v", q)
	}
}

// TestBR1RocketEdges: the ship stops at the map's edge (column and row 0..21), the clock still runs.
func TestBR1RocketEdges(t *testing.T) {
	e := br1Flying(t)
	e.world.RocketX, e.world.RocketY = 0, 21
	rocketMove(t, e, 6) // west
	rocketMove(t, e, 4) // south
	if e.world.RocketX != 0 || e.world.RocketY != 21 || e.world.OrbitTick != 6 {
		t.Errorf("%d,%d tick %d", e.world.RocketX, e.world.RocketY, e.world.OrbitTick)
	}
	e.world.RocketX, e.world.RocketY = 21, 0
	rocketMove(t, e, 2)
	rocketMove(t, e, 0)
	if e.world.RocketX != 21 || e.world.RocketY != 0 {
		t.Errorf("%d,%d", e.world.RocketX, e.world.RocketY)
	}
}

// TestBR1RocketOrbitCycle: the orbit step wraps 0 to 223 and a full cycle returns it (explore.pas PHASE_CYCLE).
func TestBR1RocketOrbitCycle(t *testing.T) {
	e := br1Flying(t)
	e.world.OrbitStep, e.world.OrbitTick = 1, 31
	rocketMove(t, e, 6)
	if e.world.OrbitStep != 0 || e.world.OrbitTick != 2 {
		t.Errorf("step %d tick %d", e.world.OrbitStep, e.world.OrbitTick)
	}
}

// TestBR1RocketSaveLoad: a game saved while flying loads back flying, at the same square, heading, counters and map, and asks the
// menu again.
func TestBR1RocketSaveLoad(t *testing.T) {
	a := br1Flying(t)
	rocketMove(t, a, 4) // south
	rocketMove(t, a, 2) // east
	blob, err := a.Save()
	if err != nil {
		t.Fatal(err)
	}
	b := br1Loaded(t, blob)
	a.world.AreaKind = b.world.AreaKind // the running block, which a save records
	if a.world != b.world {
		t.Errorf("world %+v, loaded %+v", a.world, b.world)
	}
	ca, cb := commonOf(a.build()), commonOf(b.build())
	if !reflect.DeepEqual(ca.Automap, cb.Automap) || ca.Where.Mode != "space" || ca.Where.X != cb.Where.X || ca.Where.Facing != cb.Where.Facing {
		t.Errorf("where %+v, loaded %+v", ca.Where, cb.Where)
	}
	if q := rocketAsk(t, b); q.Heading || len(q.Replies) != 1 {
		t.Errorf("loaded into %+v, want the menu", q)
	}
}
