package pool

import (
	"rules1e"
	"testing"
)

// TestDispatchTablesFull (BUG-CLASSES 25): every slot the original fills
// has a port handler, and the tables are the original's size.
func TestDispatchTablesFull(t *testing.T) {
	e := testEngine(t)
	e.Init()
	// ovr13:31A2 fills DS:6828 slots 1..0x8B (0x58, breath, is set by the spell setup).
	for i := 1; i < len(rules1e.AffectHandlers); i++ {
		if rules1e.AffectHandlers[i] == nil {
			t.Errorf("affect %#x has no handler", i)
		}
	}
	if len(rules1e.AffectHandlers) != 0x8C {
		t.Errorf("affect table has %d slots, the original 0x8C", len(rules1e.AffectHandlers))
	}
	// ECL ops: ovr04:2F2D's cases are 0x00..0x3D except 0x1F (it has none;
	// the port's pager, Journal, is there: gobox ecl V1). Spell (0x3B) is
	// gone by Felix d3: area 14's one use is the native CastKnock, and
	// TestLayoutCoversScripts refuses Spell in any script.
	const spell = 0x3B
	for i, op := range e.vm.Ops[:0x3E] {
		if want := i != spell; (op != nil) != want {
			t.Errorf("ECL op %#x: handler %v, want %v", i, op != nil, want)
		}
	}
}
