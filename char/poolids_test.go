package char

import (
	"testing"

	"rules1e"
)

// TestSplitPoolIDs: PoolTrailer, COTAB's old GOBXID, POD's old bare art
// trailer and SotSB's old .ids file read as the same ids.
func TestSplitPoolIDs(t *testing.T) {
	want := rules1e.PoolIDs{PortraitHeadID: "h", IconBodyID: "ib", Roster: 7}
	old := []byte{1, 0, 'h', 0, 0, 0, 0, 2, 0, 'i', 'b', 1, 0, '7'}
	for name, c := range map[string]struct {
		b []byte
		n int
	}{
		"PoolTrailer": {append(make([]byte, Layout422), PoolTrailer(want)...), int(Layout422)},
		"GOBXID":      {append(append(make([]byte, Layout422), "GOBXID\x05"...), old...), int(Layout422)},
		"POD art":     {append(make([]byte, Layout510), old...), int(Layout510)},
	} {
		rec, got := SplitPoolIDs(c.b, c.n)
		if got != want || len(rec) != c.n {
			t.Errorf("%s: %+v (%d bytes), want %+v", name, got, len(rec), want)
		}
	}
	if got := ReadIDsFile([]byte("h\n\n\nib\n7\n")); got != want {
		t.Errorf(".ids: %+v", got)
	}
}
