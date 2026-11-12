package pool

import "testing"

// TestDispatchTablesFull (BUG-CLASSES 25): every slot the original fills
// has a port handler, and the tables are the original's size.
func TestDispatchTablesFull(t *testing.T) {
	e := testEngine(t)
	e.Init()
	// the affect table: rules1e TestAffectHandlersFull.
	// ECL ops: ovr04:2F2D's cases are 0x00..0x3D except 0x1F (it has none;
	// the port's pager, Journal, is there: gobox ecl V1). Spell (0x3B) is
	// back by ruling S11 (POR gains the handler; area 14's one use stays the
	// native CastKnock, and TestLayoutCoversScripts refuses Spell in any script).
	for i, op := range e.vm.Ops[:0x3E] {
		if op == nil {
			t.Errorf("ECL op %#x: no handler", i)
		}
	}
}
