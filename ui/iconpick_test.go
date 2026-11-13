package ui

import (
	"testing"
)

// TestIconPickStable: of two names as long in an item's name, the icon is
// always the later one's (a Staff Sling is a sling), never the map's order.
func TestIconPickStable(t *testing.T) {
	inv := &inventory{icons: map[string]int{"staff": 1, "sling": 2}}
	for range 200 {
		if n := inv.icon("Staff Sling", nil); n != 2 {
			t.Fatalf("Staff Sling's icon %d, want the sling's 2", n)
		}
	}
}
