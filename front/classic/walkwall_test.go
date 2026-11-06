package classic

import (
	"image"
	"testing"

	"gobox/front"
	"gobox/game"
)

// wallMap: every side is a wall, and a script waits ahead of every step.
type wallMap struct{}

func (wallMap) Name() string                    { return "" }
func (wallMap) Size() image.Point               { return image.Pt(4, 4) }
func (wallMap) Edge(image.Point, int) game.Edge { return game.Edge{Kind: game.EdgeWall} }
func (wallMap) Departure(image.Point, int) bool { return true }
func (wallMap) Event(image.Point) bool          { return false }

type wallRules struct {
	game.Rules
	sent []game.Decision
}

func (r *wallRules) State() game.State { return game.State{} }
func (r *wallRules) Map() game.Map     { return wallMap{} }
func (r *wallRules) Apply(d game.Decision) game.Result {
	r.sent = append(r.sent, d)
	return game.Result{}
}

// TestWallStepRunsScriptAhead: facing a wall, the square's script ahead of
// the step still runs (it may move the party through, as SOTSB's town gate
// does); only the Arrive is held back.
func TestWallStepRunsScriptAhead(t *testing.T) {
	r := &wallRules{}
	g := &Game{R: r}
	if err := g.walk(front.ActForward); err != nil {
		t.Fatal(err)
	}
	if len(r.sent) != 1 || r.sent[0] != (game.RunEvent{Before: true, Side: 0}) {
		t.Fatalf("sent %v", r.sent)
	}
}
