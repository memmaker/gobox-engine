package present

import (
	"image"
	"testing"

	"gobox/front"
	"gobox/front/classic"
	"gobox/game"
	"gobox/gpu"
)

// endRules is a fight one Quick ends: the Result carries its blows and the
// state after it is a page (no Combat).
type endRules struct{ over bool }

func (r *endRules) State() game.State {
	if r.over {
		return game.State{Mode: game.ModeStory, Page: &game.Page{Prompt: []game.TextRun{{Text: "DONE"}}}}
	}
	return game.State{Mode: game.ModeCombat, Combat: &game.Combat{Size: image.Pt(4, 4),
		Units: []game.Unit{{ID: 1, At: image.Pt(1, 1)}, {ID: 2, At: image.Pt(2, 1), Foe: true}}}}
}
func (r *endRules) Apply(d game.Decision) game.Result {
	if _, ok := d.(game.Quick); !ok {
		return game.Result{}
	}
	r.over = true
	var evs []game.Event
	for range 5 {
		evs = append(evs, game.Event{Kind: game.EvAttack, Unit: 1, Target: 2, Pose: 1, Ms: 300})
	}
	return game.Result{Events: append(evs, game.Event{Kind: game.EvDown, Unit: 2})}
}
func (r *endRules) Pending() *game.Ask { return nil }
func (r *endRules) Step() game.Result  { return game.Result{} }

// TestEndingEventsPlay drives the front's loop (Act, Draw, Tick at 60 a
// second): the blows of a Result that ends the fight play on the fight,
// 1.5 s of them, and then the page comes up; it never hangs.
func TestEndingEventsPlay(t *testing.T) {
	defer func(s bool, n int64) { Scripted, frames = s, n }(Scripted, frames)
	r := &endRules{}
	f := &Front{Game: classic.NewGame(r, game.Data{}), started: true}
	f.Glyph, f.Font = func(byte) [8]byte { return [8]byte{} }, func(byte, bool) [8]byte { return [8]byte{} }
	f.P.SetEngine(f.Game)
	dst := gpu.NewImage(640, 400)
	f.Act()
	f.Draw(dst)
	f.Tick()
	f.Game.Answer(front.Action{Kind: front.ActQuick})
	f.P.Play(f.Game.Settle().Events)
	fight, page := 0, -1
	for i := range 600 {
		f.Act()
		f.Draw(dst)
		f.Tick()
		if f.P.Playing() && f.P.fight {
			fight++
		}
		if f.P.Playing() && f.Idle() {
			t.Fatalf("frame %d: idle while the blows play (a script would end on them)", i)
		}
		if !f.P.Playing() && !f.P.fight && page < 0 {
			page = i
		}
	}
	if fight < 80 || page < 80 || page > 120 {
		t.Fatalf("%d frames of the fight while its blows play, the page from frame %d (want about 90 of each)", fight, page)
	}
}
