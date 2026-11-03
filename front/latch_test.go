package front

import "testing"

// The input that opens a screen never acts on it: held through the change
// it is ignored until let go, then works again.
func TestLatch(t *testing.T) {
	var l Latch
	l.See("", nil)
	if !l.Down("left", true) {
		t.Fatal("a press on the same screen should count")
	}
	l.See("inventory", []any{"left"}) // the click opened the inventory
	if l.Down("left", true) {
		t.Fatal("the click that opened the screen acted on it")
	}
	l.See("inventory", nil) // let go
	if !l.Down("left", true) {
		t.Fatal("a new press on the new screen should count")
	}
	l.See("inventory", []any{"left"}) // the same screen: nothing turns stale
	if !l.Down("left", true) {
		t.Fatal("a held input on an unchanged screen was dropped")
	}
}

// Calls hands the queued calls over once, in order.
func TestCalls(t *testing.T) {
	var c Calls[*[]int]
	c.Add(func(s *[]int) { *s = append(*s, 1) })
	c.Add(func(s *[]int) { *s = append(*s, 2) })
	var got []int
	for _, f := range c.Take() {
		f(&got)
	}
	if len(got) != 2 || got[0] != 1 || got[1] != 2 || c.Take() != nil {
		t.Fatal(got)
	}
}
