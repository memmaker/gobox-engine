package classic

import (
	"image"
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// wallMap: every side is a wall, and a script waits ahead of every step.
type wallMap struct{}

func (wallMap) Name() string                    { return "" }
func (wallMap) Size() image.Point               { return image.Pt(4, 4) }
func (wallMap) Edge(image.Point, int) game.Edge { return game.Edge{Kind: game.EdgeWall} }
func (wallMap) Departure(image.Point, int) bool { return true }
func (wallMap) Event(image.Point) bool          { return false }

// TestWallStepRunsScriptAhead: facing a wall, the square's script ahead of
// the step still runs (it may move the party through, as SOTSB's town gate
// does); only the Arrive is held back.
func TestWallStepRunsScriptAhead(t *testing.T) {
	r := &fake{st: play.Exploring{Map: play.ExploreMap{Walk: wallMap{}}}}
	g := NewGame(r, game.Data{})
	if err := g.walk(front.ActForward); err != nil {
		t.Fatal(err)
	}
	if len(r.sent) != 1 || r.sent[0] != (play.RunEvent{Before: true, Side: 0}) {
		t.Fatalf("sent %v", r.sent)
	}
}
