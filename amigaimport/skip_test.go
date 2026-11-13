package amigaimport

import (
	"fmt"
	"strings"
	"testing"

	"gobox/adf"
)

// TestLogSkipped: a disk file that is neither converted, copied nor given a
// reason is logged as UNKNOWN; the known ones are logged with their reason.
func TestLogSkipped(t *testing.T) {
	var got []string
	old := Logf
	Logf = func(f string, a ...any) { got = append(got, fmt.Sprintf(f, a...)) }
	defer func() { Logf = old }()
	disk := []adf.AdfFile{{Path: "D/PIC.TLB"}, {Path: "D/ALWAYS.TLB"}, {Path: "D/NEW.TLB"}, {Path: "D/SOUNDS.GLB"}, {Path: "c/echo"}, {Path: "D/X.info"}}
	logSkipped(disk, Spec{Gfx: map[string]Gfx{"pic": {}}, Files: []string{"SOUNDS.GLB"}})
	if len(got) != 2 || !strings.Contains(got[0], "ALWAYS.TLB: the Amiga's own UI") || !strings.Contains(got[1], "NEW.TLB: UNKNOWN") {
		t.Fatalf("logged %q", got)
	}
}
