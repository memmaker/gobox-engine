package games

import (
	"testing"

	"gobox/importer"
)

// TestStartCreditsPicture: the Amiga credits card (TITLE 5) is the one
// flagged credits picture, so gobox adds its remaster credit line to it.
func TestSotsbStartCreditsPicture(t *testing.T) {
	n := 0
	for _, s := range sotsbStart().Steps {
		if s.Credits {
			n++
			if s.Kind != importer.StepPicture || s.Layers[0].Block != titleAmigaCredits || s.Edition != importer.EditionAmiga {
				t.Errorf("credits step %+v", s)
			}
		}
	}
	if n != 1 {
		t.Errorf("%d credits pictures", n)
	}
}
