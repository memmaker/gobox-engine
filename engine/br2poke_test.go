package engine

import "testing"

// TestBR2DebugPoke: area 1's variable editor (`Word4B00[&NewGameBytes_2] = NewGameBytes_4`, eclvm.pas EclSaveTable: WriteGbvm of base +
// offset) writes the zone 0 byte at that address: a named field where there is one (4B00 + 289 = 4C21, Word4C21) and, in
// area 1's layout, the bytes no script names (Zone0Gap: 4B00 + 404 = 4C94, and the last typed offset 4B00 + 999 = 4EE7). Block 1 is the test's own here, so its layout is area 1's.
func TestBR2DebugPoke(t *testing.T) {
	e := brBegun(t, gameBR2)
	br1OwnBlock(t, e, 1, "    Word4B00[289] = 7\n    Word4B00[404] = 5\n    Word4B00[999] = 6\n    Word4BE7 = Zone0Gap4C94")
	if got := br1Cell(t, e, "Word4C21"); got != 7 {
		t.Errorf("Word4C21 %d, want 7: the poke at 4B00+289", got)
	}
	if got := br1Cell(t, e, "Word4BE7"); got != 5 {
		t.Errorf("the gap byte 4EE7 reads %d, want 5", got)
	}
	r, _ := e.layoutField("Zone0Gap4D82")
	if got := e.host.Get(r.Plus(0x4EE7 - 0x4D82)); got != 6 {
		t.Errorf("the last typed offset's byte 4EE7 reads %d, want 6", got)
	}
}
