package engine

import (
	"testing"
)

// TestInteractionLabels: every label interactions.txt names is a label of its area's script.
func TestInteractionLabels(t *testing.T) {
	e := testEngine(t)
	e.Init()
	dir := e.scriptDir()
	for area, labels := range interactions(dir) {
		e.LoadArea(area)
		s := e.vm.Script
		for _, l := range labels {
			if _, ok := s.Labels[l]; !ok {
				t.Errorf("area %d: no label %s", area, l)
			}
		}
	}
}
