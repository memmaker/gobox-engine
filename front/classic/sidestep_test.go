package classic

import (
	"testing"

	"gobox/game"
)

// walkRules has Forward and TurnBy only: it refuses Sidestep.
type walkRules struct{ w game.Where }

func (r *walkRules) State() game.State {
	return game.State{Mode: game.ModeExplore, Where: r.w}
}
func (r *walkRules) Apply(d game.Decision) game.Result {
	switch d := d.(type) {
	case game.Forward:
		r.w.X += [8]int{0, 0, 1, 0, 0, 0, -1, 0}[r.w.Facing]
		r.w.Y += [8]int{-1, 0, 0, 0, 1, 0, 0, 0}[r.w.Facing]
	case game.TurnBy:
		r.w.Facing = (r.w.Facing + d.By) % 8
	default:
		return game.Result{Err: &game.Refusal{Text: "no"}}
	}
	return game.Result{}
}
func (r *walkRules) Pending() *game.Ask { return nil }
func (r *walkRules) Step() game.Result  { return game.Result{} }

// Rules without a Sidestep of their own still step back and aside, the facing kept.
func TestSidestepFallback(t *testing.T) {
	r := &walkRules{w: game.Where{X: 5, Y: 5}}
	g := NewGame(r, game.Data{})
	for _, c := range []struct{ by, x, y int }{{4, 5, 6}, {6, 4, 6}, {2, 5, 6}} {
		if err := g.do(game.Sidestep{By: c.by}); err != nil {
			t.Fatal(err)
		}
		if r.w.X != c.x || r.w.Y != c.y || r.w.Facing != 0 {
			t.Fatalf("by %d: at %d,%d facing %d, want %d,%d facing 0", c.by, r.w.X, r.w.Y, r.w.Facing, c.x, c.y)
		}
	}
}
