package pool

import "testing"

func TestMakeFileName(t *testing.T) {
	if got := MakeFileName("Sir Bob.the Great"); got != "SIRBOBTH" {
		t.Errorf("MakeFileName = %q, want SIRBOBTH", got)
	}
}
