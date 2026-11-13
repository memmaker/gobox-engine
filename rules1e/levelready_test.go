package rules1e

import (
	"testing"
)

// TestLevelReady: the roster's purple name (c70cbee engine/train.go
// levelReady) needs the experience for the next level in an active class,
// not a dormant one.
func TestLevelReady(t *testing.T) {
	c := &Character{}
	c.Race = RaceHuman
	c.Stats[StatStr] = 16
	c.Levels[ClsFighter] = 1
	need := uint32(XPForLevel(CoreClassFighter, 2))
	c.Exp = need - 1
	if levelReady(c) {
		t.Fatal("ready below the experience")
	}
	c.Exp = need
	if !levelReady(c) {
		t.Fatal("not ready at the experience")
	}
	c.OldLevels[ClsFighter] = 1
	if levelReady(c) {
		t.Fatal("a dormant class is ready")
	}
}
