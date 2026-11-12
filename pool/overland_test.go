package pool

import (
	"image"
	"testing"

	"gobox/game/play"
)

// TestOverlandIsTheMap: on the overland the state says so (ExploreMap.Overland),
// so gobox shows the map screen, not the page's small map as a picture window,
// and walks it each way (ActGo); a step east lands east whatever the facing,
// and its events wait on the overland square gobox asks about.
func TestOverlandIsTheMap(t *testing.T) {
	e := testEngine(t)
	e.Explore = true
	e.DevAction = e.DevScene("overland")
	e.Init()
	e.Start(e.Play)
	if !playOn(e, 50, func(s play.State) bool { x, ok := s.(play.Exploring); return ok && x.Ask == nil && x.Map.Walk != nil }) {
		t.Fatal("not exploring the overland")
	}
	x := e.State().(play.Exploring)
	if !x.Map.Overland {
		t.Fatal("overland: ExploreMap.Overland false")
	}
	from := image.Pt(x.Where.X, x.Where.Y)
	if _, err := e.Do(play.RunEvent{Before: true, Side: 2}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Do(play.Arrive{From: from, To: from.Add(image.Pt(1, 0)), Facing: x.Where.Facing}); err != nil {
		t.Fatal(err)
	}
	w := commonOf(e.State()).Where
	if w.X != from.X+1 || w.Y != from.Y {
		t.Errorf("step east from %v (facing %d): at %d,%d", from, x.Where.Facing, w.X, w.Y)
	}
	// The step's events wait for RunEvent: Map.Event says so on the party's
	// overland square (not on PosX, PosY, the 3D view's), or gobox never
	// runs them and the overland's encounters and places never happen.
	if !e.stepOwed() || !e.walkMap().Event(image.Pt(w.X, w.Y)) {
		t.Errorf("after the step: owed %v, Map.Event(%d,%d) false", e.stepOwed(), w.X, w.Y)
	}
}
