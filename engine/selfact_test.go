package engine

import (
	"gobox/rules1e"
	"testing"
)

// TestSelRaceAlignmentOutOfTable: a record byte outside the race or alignment
// table reads as the byte itself (SpecialReadZ1 $D8: ShortInt(alignment)),
// never a panic.
func TestSelRaceAlignmentOutOfTable(t *testing.T) {
	e := &Engine{}
	c := &rules1e.Character{}
	c.Race, c.Alignment = 9, 200
	e.state().Party = []*rules1e.Character{c}
	f := e.newEngineFields()
	for _, c := range []struct {
		name string
		want int64
	}{{"SelRace", 9}, {"SelAlignment", int64(signedByte(200))}} {
		if got, _ := f[c.name].get(0); got != c.want {
			t.Errorf("%s = %d, want %d", c.name, got, c.want)
		}
	}
}

// TestSelMemorized: a script's spell number reaches the slot through the
// typed Memo, the pending flag (0x80) and an out-of-table number unchanged.
func TestSelMemorized(t *testing.T) {
	e := &Engine{}
	c := &rules1e.Character{}
	e.state().Party = []*rules1e.Character{c}
	f := e.newEngineFields()["SelMemorized"]
	for i, v := range []int64{15, 0x80 | 3, 0, 0x7F} {
		f.set(i, v)
		if got := int64(c.Memorized[i+1].Byte()); got != v {
			t.Errorf("slot %d = %#x, want %#x", i+1, got, v)
		}
	}
}

// TestSelStatus: 0x87 stones and deactivates, other numbers >= 0x80 only
// deactivate; the read is 1 active, 0x80 not.
func TestSelStatus(t *testing.T) {
	for _, c := range []struct {
		v      int64
		status byte
	}{{0x80, rules1e.StatusOkay}, {0x87, rules1e.StatusStoned}} {
		e := &Engine{}
		m := &rules1e.Character{}
		m.Okay = 1
		e.state().Party = []*rules1e.Character{m}
		f := e.newEngineFields()["SelStatus"]
		if got, _ := f.get(0); got != 1 {
			t.Errorf("active read %#x", got)
		}
		f.set(0, c.v)
		if got, _ := f.get(0); got != 0x80 || m.Status != c.status {
			t.Errorf("%#x: read %#x status %v, want 0x80 %v", c.v, got, m.Status, c.status)
		}
	}
}

// TestSelAlignmentWrite: a script's write sets the member's alignment.
func TestSelAlignmentWrite(t *testing.T) {
	e := &Engine{}
	c := &rules1e.Character{}
	e.state().Party = []*rules1e.Character{c}
	f := e.newEngineFields()["SelAlignment"]
	f.set(0, 5)
	if c.Alignment != 5 {
		t.Errorf("alignment = %d, want 5", c.Alignment)
	}
}
