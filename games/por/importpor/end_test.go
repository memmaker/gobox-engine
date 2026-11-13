package importpor

import "testing"

// TestEndPartyAndText: every step of POR's finale says its text, and the party pictures
// show only with as many members as they stand for.
func TestEndPartyAndText(t *testing.T) {
	steps := end().Steps
	if len(steps) != 3 {
		t.Fatalf("%d steps", len(steps))
	}
	for i, st := range steps {
		if len(st.Text) == 0 {
			t.Errorf("step %d says nothing", i)
		}
	}
	for _, l := range steps[1].Layers {
		want := map[int]int{3: 0, 4: 2, 5: 3, 6: 4}
		if m, ok := want[l.Block]; ok && l.MinParty != m {
			t.Errorf("FINAL %d: MinParty %d, want %d", l.Block, l.MinParty, m)
		}
		if len(l.Frames) > 0 && !l.MsgDelayed {
			t.Errorf("animation %d ignores the message delay", l.Block)
		}
	}
}
