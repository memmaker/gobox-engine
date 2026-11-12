package pool

import (
	"reflect"
	"testing"

	"gobox/core"
	"gobox/game"
)

// TestQuickPartyData: Data().QuickParty is the old quickParty table.
func TestQuickPartyData(t *testing.T) {
	e := testEngine(t)
	old := []struct {
		name       string
		head, body int
		need       map[string]int
	}{
		{"Thorgrim", 3, 2, map[string]int{"str": 18, "con": 17}},
		{"Mirelle", 6, 4, map[string]int{"wis": 17, "con": 15, "str": 15}},
		{"Aldric", 9, 7, map[string]int{"str": 17, "wis": 16, "con": 15}},
		{"Sylvaine", 12, 9, map[string]int{"str": 16, "int": 17, "dex": 16}},
		{"Corwin", 1, 11, map[string]int{"int": 18, "dex": 16, "con": 14}},
		{"Pip", 14, 5, map[string]int{"str": 16, "dex": 17, "con": 15}},
	}
	var got []game.QuickPick = Data().QuickParty
	if len(got) != len(old) {
		t.Fatalf("%d picks, want %d", len(got), len(old))
	}
	for i, o := range old {
		g := got[i]
		if g.Name != o.name || !e.isPort("HEAD", g.Head, o.head) || !e.isPort("BODY", g.Body, o.body) || !reflect.DeepEqual(g.Need, o.need) {
			t.Errorf("pick %d = %+v, want %+v", i, g, o)
		}
	}
}

// isPort: the template's picture id is POR's own portrait number want.
func (e *Engine) isPort(family, id string, want int) bool {
	b, ok := e.portByte(family, core.PicID(id))
	return ok && int(b) == want
}
