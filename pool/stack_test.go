package pool

import (
	"rules1e"
	"testing"
)

func TestAddItemCopyJoinsStacks(t *testing.T) {
	e := &Engine{}
	e.initVM()
	c := &rules1e.Character{}
	arrows := rules1e.Item{Type: 1, Count: 200, Weight: 1}
	sword := rules1e.Item{Type: 2}
	e.state().AddItemCopy(c, &arrows)
	e.state().AddItemCopy(c, &sword)
	e.state().AddItemCopy(c, &sword)
	if len(c.Items) != 3 {
		t.Fatalf("non-stackables must not join: %d items", len(c.Items))
	}
	arrows.Count = 30
	e.state().AddItemCopy(c, &arrows)
	if len(c.Items) != 3 || c.Items[0].Count != 230 {
		t.Fatalf("arrows should join the first stack: %d items, %d", len(c.Items), c.Items[0].Count)
	}
	arrows.Count = 50
	e.state().AddItemCopy(c, &arrows)
	if len(c.Items) != 4 || c.Items[0].Count != 255 || c.Items[3].Count != 25 {
		t.Fatalf("overflow past 255 should leave a second stack: %d items", len(c.Items))
	}
}
