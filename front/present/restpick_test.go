package present

import (
	"testing"

	"gobox/front"
)

// The rest picker: Hours, Add five times, Mins Up Up Down, Rest: 0d 5h 5m, 305 minutes.
func TestRestPick(t *testing.T) {
	var p Presenter
	p.askRest(2)
	q := p.Asking()
	for _, n := range []int{1, 3, 3, 3, 3, 3} {
		if q.Answer(front.Action{Kind: front.ActChoose, N: n}) {
			t.Fatalf("closed on %d", n)
		}
	}
	q.Answer(front.Action{Kind: front.ActChoose, N: 2}) // Mins: Up adds 5, Down takes it back
	q.Answer(front.Action{Kind: front.ActPage, N: -1})
	q.Answer(front.Action{Kind: front.ActPage, N: -1})
	q.Answer(front.Action{Kind: front.ActPage, N: 1})
	if !q.Answer(front.Action{Kind: front.ActChoose, N: 5}) {
		t.Fatal("Rest did not close it")
	}
	if a := p.inv.acts; len(a) != 1 || a[0] != (front.Action{Kind: front.ActRest, N: 2, X: 305}) {
		t.Fatalf("%+v", a)
	}
}
