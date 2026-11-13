package importer

import (
	"testing"
)

// TestStartCreditsPicture: exactly one title step, TITLE 3, is flagged as the
// credits picture, so gobox adds its remaster credit line to it.
func TestPodStartCreditsPicture(t *testing.T) {
	n := 0
	for _, s := range StartSequence().Steps {
		if s.Credits {
			n++
			if s.Kind != StepPicture || s.Layers[0].Block != 3 {
				t.Errorf("credits step %+v", s)
			}
		}
	}
	if n != 1 {
		t.Errorf("%d credits pictures", n)
	}
}
