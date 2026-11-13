package ui

import (
	"testing"

	"gobox/game/play"
)

// The rest picker: Hours, Add five times, Mins Up Up Down, Rest: 0d 5h 5m, 305 minutes.
func TestRestPick(t *testing.T) {
	var p Presenter
	p.askRest(2)
	q := p.Asking()
	for _, n := range []int{1, 3, 3, 3, 3, 3} {
		if q.Answer(Action{Kind: ActChoose, N: n}) {
			t.Fatalf("closed on %d", n)
		}
	}
	q.Answer(Action{Kind: ActChoose, N: 2}) // Mins: Up adds 5, Down takes it back
	q.Answer(Action{Kind: ActPage, N: -1})
	q.Answer(Action{Kind: ActPage, N: -1})
	q.Answer(Action{Kind: ActPage, N: 1})
	if !q.Answer(Action{Kind: ActChoose, N: 5}) {
		t.Fatal("Rest did not close it")
	}
	if a := p.inv.acts; len(a) != 1 || a[0] != (Action{Kind: ActRest, N: 2, X: 305}) {
		t.Fatalf("%+v", a)
	}
}

// The watch: a fight breaks into a rest. While a member's turn waits, the
// rest is not paced (its step would answer the turn) and the turn's keys
// are not "Stop resting?".
func TestRestNotPacedOverTurn(t *testing.T) {
	var p Presenter
	p.board = &play.Combat{Turn: &play.Turn{Unit: 1}}
	p.Pace()
	if len(p.Calls()) != 0 {
		t.Fatal("the rest stepped over the member's turn")
	}
}
