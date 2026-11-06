package combat

import "testing"

// A unit killed in the tick it is hit keeps its place for its number; a new
// fight's same id does not float at the old fight's place.
func TestTopsKeepFallen(t *testing.T) {
	a, b := vec{1, 2, 3}, vec{4, 2, 6}
	tops := nextTops(nil, nil, []head{{Unit{ID: 1}, a}, {Unit{ID: 2}, b}})
	tops = nextTops(tops, []int{1, 2}, []head{{Unit{ID: 1}, a}}) // 2 fell this tick
	if tops[2] != b {
		t.Fatalf("the fallen unit's place lost: %v", tops)
	}
	tops = nextTops(tops, []int{1}, []head{{Unit{ID: 7}, a}}) // a new fight
	if _, ok := tops[2]; ok {
		t.Fatal("an old fight's place kept")
	}
}

// A Fighting state with Turn 0 (no unit active) keeps the eye where it
// was, else on the units' middle, never the map's corner.
func TestRestFocus(t *testing.T) {
	b := &Board{Units: []Unit{{ID: 1, X: 10, Y: 10}, {ID: 2, X: 12, Y: 10}}}
	if f := restFocus(vec{}, b); f != (vec{11.5, 0, 10.5}) {
		t.Fatalf("middle %v", f)
	}
	if f := restFocus(vec{3, 0, 4}, b); f != (vec{3, 0, 4}) {
		t.Fatalf("last %v", f)
	}
}
