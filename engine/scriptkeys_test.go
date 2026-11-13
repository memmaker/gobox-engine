package engine

import (
	"errors"
	"testing"

	"gobox/ecl"
)

// TestScriptsCurrent: every imported game's area scripts are keyed as the
// engine keys them (pod's panicked at start: its tables lacked the import's
// ScriptFixes, and its entries file was written after the texts were keyed).
func TestScriptsCurrent(t *testing.T) {
	for _, g := range []string{"por", "cotab", "sotsb", "pod"} {
		e, err := NewPool("../data/"+g, "../data/"+g+"/sources")
		if err != nil {
			t.Logf("%s: %v", g, err)
			continue
		}
		for area := range 0x100 {
			if _, err := e.parseScriptFile(area); errors.Is(err, ecl.ErrNotImported) {
				t.Errorf("%s: %v", g, err)
				break
			}
		}
	}
}
