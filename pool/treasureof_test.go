package pool

import (
	"gobox/ecl/script"
	"rules1e"
	"testing"
)

// opArgs runs one instruction with the operands n on the engine's VM.
func opArgs(e *Engine, op func(), n ...int) {
	args := make([]script.Arg, len(n))
	for i, v := range n {
		args[i] = script.Arg{N: v}
	}
	e.vm.Script = &script.Script{Code: []script.Instr{{Args: args}}, Layout: e.vm.Script.Layout}
	e.vm.PC = 0
	e.vm.Fetch()
	op()
}

// libraryEngine is quickStarted with a two-entry item library (types 5 and 6, names hidden) and treasure sets.
func libraryEngine(t *testing.T) *Engine {
	e := quickStarted(t)
	tb := rules1e.POR
	tb.Library = []*rules1e.Item{{Type: 5, Hidden: 7}, {Type: 6, Hidden: 7}}
	e.state().Tables = &tb
	g := e.game()
	g.TreasureSets = [][]int32{nil, {1}, {1, 2}}
	e.given = &g
	return e
}

func treasureTypes(e *Engine) []int {
	var out []int
	for _, it := range e.state().Treasure {
		out = append(out, int(it.Type))
	}
	return out
}

// TestTreasureOf: the coins add to the loot (the pre-set ones stay), the listed library items land at the head in turn.
func TestTreasureOf(t *testing.T) {
	e := libraryEngine(t)
	opArgs(e, e.opTreasureOf, 2, 0, 0, 2, 1, 2)
	opArgs(e, e.opTreasureOf, 3, 0, 0, 0)
	want := rules1e.CoinsToPurse([7]uint16{4: 5})
	if e.state().Loot != want {
		t.Errorf("loot %+v, want %+v", e.state().Loot, want)
	}
	if got := treasureTypes(e); len(got) != 2 || got[0] != 6 || got[1] != 5 {
		t.Errorf("treasure types %v, want [6 5]", got)
	}
}

// TestItemSet: the set replaces the treasure with its items, names known; set 0 and one past the table are set 1.
func TestItemSet(t *testing.T) {
	e := libraryEngine(t)
	for _, tc := range []struct {
		set  int
		want int
	}{{2, 2}, {0, 1}, {9, 1}} {
		opArgs(e, e.opItemSet, tc.set)
		if got := e.state().Treasure; len(got) != tc.want || got[0].Hidden != 0 {
			t.Errorf("set %d: %d items, first hidden %d", tc.set, len(got), got[0].Hidden)
		}
	}
}
