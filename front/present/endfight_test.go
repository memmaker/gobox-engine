package present

import (
	"image"
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
	"gobox/gpu"
)

// endGame is a fight one Quick ends: the Outcome carries its blows and the
// state after it is the result page.
type endGame struct{ over bool }

func (r *endGame) State() play.State {
	if r.over {
		return play.Scene{Common: play.Common{Status: "DONE"}}
	}
	return play.Fighting{Common: play.Common{Allowed: play.Allowed{Plain: []play.Action{play.Quick{}}}}, Field: play.Battlefield{Size: image.Pt(4, 4),
		Units: []game.Unit{{ID: 1, Member: -1, At: image.Pt(1, 1)}, {ID: 2, Member: -1, At: image.Pt(2, 1), Foe: true}}}}
}
func (r *endGame) Do(a play.Action) (play.Outcome, error) {
	if _, ok := a.(play.Quick); !ok {
		return play.Outcome{}, nil
	}
	r.over = true
	var evs []play.Event
	for range 5 {
		evs = append(evs, play.Attacked{From: 1, To: 2})
	}
	return play.Outcome{Events: append(evs, play.Fell{Unit: 2})}, nil
}
func (r *endGame) Save() ([]byte, error) { return nil, nil }
func (r *endGame) Load([]byte) error     { return nil }

// TestEndingEventsPlay drives the front's loop (Act, Draw, Tick at 60 a
// second): the blows of an Outcome that ends the fight play on the fight,
// about 2 s of them at the default speed, and then the page comes up; it never hangs.
func TestEndingEventsPlay(t *testing.T) {
	defer func(s bool, n int64) { Scripted, frames = s, n }(Scripted, frames)
	r := &endGame{}
	f := &Front{Game: &classicGame{Game: NewGame(r, game.Data{})}, started: true}
	f.Glyph, f.Font = func(byte) [8]byte { return [8]byte{} }, func(byte, bool) [8]byte { return [8]byte{} }
	f.P.SetEngine(f.Game)
	dst := gpu.NewImage(640, 400)
	f.Act()
	f.Draw(dst)
	f.Tick()
	f.Game.Answer(front.Action{Kind: front.ActQuick})
	f.P.Play(f.Game.Settle())
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
	if fight < 80 || page < 80 || page > 160 {
		t.Fatalf("%d frames of the fight while its blows play, the page from frame %d (want about 90 of each)", fight, page)
	}
}
