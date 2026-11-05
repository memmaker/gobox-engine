package front

import "testing"

// Status is in the treasure menu only when the game asks, before Exit.
func TestGateStatus(t *testing.T) {
	_, items := Gate(nil, &Loot{}, nil)
	if len(items) != 2 {
		t.Fatalf("%d items", len(items))
	}
	_, items = Gate(nil, &Loot{Status: true}, nil)
	if len(items) != 3 || items[1].text != "Status" || items[2].text != "Exit" {
		t.Fatalf("%+v", items)
	}
}
