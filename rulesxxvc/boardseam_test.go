package rulesxxvc

import (
	"image"
	"testing"

	"gobox/rules"
)

// The seam: a Fight with no board runs on the ruleset's own (r.ground); its units follow the Effects.
func TestBoardSeam(t *testing.T) {
	r := New()
	r.ground = rules.Board{Size: image.Pt(4, 3), Units: []rules.BoardUnit{{ID: 1, At: rules.BoardPos{X: 1, Y: 1}}, {ID: 2}}}
	b := r.board(rules.Board{Straight: true})
	if b.Size != r.ground.Size || len(b.Units) != 2 || len(b.Terrain) != 12 || b.Terrain[0] != openTerrain {
		t.Fatalf("own board %+v", b)
	}
	given := rules.Board{Size: image.Pt(2, 2)}
	if r.board(given).Size != given.Size {
		t.Fatal("a handed board must stay")
	}
	r.keepOnGround([]rules.Effect{
		rules.Walked{Unit: 1, Path: []rules.BoardPos{{X: 2, Y: 1}, {X: 3, Y: 1}}},
		rules.Turned{Unit: 1, Facing: 4},
		rules.Left{Unit: 2},
	})
	if len(r.ground.Units) != 1 || r.ground.Units[0].At != (rules.BoardPos{X: 3, Y: 1}) || r.ground.Units[0].Facing != 4 {
		t.Errorf("units %+v", r.ground.Units)
	}
}
