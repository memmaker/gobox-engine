package engine

import (
	"gobox/rules1e"
	"testing"
)

// TestIconSlots: the Engine's icon slots are the rules' combatants.
func TestIconSlots(t *testing.T) {
	if iconSlots != rules1e.MaxCombatants {
		t.Fatalf("iconSlots %d, rules1e.MaxCombatants %d", iconSlots, rules1e.MaxCombatants)
	}
}
