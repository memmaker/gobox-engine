package pool

import (
	"testing"

	"gobox/core"
)

// TestDevFlags: -freetrain turns the debug flag on at the party menu (the hall
// trains free), -fight's monster is the combat scene's, -at enters a block at x, y.
func TestDevFlags(t *testing.T) {
	e := quickStarted(t)
	e.FreeTraining = true
	e.push(cont{kind: kNop})
	e.partyMenu(true, true)
	if e.worldWord(core.WordDebug) == 0 {
		t.Error("-freetrain: the debug flag is off at the party menu")
	}
	e = quickStarted(t)
	e.push(cont{kind: kNop})
	e.DevAt(1, 3, 4) // the area's start events may move the party on
	if e.gs.areaKind != 1 || e.run.areaChanged {
		t.Errorf("-at: block %d, area change pending %v", e.gs.areaKind, e.run.areaChanged)
	}
}
