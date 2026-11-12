package pool

import (
	"rules1e"
	"testing"
)

// TestSellValue: ovr20:1CEE pays half the item's Price (+0x3A, gold), a
// stack Count/20 of that; RandomItem's weapons cost Plus*2000 (a hidden +2
// Javelin: 4000, sold for 2000 gold, shown as "Javelin").
func TestSellValue(t *testing.T) {
	jav := &rules1e.Item{Type: rules1e.ItemJavelin, Plus: 2, Price: 2 * 2000}
	if v := jav.SellValue(); v/200 != 2000 || v%200 != 0 {
		t.Errorf("+2 Javelin sells for %d copper, want 2000 gold", v)
	}
	arrows := &rules1e.Item{Type: rules1e.ItemArrow, Plus: 1, Price: 150, Count: 20}
	if v := arrows.SellValue(); v != 150*100 {
		t.Errorf("20 arrows sell for %d copper, want %d", v, 150*100)
	}
}
