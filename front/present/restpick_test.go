package present

import (
	"testing"

	"gobox/front"
)

// The rest picker: Hours, Add five times, Rest gives 0d 5h 0m, 300 minutes.
func TestRestPick(t *testing.T) {
	var p Presenter
	p.askRest(2)
	q := p.Asking()
	for _, n := range []int{1, 3, 3, 3, 3, 3} {
		if q.Answer(front.Action{Kind: front.ActChoose, N: n}) {
			t.Fatalf("closed on %d", n)
		}
	}
	if !q.Answer(front.Action{Kind: front.ActChoose, N: 5}) {
		t.Fatal("Rest did not close it")
	}
	if a := p.inv.acts; len(a) != 1 || a[0] != (front.Action{Kind: front.ActRest, N: 2, X: 300}) {
		t.Fatalf("%+v", a)
	}
}
