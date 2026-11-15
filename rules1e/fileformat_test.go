package rules1e

import (
	"bytes"
	"testing"

	"gobox/rules"
)

// TestFileFormatBytes: the FileFormat seam writes the bytes the functions it
// wraps do.
func TestFileFormatBytes(t *testing.T) {
	var ff rules.FileFormat = &Rules{}
	rec := append(Encode(Record285{}), PoolTrailer(PoolIDs{Roster: 3, IconHeadID: "h"})...)
	files := Files{Record: rec}
	c, err := ff.Pack(files)
	want, err2 := Pack285(files)
	if err != nil || err2 != nil || !bytes.Equal(c, want) {
		t.Fatalf("Pack differs: %v %v", err, err2)
	}
	got, err := ff.Unpack(c, true)
	if err != nil || !bytes.Equal(got.Record, rec) {
		t.Fatalf("Unpack record differs: %v", err)
	}
	if ff.RosterOf(rec) != 3 {
		t.Fatal("RosterOf")
	}
	ids := ReadPoolIDs(rec)
	ids.Roster = 9
	if w := ff.WithRoster(rec, 9); !bytes.Equal(w, append(rec[:Layout285:Layout285], PoolTrailer(ids)...)) || ff.RosterOf(w) != 9 {
		t.Fatal("WithRoster")
	}
}
