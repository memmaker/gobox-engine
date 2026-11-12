package pool

import (
	"image"
	"rules1e"
	"testing"

	"gobox/game/play"
)

// Sidestep{4} after a step forward walks back to the square, the facing kept.
func TestSidestepBack(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	for range 4 {
		x, y, f := e.posX, e.posY, e.facing
		stepBy(e, 0)
		if e.posX == x && e.posY == y {
			turnRight(e)
			continue
		}
		if err := stepBy(e, 4); err != nil {
			t.Fatal(err)
		}
		if e.posX != x || e.posY != y || e.facing != f {
			t.Fatalf("back: at %d,%d facing %d, want %d,%d facing %d", e.posX, e.posY, e.facing, x, y, f)
		}
		return
	}
	t.Skip("no open square ahead")
}

// Sidestep{6} then {2} walk aside and back; no step turns the party, and the
// published facing never changes.
func TestSidestepAside(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	for range 4 {
		x, y, f := e.posX, e.posY, e.facing
		stepBy(e, 6)
		if e.facing != f || commonOf(e.State()).Where.Facing != f/2&3*2 {
			t.Fatalf("aside: facing %d (published %d), want %d", e.facing, commonOf(e.State()).Where.Facing, f)
		}
		if e.posX == x && e.posY == y {
			turnRight(e)
			continue
		}
		stepBy(e, 2)
		if e.posX != x || e.posY != y || e.facing != f {
			t.Fatalf("aside and back: at %d,%d facing %d, want %d,%d facing %d", e.posX, e.posY, e.facing, x, y, f)
		}
		return
	}
	t.Skip("no open square aside")
}

// stepBy is gobox's step by (0 ahead, 4 back, 6 left, 2 right of the facing):
// RunEvent{Before}, then an Arrive.
func stepBy(e *Engine, by int) error {
	d := (e.facing + by) % 8
	from := image.Pt(e.posX, e.posY)
	if _, err := e.Do(play.RunEvent{Before: true, Side: d}); err != nil {
		return err
	}
	_, err := e.Do(play.Arrive{From: from, To: from.Add(image.Pt(rules1e.DirDX[d], rules1e.DirDY[d])), Facing: e.facing})
	return err
}

// turnRight is gobox's turn to the right: a Face.
func turnRight(e *Engine) { e.Do(play.Face{Facing: (e.facing + 2) % 8}) }

// Arrive moves and leaves the step's events owed (Map.Event); RunEvent runs them.
func TestArriveThenRunEvent(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	for range 4 {
		x, y := e.posX, e.posY
		stepBy(e, 0)
		if e.posX == x && e.posY == y {
			e.ext.owed = nil
			turnRight(e)
			continue
		}
		if e.ext.pre != nil {
			t.Fatal("Arrive left the pre-step script owed")
		}
		at := image.Pt(e.posX, e.posY)
		xs, ok := e.State().(play.Exploring)
		if e.ext.owed == nil || !ok || xs.Map.Walk == nil || !xs.Map.Walk.Event(at) {
			t.Fatalf("after Arrive: owed %v, exploring %v", e.ext.owed, ok)
		}
		if _, err := e.Do(play.RunEvent{}); err != nil {
			t.Fatal(err)
		}
		if e.ext.owed != nil {
			t.Fatal("RunEvent left the step's events owed")
		}
		return
	}
	t.Skip("no open square ahead")
}

// An Arrive no RunEvent{Before} came ahead of is refused (it would run the pre-step script).
func TestArriveNeedsBefore(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	from := image.Pt(e.posX, e.posY)
	_, err := e.Do(play.Arrive{From: from, To: from.Add(image.Pt(rules1e.DirDX[e.facing], rules1e.DirDY[e.facing])), Facing: e.facing})
	if err == nil || e.posX != from.X || e.posY != from.Y {
		t.Fatalf("Arrive alone: err %v, at %d,%d", err, e.posX, e.posY)
	}
}
