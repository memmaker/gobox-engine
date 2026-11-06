package present

import "testing"

// TestFoePanelOnIcons: the foe panel's backdrop (labels and grid) starts on
// the first foe icon's cell, so the icons sit beside the labels, not over
// their first letters, and each label is on its figure's row (12 on).
func TestFoePanelOnIcons(t *testing.T) {
	if sfCells[0] != sfCells[23] {
		t.Fatalf("foe panel at %v, its first icon at %v", sfCells[0], sfCells[23])
	}
}
