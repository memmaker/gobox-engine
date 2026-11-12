package pool

import (
	"gobox/ecl/script"
	"os"
	"testing"
)

// TestVendorGatesFolded: no checked-in script keeps a gate the pass cuts, and a
// folded vendor's exit goes to the cut No branch.
func TestVendorGatesFolded(t *testing.T) {
	e := quickStarted(t)
	for area := range 40 {
		s, err := e.parseScriptFile(area)
		if err != nil {
			continue
		}
		names, err := script.ReadNames(os.DirFS(e.scriptDir()), area, false, script.V1)
		if err != nil {
			t.Fatal(err)
		}
		if f := script.FoldVendorGates(s, area, names); len(f) != 0 {
			t.Errorf("area %d: unfolded %+v", area, f)
		}
	}
	for _, c := range []struct {
		area   int
		vendor string
		no     string
	}{{0, "TempleHealing", "MustLeave"}, {11, "OpenTraining", "ClearAndExit"}} {
		s, err := e.parseScriptFile(c.area)
		if err != nil {
			t.Fatal(err)
		}
		i := s.Labels[c.vendor]
		for i < len(s.Code) && s.Code[i].Op != script.OpGoto && s.Code[i].Op != script.OpExit {
			i++
		}
		if in := s.Code[i]; in.Op != script.OpGoto || in.Args[0].N != s.Labels[c.no] {
			t.Errorf("area %d: %s leaves by %+v, not %s", c.area, c.vendor, in, c.no)
		}
	}
}
